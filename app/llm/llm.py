from flask import Flask, request, jsonify
import openai
import os
from dotenv import load_dotenv
import urllib.parse
import hashlib

app = Flask(__name__)

# Loading information structures
load_dotenv("../../.env/file")
OPEN_AI_API_KEY = os.environ.get("OPEN_AI_API_KEY")
openai.api_key = OPEN_AI_API_KEY
PASS_KEY = os.environ.get("PASS_KEY")
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
        response = openai.Completion.create(
            engine="text-davinci-001",
            prompt=decoded_prompt,
            max_tokens=225  # Adjust as needed
        )

        generated_text = response.choices[0].text.strip()
        print(generated_text)

        return jsonify({'response': generated_text})

    except openai.error.OpenAIError as e:
        print(f"OpenAI error: {e}")
        return jsonify({'error': str(e)}), 500
    except Exception as e:
        print(f"Error: {e}")
        return jsonify({'error': 'Failed to generate response'}), 500

if __name__ == '__main__':
    app.run(debug=True)
