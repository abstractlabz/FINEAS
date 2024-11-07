from datetime import datetime
import os
from flask import Flask, request, jsonify, make_response, redirect
from pymongo.mongo_client import MongoClient
from pymongo.server_api import ServerApi
import stripe
from flask_cors import CORS

app = Flask(__name__)
CORS(app)
MONGO_PASS = str(os.getenv("MONGO_DB_LOGGER_PASSWORD")) or ""
REDIRECT_DOMAIN = str(os.getenv("REDIRECT_DOMAIN")) or ""
uri = "mongodb+srv://kobenaidun:"+MONGO_PASS+"@cluster0.z9znpv9.mongodb.net/?retryWrites=true&w=majority"
# Create a new client and connect to the server
client = MongoClient(uri, server_api=ServerApi('1'))
db = client['User']  # Your MongoDB database
userlist = db['UserInformation']  # Your MongoDB collection name

stripe.api_key = os.getenv("STRIPE_SECRET_KEY")  # Set your Stripe API key

@app.route('/upgrade_membership', methods=['POST'])
def upgrade_membership():
    data = request.json
    id_hash = data.get('id_hash')
    email = data.get('email', None)

    if id_hash is None:
        return make_response(jsonify({'error': 'You must be logged in'}), 400)


    # Retrieve or create a user document
    user = userlist.find_one({'id_hash': id_hash})
    if not user:
        user = {
            'id_hash': id_hash,
            'credits': 25,
            'ismember': False
        }
        user['_id'] = userlist.insert_one(user).inserted_id
        user['_id'] = str(user['_id'])

    # Check if user already has a Stripe customer ID
    if 'stripe_customer_id' not in user:
        # Create a new Stripe customer
        customer = stripe.Customer.create()
        stripe_customer_id = customer.id

        # Log the created Stripe customer ID
        print(f"Created Stripe customer ID: {stripe_customer_id}")

        # Save the Stripe customer ID in the user's MongoDB document
        update_result = userlist.update_one({'id_hash': id_hash}, {'$set': {'stripe_customer_id': stripe_customer_id}})

        # Check if the update was successful
        if update_result.modified_count == 1:
            print(f"Successfully saved stripe_customer_id for user with id_hash {id_hash}")
        else:
            print(f"Failed to save stripe_customer_id for user with id_hash {id_hash}. Update result: {update_result.raw_result}")

        # Verify the update by fetching the document again (optional, for debugging)
        user = userlist.find_one({'id_hash': id_hash})
        print(f"User document after update: {user}")

    else:
        stripe_customer_id = user['stripe_customer_id']

    # Before proceeding with creating a checkout session, check if stripe_customer_id is valid
    if not stripe_customer_id or stripe_customer_id.strip() == "":
        print(f"Invalid stripe_customer_id for user with id_hash {id_hash}: '{stripe_customer_id}'")
        return make_response(jsonify({'error': 'Failed to create or retrieve a valid Stripe customer ID'}), 500)


    # Proceed with creating a checkout session
    # Now, include the customer ID in the session creation
    try:
        price_id = "price_1P9N2JCajx6ndCSaSMUbXE57"  # Replace with your price ID

        checkout_session = stripe.checkout.Session.create(
            customer=stripe_customer_id,
            line_items=[
                {
                    'price': price_id,
                    'quantity': 1,

                },
            ],
            metadata={
                'id_hash': id_hash,
            },
            mode='subscription',
            success_url=REDIRECT_DOMAIN +
            '?success=true&session_id={CHECKOUT_SESSION_ID}',
            cancel_url=REDIRECT_DOMAIN + '?canceled=true',
        )
        # Prepare user information for JSON response (excluding sensitive fields like _id)
        user_info = {
            "id_hash": user['id_hash'],
            "credits": user.get('credits', 0),
            "ismember": user.get('ismember', False),
            "stripe_customer_id": stripe_customer_id
        }

        return jsonify({'checkout_session_id': checkout_session.id, 'user': user_info, 'url' : checkout_session.url}, 200)

    except Exception as e:
        print(e)
        return "Server error", 500

@app.route('/session-status', methods=['GET'])
def session_status():
  session = stripe.checkout.Session.retrieve(request.args.get('session_id'))
  return jsonify(status=session.status, customer_email=session.customer_details.email, metadata=session.metadata)

@app.route('/get-user-info', methods=['GET'])
def get_user_info():
    id_hash = request.args.get('id_hash', None)
    email = request.args.get('email', None)
    
    if not id_hash:
        return make_response(jsonify({'error': 'ID hash is required as a query parameter'}), 400)
    
    # Retrieve user document based on id_hash
    user = userlist.find_one({'id_hash': id_hash}, {'_id': 0})  # Exclude the MongoDB-generated ID from the response
    
    if user:
        if email:
            if 'email' not in user:
                user['email'] = email
                user['signup_date'] = datetime.now()
                userlist.update_one({'id_hash': id_hash}, {'$set': {'email': email, 'signup_date': datetime.now()}})

    if not user:
        # insert a new user document if not found
        user = {
            'id_hash': id_hash,
            'credits': 25,
            'ismember': False
        }
        if email:
            user['email'] = email
            user['signup_date'] = datetime.now()
        userlist.insert_one(user)
        
        return jsonify({'user': user}), 200
    
    return jsonify({'user': user}), 200

@app.route('/cancel-subscription', methods=['POST'])
def cancel_subscription():
    data = request.json
    stripe_customer_id = data.get('stripe_customer_id')

    if not stripe_customer_id:
        return make_response(jsonify({'error': 'Stripe customer ID is required'}), 400)

    try:
        # List all active subscriptions for the customer
        subscriptions = stripe.Subscription.list(customer=stripe_customer_id, status='active')
        
        # Cancel all active subscriptions
        for sub in subscriptions.auto_paging_iter():
            stripe.Subscription.delete(sub.id)

        # After successfully canceling subscriptions, update the user in MongoDB
        update_result = userlist.update_one(
            {'stripe_customer_id': stripe_customer_id},
            {'$set': {'ismember': False, 'credits': 3}}
        )

        # Check if the user document was found and updated
        if update_result.matched_count == 0:
            # No user found with that Stripe customer ID
            return jsonify({'warning': 'User not found in database, but subscriptions were canceled'}), 200
        elif update_result.modified_count == 0:
            # Document found but not modified - could indicate the values were already set
            return jsonify({'info': 'User found, but document was not modified'}), 200

        return jsonify({'success': True, 'message': 'All subscriptions canceled and user updated in database'}), 200
    except Exception as e:
        print(e)
        return jsonify({'error': 'Failed to cancel subscription'}), 500
    

@app.route('/enforce-credits', methods=['POST'])
def enforce_credits():
    data = request.json
    id_hash = data.get('id_hash')
    
    if not id_hash:
        return make_response(jsonify({'error': 'ID hash is required'}), 400)
    
    user = userlist.find_one({'id_hash': id_hash})

    if user:
        # User found, check credits
        if user['credits'] > 0:
            # Decrement credits
            new_credits = user['credits'] - 1
            userlist.update_one({'id_hash': id_hash}, {'$set': {'credits': new_credits}})
            user['credits'] = new_credits  # Update local copy for response
        elif not user.get('ismember', False):
            # No credits left and not a member
            return make_response(jsonify({'error': 'Ran out of credits'}), 402)  # 402 Payment Required
    else:
        # User not found, insert a new document with default values
        user = {
            'id_hash': id_hash,
            'credits': 25,
            'ismember': False,
            'stripe_customer_id': ''
        }
        userlist.insert_one(user)
    
    # Return the document, but without MongoDB's internal ID
    user.pop('_id', None)
    return jsonify(user)

@app.route('/savechat', methods=['POST'])
def save_chat():
    data = request.json
    chat_name = data.get('chatname')
    id_hash = data.get('id_hash')
    chat_history = data.get('chat_history')

    if not chat_name or not id_hash or not chat_history:
        return jsonify({"error": "Chat name, ID hash, and chat history are required"}), 400

    # Assuming a separate collection for chats
    chat_history_collection = db['chats']
    chat_name_collection = db['chatnames']

    # Check for existing chat with the same name for the user
    existing_chat = chat_history_collection.find_one({"chatname": chat_name, "id_hash": id_hash})
    if existing_chat:
        chat_history_collection.update_one({"chatname": chat_name, "id_hash": id_hash}, {"$set": {"chat_history": chat_history}})
        chat_names = chat_name_collection.find({"id_hash": id_hash}, {'_id': 0})
        return jsonify({"message": "Chat updated successfully", "chat_names": [chat['chatname'] for chat in chat_names]})

    # Save the new chat history
    chat_document = {
        "chatname": chat_name,
        "id_hash": id_hash,
        "chat_history": chat_history
    }
    chat_history_collection.insert_one(chat_document)
    chat_name_collection.insert_one({"chatname": chat_name, "id_hash": id_hash})
    chat_names = chat_name_collection.find({"id_hash": id_hash}, {'_id': 0})
    return jsonify({"message": "Chat saved successfully", "chat_names": [chat['chatname'] for chat in chat_names]})

@app.route('/loadchat', methods=['POST'])
def load_chat():
    data = request.json
    chat_name = data.get('chatname')
    id_hash = data.get('id_hash')

    # Use the same collection for chats as in savechat
    chats_collection = db['chats']

    chat = chats_collection.find_one({"chatname": chat_name, "id_hash": id_hash}, {'_id': 0})
    if chat:
        return jsonify(chat)
    else:
        return jsonify({"error": "Chat not found"}), 404
    

@app.route('/delete-chats', methods=['GET'])
def delete_chat():
    chat_name = request.args.get('chatname')
    id_hash = request.args.get('id_hash')

    if not chat_name or not id_hash:
        return jsonify({"error": "Chat name and ID hash are required"}), 400

    # Use the same collection for chats as in savechat
    chats_collection = db['chats']
    chat_name_collection = db['chatnames']

    # Delete the chat history
    chats_collection.delete_one({"chatname": chat_name, "id_hash": id_hash})
    chat_name_collection.delete_one({"chatname": chat_name, "id_hash": id_hash})

    chat_names = chat_name_collection.find({"id_hash": id_hash}, {'_id': 0})
    return jsonify({"message": "Chat deleted successfully", "chat_names": [chat['chatname'] for chat in chat_names]})

@app.route('/get-chat-names', methods=['GET'])
def get_chat_names():
    id_hash = request.args.get('id_hash')

    if not id_hash:
        return jsonify({"error": "ID hash is required"}), 400

    chat_name_collection = db['chatnames']
    chat_names = chat_name_collection.find({"id_hash": id_hash}, {'_id': 0})
    return jsonify([chat['chatname'] for chat in chat_names])

    
if __name__ == '__main__':
    app.run()
