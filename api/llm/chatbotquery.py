from flask import Flask, request, jsonify
import os
import urllib.parse
import hashlib
import requests
from flask_cors import CORS
from langchain_openai import OpenAIEmbeddings
from pinecone import Pinecone, Config, PodSpec
app = Flask(__name__)
CORS(app)

# Environment Variables
PASS_KEY = os.getenv("PASS_KEY")
OPEN_AI_API_KEY = os.getenv("OPEN_AI_API_KEY")
MODEL = 'text-embedding-ada-002'
PINECONE_API_KEY = os.getenv("PINECONE_API_KEY")
LLM_SERVICE_URL = os.getenv("LLM_SERVICE_URL")

# Initialize Pinecone client
pinecone_config = Config(api_key=PINECONE_API_KEY,host='pre-alpha-vectorstore-prd-d284fae.svc.gcp-starter.pinecone.io')
print(pinecone_config)
pinecone_client = Pinecone(config=pinecone_config)

# Create/Open Pinecone index
index = pinecone_client.Index(host='pre-alpha-vectorstore-prd-d284fae.svc.gcp-starter.pinecone.io', name='pre-alpha-vectorstore-prd')

# Initialize OpenAI Embeddings
embed = OpenAIEmbeddings(api_key=OPEN_AI_API_KEY)

@app.route('/chat', methods=['POST'])
def chatbot():
    raw_data = request.args.get('prompt')
    
    passhash = request.headers.get('Authorization')[7:]
    # Security measures
    sha256_hash = hashlib.sha256()
    sha256_hash.update(PASS_KEY.encode('utf-8'))
    HASH_KEY = sha256_hash.hexdigest()
    if passhash != HASH_KEY:
        return jsonify({'error': 'Unauthorized access'}), 401
    
    if not raw_data:
        return jsonify({'error': 'No data given'}), 400

    raw_data = urllib.parse.unquote(raw_data)  # Decode URL-encoded prompt
        
    query_vector = embed.embed_query(raw_data)
    context = index.query(vector=query_vector, top_k=2)

    prompt_payload = f"You are an AI named FINEAS tasked to generate investment advice about the following company for the user. You will analyze, make inferences and value judgements off of financial data from the following prompt:\n\nPROMPT:\n{raw_data}\n\nThe following is the only data context from which you will answer this prompt:\n\nCONTEXT:\n{str(context)}"
    
    url = LLM_SERVICE_URL + "/llm"

    headers = {'Authorization': 'Bearer ' + HASH_KEY}
    params = {'prompt': prompt_payload}

    chatresponse = requests.get(url, headers=headers, params=params)
    return chatresponse.text

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=6002, debug=False, ssl_context=('../../utils/keys/query.fineasapp.io.cer', '../../utils/keys/query.fineasapp.io.key'))
