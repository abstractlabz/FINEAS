from flask import Flask, request, jsonify
import os
import hashlib
import requests
from flask_cors import CORS
from tenacity import retry, wait_fixed, stop_after_attempt
from langchain_openai import OpenAIEmbeddings
from pinecone import Pinecone, Config
import logging

app = Flask(__name__)
CORS(app)

# Environment Variables
PASS_KEY = os.getenv("PASS_KEY")
CLAUDE_API_KEY = os.getenv("CLAUDE_API_KEY")
MODEL = 'multilingual-e5-large'
PINECONE_API_KEY = os.getenv("PINECONE_API_KEY")
LLM_SERVICE_URL = os.getenv("LLM_SERVICE_URL")

# Initialize Pinecone client
host = "https://main-uajrq2f.svc.aped-4627-b74a.pinecone.io"
pinecone_config = Config(api_key=PINECONE_API_KEY, host=host)
pinecone_client = Pinecone(config=pinecone_config)

# Create/Open Pinecone index
index = pinecone_client.Index(host=host, name='pre-alpha-vectorstore-prd')

# Initialize OpenAI Embeddings
embed = OpenAIEmbeddings(api_key=CLAUDE_API_KEY)

# Configure Logging
logging.basicConfig(level=logging.INFO)

# Retry logic for external service calls
@retry(wait=wait_fixed(2), stop=stop_after_attempt(3))
def fetch_chat_response(url, headers, payload):
    try:
        response = requests.post(url, headers=headers, json=payload)
        response.raise_for_status()  # This will raise HTTPError for bad responses
        return response
    except requests.exceptions.HTTPError as http_err:
        logging.error(f"HTTP error occurred: {http_err}")  # Log the HTTP error details
        raise
    except Exception as err:
        logging.error(f"An error occurred: {err}")  # Log any other exceptions
        raise

@app.route('/chat', methods=['POST'])
def chatbot():
    try:
        data = request.get_json()
        raw_data = data.get('prompt')  # Get prompt from JSON payload
        if not raw_data:
            return jsonify({'error': 'No data given'}), 400

        passhash = request.headers.get('Authorization')[7:]
        sha256_hash = hashlib.sha256()
        sha256_hash.update(PASS_KEY.encode('utf-8'))
        HASH_KEY = sha256_hash.hexdigest()
        if passhash != HASH_KEY:
            return jsonify({'error': 'Unauthorized access'}), 401

        # Embedding and context retrieval
        query_vector = embed.embed_query(raw_data)
        context = index.query(vector=query_vector, top_k=7, include_metadata=True)
        context = [match['metadata']['text'] for match in context['matches']]

        #Get search information from google search
        search_information = get_search_query(raw_data, HASH_KEY)
        print(search_information)

        # Construct the prompt payload
        prompt_payload = {
            "prompt": f"""You are an AI assistant named Fineas tasked with giving stock market alpha to retail investors
              by summarizing and analyzing financial information in the form of market research. When displaying numbers, show two decimal places.
              Your response will answer the following prompt using structured informative headers, short paragraph segments, annotations, and bullet points for the given financial data. 
              If relevant to the prompt, include general company information such as background history, founder history, current leadership, product history, business segments and their revenue contributions, and anything else pertinent like M&A transactions.
              You will also attach annotation information only defined within the annotations section of this prompt throughout response segments in the text.
              \n\nPROMPT:\n{raw_data}\n\n
              The following is the only data context and annotations data from which you will answer this prompt. Only use the annotations which are relevant to the prompt. Ignore the irrelevant annotations and don't include them in your response nor make any reference to them.
              only based off of the most relevant information with 250 words maximum.:
              \n\nANNOTATIONS:\n{search_information}
              \n\nCONTEXT:\n{str(context)}"""
        }

        # Prepare the POST request to LLM service
        url = LLM_SERVICE_URL + "/llm"
        headers = {'Authorization': 'Bearer ' + HASH_KEY, 'Content-Type': 'application/json'}

        # Send POST request with the prompt payload in the body
        chatresponse = fetch_chat_response(url, headers, prompt_payload)
        return chatresponse.text, 200

    except requests.RequestException as e:
        logging.error(f"Request error: {str(e)}")
        return jsonify({'error': 'Failed to fetch response from LLM service'}), 500
    except Exception as e:
        logging.error(f"Internal error: {str(e)}")
        return jsonify({'error': 'Internal Server Error'}), 500

def get_search_query(raw_data, passhash):
    # Prepare the data and headers for the POST request
    data = {"query": raw_data}
    headers = {
        "Content-Type": "application/x-www-form-urlencoded",
        "Authorization": f"Bearer {passhash}"
    }

    # Make a POST request to the search service
    response = requests.post("http://localhost:8070/search", headers=headers, data=data)
    return response.text

if __name__ == '__main__':
    app.run()
