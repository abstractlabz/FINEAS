import json
import stripe
from flask import Flask, request, jsonify, abort
from pymongo.server_api import ServerApi
from pymongo.mongo_client import MongoClient
import os
from flask_cors import CORS

app = Flask(__name__)
CORS(app)
MONGO_PASS = str(os.getenv("MONGO_DB_LOGGER_PASSWORD")) or ""
STRIPE_ENDPOINT_SECRET = str(os.getenv("STRIPE_ENDPOINT_SECRET")) or ""
uri = "mongodb+srv://kobenaidun:"+MONGO_PASS+"@cluster0.z9znpv9.mongodb.net/?retryWrites=true&w=majority"
client = MongoClient(uri, server_api=ServerApi('1'))
db = client['User']  # Your MongoDB database
userlist = db['UserInformation']  # Your MongoDB collection name

# Set your Stripe secret key: remember to switch to your live secret key in production
# See your keys here: https://dashboard.stripe.com/apikeys
stripe.api_key = os.getenv('STRIPE_SECRET_KEY')  # Or directly assign your Stripe secret key
# Stripe webhook endpoint
@app.route('/stripe_webhook', methods=['POST'])
def stripe_webhook():
    payload = request.get_data(as_text=True)  # Get the raw payload
    sig_header = request.headers.get('Stripe-Signature')
    print('Payload: {}'.format(payload))
    print('Signature: {}'.format(sig_header))

    if not STRIPE_ENDPOINT_SECRET:
        print('Stripe endpoint secret is not set.')
        return jsonify(success=False), 400

    try:
        event = stripe.Webhook.construct_event(
            payload, sig_header, STRIPE_ENDPOINT_SECRET
        )
    except ValueError:
        # Invalid payload
        print('Error while decoding event!')
        return jsonify(success=False), 400
    except stripe.error.SignatureVerificationError:
        # Signature verification failed
        print('Invalid signature!')
        return jsonify(success=False), 400

    # Handle the event
    if event and event['type'] == 'checkout.session.completed':
        payment_intent = event['data']['object']  # contains a stripe.PaymentIntent
        print('Payment for {} succeeded'.format(payment_intent['metadata']))
        id_hash = payment_intent['metadata']['id_hash']
        # Then define and call a method to handle the successful payment intent.
        user = userlist.find_one({'id_hash': id_hash})
        if not user:
            # Insert a new user document if not found
            user = {
                'id_hash': id_hash,
                'credits': 5,  # Assuming starting credits
                'ismember': False
            }
            user['_id'] = userlist.insert_one(user).inserted_id
            user['_id'] = str(user['_id'])  # Convert ObjectId to string
         #Update the user's membership status
        userlist.update_one({'id_hash': id_hash}, {'$set': {'ismember': True}})
        #give the user 1000000 credits
        userlist.update_one({'id_hash': id_hash}, {'$set': {'credits': 1000000}})
    else:
        # Unexpected event type
        print('Unhandled event type {}'.format(event['type']))

    return jsonify(success=True)
    
if __name__ == '__main__':
    app.run(host='0.0.0.0', port=7000,debug=True, ssl_context=('../../utils/keys/webhook/fullchain.pem', '../../utils/keys/webhook/privkey.pem'))
