import json
import requests
from pinecone.grpc import PineconeGRPC as Pinecone
from pinecone import ServerlessSpec
from tqdm.auto import tqdm
from uuid import uuid4
from langchain.text_splitter import RecursiveCharacterTextSplitter
from flask import Flask, request, jsonify
from pinecone import Pinecone, Config
import os
import urllib.parse
import hashlib
from datasets import Dataset
import datetime
from flask_cors import CORS

app = Flask(__name__)
CORS(app)

# Environment variables
PASS_KEY = os.getenv("PASS_KEY", "default_pass_key")
MODEL = "multilingual-e5-large"
PINECONE_API_KEY = "e914eca9-0bcb-4380-94af-c83de762252b"

# Initialize Pinecone client
host = "https://main-uajrq2f.svc.aped-4627-b74a.pinecone.io"
pinecone_config = Config(api_key=PINECONE_API_KEY, host=host)
pinecone_client = Pinecone(config=pinecone_config)
index = pinecone_client.Index(host=host, name="main")


@app.route("/ingestor", methods=["POST"])
def ingest_data():
    try:
        # Security check
        raw_data = request.form.get("info")
        passhash = (request.headers.get("Authorization", ""))[7:]

        sha256_hash = hashlib.sha256()
        sha256_hash.update(PASS_KEY.encode("utf-8"))
        HASH_KEY = sha256_hash.hexdigest()

        if not raw_data:
            return jsonify({"error": "No data given"}), 400

        raw_data = urllib.parse.unquote(raw_data)  # URL-decode data
        data = json.loads(raw_data)  # Parse JSON input

        # Prepare data for Pinecone ingestion
        current_date = datetime.date.today().strftime("%Y-%m-%d")
        #turn the current date into a integer
        current_date = int(current_date.replace("-", ""))
        list_dict = [
            {"id": i, "info": key, "current_date": current_date, "text": value}
            for i, (key, value) in enumerate(data.items())
        ]

        dataset = Dataset.from_list(list_dict)

        # Text splitting and embeddings
        batch_limit = 1
        text_splitter = RecursiveCharacterTextSplitter(chunk_size=500, chunk_overlap=0)

        for record in tqdm(dataset):
            texts = text_splitter.split_text(record["text"])
            metadatas = [{"id": record["id"], "info": record["info"], "current_date": record["current_date"]} for _ in texts]

            ids = [str(uuid4()) for _ in texts]
            embeddings = embed_query(texts)

            # Upsert into Pinecone
            index.upsert(vectors=zip(ids, embeddings, metadatas))

        return jsonify({"status": "Ingestion completed successfully"}), 200

    except Exception as e:
        print("Error:", e)
        return jsonify({"error": "Internal Server Error", "details": str(e)}), 500


def embed_query(texts):
    """Generates embeddings for a list of texts using Pinecone's python library."""
    pc = Pinecone(api_key=PINECONE_API_KEY)
    embedding_model = "multilingual-e5-large"
    embeddings = pc.inference.embed(
        model=embedding_model,
        inputs=texts,
        parameters={"input_type": "query", "truncate": "END"},
    )
    print([embeddings[0]['values']])
    return [embeddings[0]['values']]




if __name__ == "__main__":
    app.run()
