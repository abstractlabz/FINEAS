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

@app.route('/ingestor', methods=['POST'])
def ingest_data():
    #get params
    print(request.form)
    raw_data = request.form.get("info")
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

    #connection to pinecone
    pinecone.init(
    api_key=PINECONE_API_KEY,
    environment="gcp-starter"
    )

    # check if 'openai' index already exists (only create index if not)
    if 'devenv' not in pinecone.list_indexes():
        pinecone.create_index('devenv', dimension=len(embeds[0]))
    # connect to index
    index = pinecone.Index('devenv')

    # load the first 1K rows of the TREC dataset
    trec = load_dataset('trec', split='train[:1000]')

    batch_size = 32  # process everything in batches of 32
    for i in tqdm(range(0, len(trec['text']), batch_size)):
        # set end position of batch
        i_end = min(i+batch_size, len(trec['text']))
        # get batch of lines and IDs
        lines_batch = trec['text'][i: i+batch_size]
        ids_batch = [str(n) for n in range(i, i_end)]
        # create embeddings
        res = openai.Embedding.create(input=lines_batch, engine=MODEL)
        embeds = [record['embedding'] for record in res['data']]
        # prep metadata and upsert batch
        meta = [{'text': line} for line in lines_batch]
        to_upsert = zip(ids_batch, embeds, meta)
        # upsert to Pinecone
        index.upsert(vectors=list(to_upsert))

    return "200 Status OK"

if __name__ == '__main__':
    app.run(host='127.0.0.1', port=6001, debug=True)