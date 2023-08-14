from flask import Flask, request, jsonify
from datasets import load_dataset
from tqdm.auto import tqdm  # this is our progress bar
import pinecone
import openai
import os
from dotenv import load_dotenv
import urllib.parse
import hashlib

app = Flask(__name__)

load_dotenv()
PASS_KEY = os.environ.get("PASS_KEY")
OPEN_AI_API_KEY = os.environ.get("OPEN_AI_API_KEY")
openai.api_key = OPEN_AI_API_KEY
MODEL = "text-embedding-ada-002"
PINECONE_API_KEY = os.environ.get("PINECONE_API_KEY")

@app.route('/chat', methods=['GET'])
def chatbot():
    #get params
    raw_data = request.args.get('prompt')
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
    pinecone.init(
    api_key=PINECONE_API_KEY,
    environment="gcp-starter"  # find next to API key in console
    )

    # connect to index
    index = pinecone.Index('devenv')

    xq = openai.Embedding.create(input=raw_data, engine=MODEL)['data'][0]['embedding']
    res = index.query([xq], top_k=5, include_metadata=True)

    res_list = []
    for match in res['matches']:
        res_list.append(match['metadata']['text'])
   

    return res_list[0]

if __name__ == '__main__':
    app.run(host='127.0.0.1', port=6002, debug=True)