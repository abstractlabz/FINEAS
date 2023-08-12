from flask import Flask, request, jsonify
import openai
import os
from dotenv import load_dotenv
import urllib.parse
import hashlib

app = Flask(__name__)

load_dotenv()
OPEN_AI_API_KEY = os.environ.get("OPEN_AI_API_KEY")
PASS_KEY = os.environ.get("PASS_KEY")
openai.api_key = OPEN_AI_API_KEY
MODEL = "text-embedding-ada-002"
@app.route('/ingestor', methods=['POST'])
def ingest_data():
    #get params
    prompt = request.args.get('prompt')
    passhash = request.args.get('passhash')

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

    res = openai.Embedding.create(
        input=[
            "Sample document text goes here",
            "there will be several phrases in each batch"
        ], engine=MODEL
    )
    embeds = [record['embedding'] for record in res['data']]
    print(embeds)

if __name__ == '__main__':
    app.run(debug=True)