from flask import Flask, request, jsonify
import os
import hashlib
from flask_cors import CORS
from openai import OpenAI

app = Flask(__name__)
CORS(app)

# Loading environment variables
client = OpenAI(
    api_key=os.getenv("OPENAI_API_KEY")
)
PASS_KEY = os.getenv("PASS_KEY")

@app.route('/llm', methods=['POST'])
def generate_response():
    try:
        # Get JSON payload from the POST request
        data = request.get_json()

        # Extract prompt and authorization from the request body
        prompt = data.get('prompt')
        passhash = request.headers.get('Authorization', '')[7:]

        # Security: Verify the hashed passkey
        sha256_hash = hashlib.sha256()
        sha256_hash.update(PASS_KEY.encode('utf-8'))
        HASH_KEY = sha256_hash.hexdigest()

        if passhash != HASH_KEY:
            return jsonify({'error': 'Unauthorized access'}), 401

        if not prompt:
            return jsonify({'error': 'Missing prompt parameter'}), 400

        # Process the prompt
        print(prompt)  # Optional: log the prompt for debugging purposes

        # Call the OpenAI API for chat completion
        try:
            response = client.chat.completions.create(
                messages=[
                    {
                        "role": "user",
                        "content": prompt
                    }
                ],
                model="gpt-4o",
                max_tokens=2048
            )

            reply = response.choices[0].message.content
            print(reply)  # Log the response for debugging purposes
            return reply, 200

        except Exception as e:
            raise Exception(f"OpenAI API Error: {str(e)}")

    except Exception as e:
        print(f"Error: {e}")
        return jsonify({'error': str(e)}), 500

if __name__ == '__main__':
    app.run()
