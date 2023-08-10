import openai
import os
from dotenv import load_dotenv

load_dotenv()
OPEN_AI_API_KEY = os.environ.get("OPEN_AI_API_KEY")
PASS_KEY = os.environ.get("PASS_KEY")
openai.api_key = OPEN_AI_API_KEY
MODEL = "text-embedding-ada-002"

res = openai.Embedding.create(
    input=[
        "Sample document text goes here",
        "there will be several phrases in each batch"
    ], engine=MODEL
)
embeds = [record['embedding'] for record in res['data']]
print(embeds)