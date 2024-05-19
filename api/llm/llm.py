from flask import Flask, request, jsonify
from openai import OpenAI
import os
import urllib.parse
import hashlib
from flask_cors import CORS
app = Flask(__name__)
CORS(app)

# Loading information structures
OPEN_AI_API_KEY = os.getenv("OPEN_AI_API_KEY")
PASS_KEY = os.getenv("PASS_KEY")

client = OpenAI(api_key=OPEN_AI_API_KEY)

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
        response = client.chat.completions.create(

            messages=[
            	{"role": "system", "content": """You are an AI agent tasked with summarizing and analyzing financial information for market research.
             	Your response will follow the task template given to you based on the financial data given to you. Give your summarized response. You will
             	respond to the following prompt in a structured bullet point based format. You will also include annotations to relevant sources from the web throughout the text, attached to the important bullet points. 
             	If the data containing the information is not relevant nor sufficient,
             	you may ask for more information in the response.
             	However, if the data containing the information is relevant to the prompt template, generate a market analysis report over the information
             	in accordance with the prompt template and categorize your analysis as either bullish, neutral, or bearish. Nothing more nothing less."""},
            	{"role": "user", "content": decoded_prompt}
        	],
            model="gpt-4o"  # Adjust model as needed
        )

        generated_text = response.choices[0].message.content
        print(generated_text)

        return generated_text

    except Exception as e:
        print(f"OpenAI error: {e}")
        return jsonify({'error': str(e)}), 500

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5432, debug=False)