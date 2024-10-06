
# Fineas.AI 🧠💼
## Your AI Powered Investment Researcher 📊🤖

Fineas' backend repo is an ETL system that feeds into a RAG knowledge base. Connected to financial data APIs, this infrastructure can perform financial analysis on over 600 financial asset tickers, including cryptocurrency support 💹💻.

LLM-powered investment researcher web-app (mobile client soon! 📱)

- 🤖 GPT-4o powered inferences
- 📊 Polygon.io financial data
- 📰 Google news data
- ✨ Proprietary web scrapers ✨

## Features 🚀

- 📅 Up-to-date analysis summaries across major categories
- 💬 Up-to-date financial chatbot
- 🤖 Discord bot integration (coming soon!)
- 📈 Portfolio optimization (coming soon!)

## Tech Stack 🛠️

Fineas uses a number of open-source projects to work properly:

- [⚙️ Golang] - For handling the ETL service
- [🔥 Flask] - For setting up the LLM and RAG services
- [🧠 Langchain] - Popular library to connect to OpenAI services
- [📚 Pinecone] - Vector database to store vectors for LLM knowledge base
- [💾 MongoDB] - Flexible No-SQL database for fast and simple I/O
- [🤖 GPT-4o] - Highest quality LLM that exists to date
- [📊 Polygon.io] - Reliable financial data API that supports streaming data via WSS
- [⚡ Akash] - Blockchain-based Docker hosting platform

## Setting Up Your Developer Environment 🖥️💻

Fineas requires [Python](https://www.python.org/) 3.9+, [Golang](https://go.dev/) 1.21.6+, and Docker to run locally.

1. Fork the repository on GitHub 🔧, then open up your text editor or terminal 💻.

2. Clone the repo:

\`\`\`bash
git clone https://github.com/abstractlabz/FINEAS.git
\`\`\`

3. Navigate to the \`FINEAS/scripts/startup\` directory:

\`\`\`bash
cd FINEAS/scripts/startup
\`\`\`

4. Open the \`startup_config_template.json\` file and create a new file named \`startup_config.json\`. Copy the contents from the template, and get the required keys from the repo admin 🔑.

\`\`\`json
{
    "API_KEY": "Ask the GitHub organization owner for env key secrets",
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
\`\`\`

5. Add the secret key files to your \`utils\` directory (contact the repo admin for files 📁).

6. Cd into the root directory and build from root...

### Windows 🖥️:

\`\`\`bash
docker build -t fineas-image:latest .
\`\`\`

### MacOS 🍎:

\`\`\`bash
docker build --platform linux/arm64 -t fineas-image:latest .
\`\`\`

## DNS Entries 🌐

Add these DNS entries to your hosts file for local development 🌍:

**For Windows:**

\`\`\`bash
C:/Windows/System32/Drivers/etc/hosts
\`\`\`

**For MacOS:**

\`\`\`bash
sudo nano /etc/hosts
\`\`\`

Add these entries:

\`\`\`bash
127.0.0.1 data.fineasapp.io
127.0.0.1 query.fineasapp.io
127.0.0.1 upgrade.fineasapp.io
127.0.0.1 webhook.fineasapp.io
\`\`\`

Flush your DNS cache after editing the hosts file:

\`\`\`bash
sudo dscacheutil -flushcache; sudo killall -HUP mDNSResponder
\`\`\`

**Note:** These entries are for development only. Comment them out for production.

## Running Locally 🏃‍♂️

Use Docker to run the Fineas image locally.

**Windows:**

\`\`\`bash
docker run -d -p 8443:8035 -p 443:6002 -p 2087:6001 -p 2083:7000 -p 2096:7002 -e API_KEY=[API_KEY] -e PASS_KEY=[PASS_KEY] -e MONGO_DB_LOGGER_PASSWORD=[MONGO_DB_LOGGER_PASSWORD] -e OPEN_AI_API_KEY=[OPEN_AI_API_KEY] -e KB_WRITE_KEY=[KB_WRITE_KEY] -e MR_WRITE_KEY=[MR_WRITE_KEY] -e PINECONE_API_KEY=[PINECONE_API_KEY] -e STRIPE_ENDPOINT_SECRET=[STRIPE_ENDPOINT_SECRET] -e STRIPE_SECRET_KEY=[STRIPE_SECRET_KEY] -e REDIRECT_DOMAIN=https://app.fineas.ai fineas-image:latest
\`\`\`

**MacOS:**

\`\`\`bash
docker run --platform linux/arm64 -d -p 8443:8035 -p 443:6002 -p 2087:6001 -p 2083:7000 -p 2096:7002 -e API_KEY=[API_KEY] -e PASS_KEY=[PASS_KEY] -e MONGO_DB_LOGGER_PASSWORD=[MONGO_DB_LOGGER_PASSWORD] -e OPEN_AI_API_KEY=[OPEN_AI_API_KEY] -e KB_WRITE_KEY=[KB_WRITE_KEY] -e MR_WRITE_KEY=[MR_WRITE_KEY] -e PINECONE_API_KEY=[PINECONE_API_KEY] -e STRIPE_ENDPOINT_SECRET=[STRIPE_ENDPOINT_SECRET] -e STRIPE_SECRET_KEY=[STRIPE_SECRET_KEY] -e REDIRECT_DOMAIN=https://app.fineas.ai fineas-image:latest
\`\`\`

## API Spec 📬

Interact with Fineas via HTTP requests or a hosted frontend 🎨.

| Description | Request Example |
| ------ | ------ |
| 🛠️ Collect new aggregated data for a ticker. | `curl "http://0.0.0.0:8080/?ticker=AMZN -H "Authorization: Bearer [Access Token]"` |
| 🤖 Process a prompt for financial info. | `curl -X POST "https://query.fineasapp.io:443/chat?prompt=What%20is%20some%20relevant%20news%20around%20amazon%3F" -H "Authorization: Bearer [Access Token]"` |

## License ⚖️

GPL 2.0 🆓

**Free Software, Hell Yeah!** 🎉
