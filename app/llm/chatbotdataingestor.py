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
    print(request.form)
    raw_data = request.form.get('data')
    passhash = (request.headers.get('Authorization'))[7:]
    #security measures
    sha256_hash = hashlib.sha256()
    sha256_hash.update(PASS_KEY.encode('utf-8'))
    HASH_KEY = sha256_hash.hexdigest()
    if passhash != HASH_KEY:
        return jsonify({'error': 'Unauthorized access'}), 401
    
    if not raw_data:
        print("Raw Data: ", raw_data)
        return jsonify({'error': 'No data given'}), 400
    
    raw_data = urllib.parse.unquote(raw_data)  # Decode URL-encoded prompt
    print(raw_data)

    res = openai.Embedding.create(
        input=[
            raw_data
        ], engine=MODEL
    )
    embeds = [record['embedding'] for record in res['data']]
    print(embeds)
    return "200 Status OK"

if __name__ == '__main__':
    app.run(host='127.0.0.1', port=6001, debug=True)