import json
import requests
from pinecone.grpc import PineconeGRPC as Pinecone
from pinecone import ServerlessSpec
from tqdm.auto import tqdm
from uuid import uuid4
from langchain.text_splitter import RecursiveCharacterTextSplitter
from flask import Flask, request, jsonify
from pinecone import Pinecone as PineconeClient, Config
import os
import urllib.parse
import hashlib
from datasets import Dataset
import datetime
from flask_cors import CORS

app = Flask(__name__)
CORS(app)

# Environment variables
PASS_KEY = os.getenv("PASS_KEY", "<YOUR_PASS_KEY_HERE>")
MODEL = "multilingual-e5-large"
PINECONE_API_KEY = os.getenv("PINECONE_API_KEY", "<YOUR_PINECONE_API_KEY>")
PINECONE_HOST = os.getenv("PINECONE_HOST", "<YOUR_PINECONE_HOST>")

# Initialize Pinecone client
host = PINECONE_HOST
pinecone_config = Config(api_key=PINECONE_API_KEY, host=host)
pinecone_client = PineconeClient(config=pinecone_config)
index = pinecone_client.Index(host=host, name="main")


@app.route("/ingestor", methods=["POST"])
def ingest_data():
    try:
        # Security check
        # We expect the aggregator to POST form-encoded data
        # under the key "info" (which is URL-escaped JSON).
        raw_data = request.form.get("info", "")
        passhash = (request.headers.get("Authorization", ""))[7:]

        if not raw_data:
            return jsonify({"error": "No data given"}), 400

        # Because aggregator is URL-escaping the JSON, we decode it:
        # e.g. {"CompanyDesc":"My Text ..."} => needs unquote first
        raw_data = urllib.parse.unquote(raw_data)  
        
        # Now parse the raw JSON
        data = json.loads(raw_data)  # Parse JSON input

        # Optionally validate passhash here if you want:
        # sha256_hash = hashlib.sha256()
        # sha256_hash.update(PASS_KEY.encode("utf-8"))
        # HASH_KEY = sha256_hash.hexdigest()
        # if passhash != HASH_KEY:
        #     return jsonify({"error": "Unauthorized"}), 403

        # Prepare data for Pinecone ingestion
        current_date = datetime.date.today().strftime("%Y-%m-%d")
        # Turn the current date into an integer, e.g. "20241220"
        current_date_int = int(current_date.replace("-", ""))

        # Build a list of dict for each key in the JSON we received
        list_dict = [
            {
                "id": i,
                "info": key, 
                "current_date": current_date_int, 
                "text": value
            }
            for i, (key, value) in enumerate(data.items())
        ]

        dataset = Dataset.from_list(list_dict)

        # Text splitting and embeddings
        batch_limit = 1
        text_splitter = RecursiveCharacterTextSplitter(
            chunk_size=500, 
            chunk_overlap=0
        )

        for record in tqdm(dataset):
            # Split the text by chunks
            texts = text_splitter.split_text(record["text"])
            # Create parallel metadata for each chunk
            metadatas = [
                {
                    "id": record["id"],
                    "info": record["info"],
                    "current_date": record["current_date"],
                    "text": record["text"]
                }
                for _ in texts
            ]
            # Create a unique ID for each chunk
            ids = [str(uuid4()) for _ in texts]

            # Embed all chunks properly (one embedding per chunk)
            embeddings = embed_query(texts)

            # Upsert into Pinecone in a single batch
            index.upsert(vectors=zip(ids, embeddings, metadatas))

        return jsonify({"status": " 200 Ingestion completed successfully"}), 200

    except Exception as e:
        print("Error:", e)
        return jsonify({"error": "Internal Server Error", "details": str(e)}), 500


def embed_query(texts):
    """Generates embeddings for a list of texts using Pinecone's python library."""
    pc = PineconeClient(api_key=PINECONE_API_KEY)
    embedding_model = "multilingual-e5-large"
    embeddings = pc.inference.embed(
        model=embedding_model,
        inputs=texts,
        parameters={"input_type": "query", "truncate": "END"},
    )
    # Return a list of vectors: one for each input text
    return [emb["values"] for emb in embeddings]


if __name__ == "__main__":
    app.run()
