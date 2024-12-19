from flask import Flask, request, jsonify
from openai import OpenAI
import os
import hashlib
from flask_cors import CORS
import requests

app = Flask(__name__)
CORS(app)

# Loading environment variables
CLAUDE_API_URL = "https://api.anthropic.com/v1/messages"
CLAUDE_API_KEY = os.getenv("CLAUDE_API_KEY")
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

        headers = {
            "x-api-key": CLAUDE_API_KEY,
            "Content-Type": "application/json",
            "anthropic-version": "2023-06-01",
        }

        payload = {
            "model": "claude-3-5-sonnet-20241022",  # Replace with the correct Claude model
            "messages": [{"role": "user", "content": prompt}],
            "max_tokens": 1024,
        }

        try:
            response = requests.post(CLAUDE_API_URL, headers=headers, json=payload)
            response.raise_for_status()
            print(response.json()["content"][0]["text"]) 
            return response.json()["content"][0]["text"], 200
        except Exception as e:
            raise Exception(f"Claude API Error: {str(e)}")

    except Exception as e:
        print(f"Claude error: {e}")
        return jsonify({'error': str(e)}), 500


if __name__ == '__main__':
    app.run()
