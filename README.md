# Fineas.AI
## Your AI Powered Investment Researcher

Fineas' backend repo is an ETL system which feeds into a RAG knowledge base. Connected to financial data apis this infrastructure is able to perform financial analysis on over 600 financial asset tickers including cryptocurrency support. 

LLM powered investment researcher web-app (mobile client soon!)

- GPT4o powered inferences
- Polygon.io financial data
- Google news data
- ✨Proprietary webscrapers✨

## Features

- Up to date analysis summaries across major categories
- Up to date financial chatbot
- Discord bot integration (coming soon!)
- Portfolio optimization (coming soon!)

## Tech Stack

Fineas uses a number of open source projects to work properly:

- [Golang] - For handling the ETL service
- [Flask] - For setting up the LLM and RAG services
- [Langchain] - Popular library to connect to OpenAI services
- [Pinecone] - Vector database to store vectors for LLM knowledge base
- [MongoDB] - Flexible No-SQL database for fast and simple I/O
- [GPT-4o] - Highest quality LLM that exists to date
- [Polygon.io] - Reliable financial data API that supports streaming data via WSS
- [Akash] - Blockchain based docker hosting platform

## Installation

Fineas requires [Python](https://www.python.org/) 3.9+, [Golang](https://go.dev/) 1.21.6+, and Docker to locally run.

Install the required languages in order to get started

Now fork and clone the repo

```sh
git clone https://github.com/abstractlabz/FINEAS.git
```

Cd into the root directory and build from root...

```sh
cd FINEAS
docker build -t fineas-image:latest .
```

Now go to your local hosts file and add these entries
**For Windows go to this directory and open the file:
```sh
C:/Windows/System32/Drivers/etc/hosts
```

Now add these DNS entries to the bottom:

```sh
127.0.0.1 data.fineasapp.io
127.0.0.1 query.fineasapp.io
127.0.0.1 upgrade.fineasapp.io
127.0.0.1 webhook.fineasapp.io
```

Run the docker image on your local machine 
**contact core team for dev level environment variables

```sh
docker run -d -p 8443:8035 -p 443:6002 -p 2087:6001 -p 2083:7000 -p 2096:7002 -e API_KEY=[API_KEY] -e PASS_KEY=[PASS_KEY] -e MONGO_DB_LOGGER_PASSWORD=[MONGO_DB_LOGGER_PASSWORD] -e OPEN_AI_API_KEY=[OPEN_AI_API_KEY] -e KB_WRITE_KEY=[KB_WRITE_KEY] -e MR_WRITE_KEY=[MR_WRITE_KEY] -e PINECONE_API_KEY=[PINECONE_API_KEY] -e STRIPE_ENDPOINT_SECRET=[STRIPE_ENDPOINT_SECRET] -e STRIPE_SECRET_KEY=[STRIPE_SECRET_KEY] -e REDIRECT_DOMAIN=https://fineas.ai fineas-image:latest
```

**Note: These entries must only be used in a development environment. Comment out these entries using '#' to access Fineas' production level domains  

## API Spec

Now you can locally interact with Fineas using http requests or any hosted front-end

| Description | Request Example |
| ------ | ------ |
| Collects aggregated data for a given ticker symbol.| curl "http://0.0.0.0:8080/?ticker=AMZN"
| Processes a prompt and returns relevant financial information. | curl -X POST "https://query.fineasapp.io:443/chat?prompt=What%20is%20some%20relevant%20news%20around%20amazon%3F"
| Retrieve the entire current cache of information for a ticker | curl -X GET "https://data.fineasapp.io:8443/ret?ticker=AAPL" -H "Authorization: Bearer [Access Token]" 
| Collect recent technical analysis data for a ticker. | curl -X GET "http://0.0.0.0:8089/ta?ticker=AAPL" -H "Authorization: Bearer [Access Token]"
| Collect recent description data for a ticker | curl -X GET "http://0.0.0.0:8084/desc?ticker=AAPL" -H "Authorization: Bearer [Access Token]" |
| Collect recent news data for a ticker | curl -X GET "http://0.0.0.0:8083/news?ticker=AAPL" -H "Authorization: Bearer [Access Token]" |
| Collect recent financial data for a ticker | curl -X GET "http://0.0.0.0:8082/fin?ticker=AAPL" -H "Authorization: Bearer [Access Token]" |
| Collect recent stock data for a ticker | curl -X GET "http://0.0.0.0:8081/stk?ticker=AAPL" -H "Authorization: Bearer [Access Token]" 


## License

GPL 2.0

**Free Software, Hell Yeah!**

[//]: # (These are reference links used in the body of this note and get stripped out when the markdown processor does its job. There is no need to format nicely because it shouldn't be seen. Thanks SO - http://stackoverflow.com/questions/4823468/store-comments-in-markdown-syntax)

   [Golang]: <https://go.dev/>
   [Flask]: <https://flask.palletsprojects.com/en/3.0.x/>
   [Langchain]: <https://www.langchain.com/>
   [Pinecone]: <https://www.pinecone.io/>
   [MongoDB]: <https://www.mongodb.com/products/platform/atlas-database>
   [GPT-4o]: <https://openai.com/index/hello-gpt-4o/>
   [Polygon.io]: <https://polygon.io/>
   [Akash]: <https://akash.network/>
