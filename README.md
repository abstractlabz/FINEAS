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

## Setting Up Your Developer Environment

Fineas requires [Python](https://www.python.org/) 3.9+, [Golang](https://go.dev/) 1.21.6+, and Docker to locally run.

Fork the repository on GitHub then open up your text editor or terminal in your desktop!

Now clone the repo

```sh
git clone https://github.com/abstractlabz/FINEAS.git
```

Next Cd into FINEAS/scripts/startup

```sh
cd FINEAS/scripts/startup
```

Open up the file startup_config_template.json it should look like this

```json
{
    "API_KEY": "Ask the github organization owner for env key secrets",
    "PASS_KEY": "",
    "MONGO_DB_LOGGER_PASSWORD": "",
    "OPEN_AI_API_KEY": "",
    "KB_WRITE_KEY": "",
    "MR_WRITE_KEY": "",
    "PINECONE_API_KEY": "",
    "STRIPE_ENDPOINT_SECRET": "",
    "STRIPE_SECRET_KEY": "",
    "REDIRECT_DOMAIN": "https://app.fineas.ai"
}

```

Create a new file in the directory named startup_config.json. Copy the contents of the template file there and ask the repo admin for the development keys. Then use those keys as values for the startup_config.json and save. 


Cd into the root directory and build from root...

```sh
cd ../..
docker build --platform linux/arm64 -t fineas-image:latest .
```

Now go to your local hosts file and add these entries
**For Windows go to this directory and open the file:

For Windows:
```sh
C:/Windows/System32/Drivers/etc/hosts
```

For MacOS:
```sh
/etc/hosts
```

Open the hosts file as administrator in your text editor then add these DNS entries to the bottom:

```sh
127.0.0.1 data.fineasapp.io
127.0.0.1 query.fineasapp.io
127.0.0.1 upgrade.fineasapp.io
127.0.0.1 webhook.fineasapp.io
```

**Note: These entries must only be used in a development environment. Comment out these entries using '#' to access Fineas' production level domains  

Ex.

```sh
#127.0.0.1 data.fineasapp.io
#127.0.0.1 query.fineasapp.io
#127.0.0.1 upgrade.fineasapp.io
#127.0.0.1 webhook.fineasapp.io
```

With the entries uncommented out, run the docker image on your local machine 
**contact repo admin for development environment variables

```sh
docker run --platform linux/arm64 -d -p 8443:8035 -p 443:6002 -p 2087:6001 -p 2083:7000 -p 2096:7002 -e API_KEY=[API_KEY] -e PASS_KEY=[PASS_KEY] -e MONGO_DB_LOGGER_PASSWORD=[MONGO_DB_LOGGER_PASSWORD] -e OPEN_AI_API_KEY=[OPEN_AI_API_KEY] -e KB_WRITE_KEY=[KB_WRITE_KEY] -e MR_WRITE_KEY=[MR_WRITE_KEY] -e PINECONE_API_KEY=[PINECONE_API_KEY] -e STRIPE_ENDPOINT_SECRET=[STRIPE_ENDPOINT_SECRET] -e STRIPE_SECRET_KEY=[STRIPE_SECRET_KEY] -e REDIRECT_DOMAIN=https://app.fineas.ai fineas-image:latest
```

**If you are getting port binding issues with this command, change the port number that is throwing the issue in the run command with one of these ports [9443, 10443, 5443, 6443, 60443]. Whatever ports you choose, consider them for the curl requests you make to the service running on it.

## API Spec

Now you can locally interact with Fineas using http requests or any hosted front-end

| Description | Request Example |
| ------ | ------ |
| collects new aggregated data for a given ticker symbol.| curl "http://0.0.0.0:8080/?ticker=AMZN -H "Authorization: Bearer [Access Token]"
| Processes a prompt and returns relevant financial information. | curl -X POST "https://query.fineasapp.io:443/chat?prompt=What%20is%20some%20relevant%20news%20around%20amazon%3F" -H "Authorization: Bearer [Access Token]"
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
