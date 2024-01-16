from flask import Flask, request, jsonify
import openai
import os
import urllib.parse
import hashlib

app = Flask(__name__)

# Loading information structures
OPEN_AI_API_KEY = os.getenv("OPEN_AI_API_KEY")
openai.api_key = OPEN_AI_API_KEY
PASS_KEY = os.getenv("PASS_KEY")

@app.route('/llm', methods=['GET'])
def generate_response():
    #get params
    prompt = request.args.get('prompt')
    passhash = (request.headers.get('Authorization'))[7:]
    print(passhash)
    #security measures
    sha256_hash = hashlib.sha256()
    sha256_hash.update(PASS_KEY.encode('utf-8'))
    HASH_KEY = sha256_hash.hexdigest()
    if passhash != HASH_KEY:
        return jsonify({'error': 'Unauthorized access'}), 401
    if not prompt:
        return jsonify({'error': 'Missing prompt parameter'}), 400

    decoded_prompt = urllib.parse.unquote(prompt)  # Decode URL-encoded prompt
    print(decoded_prompt)
    try:
        # Use the chat completion endpoint
        response = openai.ChatCompletion.create(
            model="gpt-4",  # Adjust model as needed
            messages=[
                {"role": "system", "content": """You are an AI purposed with summarizing and analyzing financial information for market research. 
                 Your response will follow the task template given to you based off of the financial data given to you. Give your summarized response.
                 If the data containing the information is not relevant nor sufficient, you may ask for more information in the response. Nothing more nothing less."""},
                {"role": "user", "content": decoded_prompt}
            ]
        )

        generated_text = response.choices[0].message['content'].strip()
        print(generated_text)

        return generated_text

    except openai.error.OpenAIError as e:
        print(f"OpenAI error: {e}")
        return jsonify({'error': str(e)}), 500
    except Exception as e:
        print(f"Error: {e}")
        return jsonify({'error': 'Failed to generate response'}), 500

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5432, debug=False)
