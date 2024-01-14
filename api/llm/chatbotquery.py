from flask import Flask, request, jsonify
from langchain.vectorstores import Pinecone
import pinecone
from langchain.embeddings.openai import OpenAIEmbeddings
import openai
import os
import urllib.parse
import hashlib
import requests

app = Flask(__name__)

PASS_KEY = os.getenv("PASS_KEY")
OPEN_AI_API_KEY = os.getenv("OPEN_AI_API_KEY")
openai.api_key = OPEN_AI_API_KEY
MODEL = 'text-embedding-ada-002'
PINECONE_API_KEY = os.getenv("PINECONE_API_KEY")
LLM_SERVICE_URL = os.getenv("LLM_SERVICE_URL")

@app.route('/chat', methods=['POST'])
def chatbot():
    #get params
    raw_data = request.args.get('prompt')
    
    passhash = (request.headers.get('Authorization'))[7:]
    #security measures
    sha256_hash = hashlib.sha256()
    sha256_hash.update(PASS_KEY.encode('utf-8'))
    HASH_KEY = sha256_hash.hexdigest()
    if passhash != HASH_KEY:
        print(passhash + "\n")
        print(HASH_KEY + "\n")
        return jsonify({'error': 'Unauthorized access'}), 401
    
    if not raw_data:
        print("Raw Data: ", raw_data)
        return jsonify({'error': 'No data given'}), 400

    raw_data = urllib.parse.unquote(raw_data)  # Decode URL-encoded prompt
        
    text_field = "text"
    index_name = "pre-alpha-vectorstore-prd"

    embed = OpenAIEmbeddings(
    document_model_name=MODEL,
    query_model_name=MODEL,
    openai_api_key=OPEN_AI_API_KEY
    )

    pinecone.init(
        api_key=PINECONE_API_KEY,  # find api key in console at app.pinecone.io
        environment="gcp-starter"  # find next to api key in console
    )
    index = pinecone.Index(index_name)

    vectorstore = Pinecone(
        index, embed.embed_query, text_field
    )

    #query the vector DB index for prompt
    context = vectorstore.similarity_search(raw_data,k=2)

    prompt_payload = "You are an AI named FINEAS tasked to generate investment advice about the following company for the user. You will analyze, make inferences and value judgements off of financial data from the following prompt:" + "\n" + "PROMPT:" + "\n" + raw_data + "\n" + "The following is the only data context from which you will answer this prompt:" + "\n" + "CONTEXT:" + "\n" + str(context)
    
    url = LLM_SERVICE_URL + "/llm"

    headers = {
    'Authorization': 'Bearer ' + HASH_KEY
    }

    params = {
    'prompt': prompt_payload
    }

    chatresponse = requests.get(url, headers=headers, params=params)
    return chatresponse.text

if __name__ == '__main__':
    app.run(host='127.0.0.1', port=6002, debug=True)