from flask import Flask, request, jsonify
import os
import urllib.parse
import hashlib
import requests
from flask_cors import CORS
from tenacity import retry, wait_fixed, stop_after_attempt
from langchain_openai import OpenAIEmbeddings
from pinecone import Pinecone, Config, PodSpec
import logging

app = Flask(__name__)
CORS(app)

# Environment Variables
PASS_KEY = os.getenv("PASS_KEY")
OPEN_AI_API_KEY = os.getenv("OPEN_AI_API_KEY")
MODEL = 'text-embedding-ada-002'
PINECONE_API_KEY = os.getenv("PINECONE_API_KEY")
LLM_SERVICE_URL = os.getenv("LLM_SERVICE_URL")

# Initialize Pinecone client
host = "https://pre-alpha-vectorstore-prd-uajrq2f.svc.apw5-4e34-81fa.pinecone.io"
pinecone_config = Config(api_key=PINECONE_API_KEY, host=host)
pinecone_client = Pinecone(config=pinecone_config)

# Create/Open Pinecone index
index = pinecone_client.Index(host=host, name='pre-alpha-vectorstore-prd')

# Initialize OpenAI Embeddings
embed = OpenAIEmbeddings(api_key=OPEN_AI_API_KEY)

# Configure Logging
logging.basicConfig(level=logging.INFO)

# Retry logic for external service calls
@retry(wait=wait_fixed(2), stop=stop_after_attempt(3))
def fetch_chat_response(url, headers, params):
    response = requests.get(url, headers=headers, params=params)
    response.raise_for_status()
    return response

@app.route('/chat', methods=['POST'])
def chatbot():
    try:
        raw_data = request.args.get('prompt')
        if not raw_data:
            return jsonify({'error': 'No data given'}), 400

        passhash = request.headers.get('Authorization')[7:]
        sha256_hash = hashlib.sha256()
        sha256_hash.update(PASS_KEY.encode('utf-8'))
        HASH_KEY = sha256_hash.hexdigest()
        if passhash != HASH_KEY:
            return jsonify({'error': 'Unauthorized access'}), 401

        raw_data = urllib.parse.unquote(raw_data)  # Decode URL-encoded prompt
        
        query_vector = embed.embed_query(raw_data)
        context = index.query(vector=query_vector, top_k=7, include_metadata=True)
        context = [match['metadata']['text'] for match in context['matches']]

        prompt_payload = f"""You are an AI assistant named Fineas tasked with giving stock market alpha to retail investors
             	by summarizing and analyzing financial information in the form of market research, when displaying numbers show two decimal places.
             	Your response will answer from the following prompt given to you in a structured bullet point format using informative headers, short paragraph segments, annotations, and bullet points based off of the given financial data. If relevant to the prompt, you should also include and offer general company information such as background history, founder history, current leadership, product history, business segments and their revenue contributions, and anything else pertinent like M&A transactions within the response. 
             	You will also attach annotations within response segments to relevant sources from the web throughout the text.
             	\n\nPROMPT:
             	\n{raw_data}\n\n
             	The following is the only data context from which you will answer this prompt. Please answer the prompt in bullet points
             	only based off of the most relevant information with 250 words maximum.:
             	\n\nCONTEXT:
            	\n{str(context)} \n
             	If the prompt given by the user is a greeting such as "Hello" or "Hi", you may respond with a greeting.
             	If the prompt given by the user is not relevant towards finance, you may respond to the prompt
             	as a default AI agent mode. However, if the prompt given by the user is finance related but the
             	context doesn't have relevant data to accurately respond to the prompt, then you may ask for more
             	information. Nothing more, nothing less."""

        url = LLM_SERVICE_URL + "/llm"
        headers = {'Authorization': 'Bearer ' + HASH_KEY}
        params = {'prompt': prompt_payload}

        chatresponse = fetch_chat_response(url, headers, params)
        return chatresponse.text

    except requests.RequestException as e:
        logging.error(f"Request error: {str(e)}")
        return jsonify({'error': 'Failed to fetch response from LLM service'}), 500
    except Exception as e:
        logging.error(f"Internal error: {str(e)}")
        return jsonify({'error': 'Internal Server Error'}), 500

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=6002, debug=False, ssl_context=('../../utils/keys/query.fineasapp.io.cer', '../../utils/keys/query.fineasapp.io.key'))
