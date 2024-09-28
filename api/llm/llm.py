from flask import Flask, request, jsonify
from openai import OpenAI
import os
import hashlib
from flask_cors import CORS

app = Flask(__name__)
CORS(app)

# Loading environment variables
OPEN_AI_API_KEY = os.getenv("OPEN_AI_API_KEY")
PASS_KEY = os.getenv("PASS_KEY")

# Initialize OpenAI client
client = OpenAI(api_key=OPEN_AI_API_KEY)

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
        response = client.chat.completions.create(
            messages=[
                {"role": "system", "content": """You are an AI agent tasked with summarizing and analyzing financial information for market research.
                Your response will follow the task template given to you based on the financial data given to you. Give your summarized response. You will
                respond to the following prompt in a structured bullet point based format. You will also include annotations to relevant sources from the web throughout the text, attached to the important bullet points. 
                If the data containing the information is not relevant nor sufficient,
                you may ask for more information in the response.
                However, if the data containing the information is relevant to the prompt template, generate a market analysis report over the information
                in accordance with the prompt template and categorize your analysis as either bullish, neutral, or bearish. Nothing more nothing less."""},
                {"role": "user", "content": prompt}
            ],
            model="gpt-4o"  # Adjust model as needed
        )

        # Extract generated content from response
        generated_text = response.choices[0].message.content
        print(generated_text)  # Optional: log the generated response

        # Return the response as plain text
        return generated_text, 200

    except Exception as e:
        print(f"OpenAI error: {e}")
        return jsonify({'error': str(e)}), 500


if __name__ == '__main__':
    app.run()
