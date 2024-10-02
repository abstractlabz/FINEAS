# Use a base image
FROM ubuntu:latest

# Install necessary tools and libraries
RUN apt-get update && apt-get install -y \
    curl \
    git \
    python3 \
    python3-pip \
    && rm -rf /var/lib/apt/lists/*

# Install Go
RUN curl -LO "https://golang.org/dl/go1.19.linux-amd64.tar.gz" \
    && tar -C /usr/local -xzf go1.19.linux-amd64.tar.gz \
    && rm go1.19.linux-amd64.tar.gz

# Set Go environment variables
ENV PATH="${PATH}:/usr/local/go/bin"
ENV GOPATH="/go"
ENV PATH="${PATH}:${GOPATH}/bin"

# Set the working directory
WORKDIR /app

# Copy the application files to the app directory
COPY . .

# Install Go dependencies
RUN go mod download

# Install Python dependencies
RUN pip3 install --no-cache-dir --break-system-packages -r utils/requirements.txt
# Set environment variables for keys
ENV API_KEY=API_KEY
ENV PASS_KEY=PASS_KEY
ENV MONGO_DB_LOGGER_PASSWORD=MONGO_DB_LOGGER_PASSWORD
ENV OPEN_AI_API_KEY=OPEN_AI_API_KEY
ENV KB_WRITE_KEY=KB_WRITE_KEY
ENV MR_WRITE_KEY=MR_WRITE_KEY
ENV PINECONE_API_KEY=PINECONE_API_KEY 
ENV STRIPE_ENDPOINT_SECRET = STRIPE_ENDPOINT_SECRET
ENV STRIPE_SECRET_KEY = STRIPE_SECRET_KEY
ENV REDIRECT_DOMAIN = REDIRECT_DOMAIN

# Set environment variables for templates
ENV STK_SERVICE_URL=http://0.0.0.0:8081
ENV FIN_SERVICE_URL=http://0.0.0.0:8082
ENV NEWS_SERVICE_URL=http://0.0.0.0:8083
ENV DESC_SERVICE_URL=http://0.0.0.0:8084
ENV LLM_SERVICE_URL=http://0.0.0.0:5432
ENV TA_SERVICE_URL=http://0.0.0.0:8089
ENV YTD_TEMPLATE="Based on the provided YTD stock data, give a brief and clear explanation of [Company Name]'s stock performance. Specify the percentage change in the stock price, notable trends, and external factors influencing performance. Conclude by categorizing the performance into one of these five categories based on predefined criteria (specify percentage ranges for each category): Bullish, Bearish, or Neutral. Ensure the summary is concise, reflects a solid financial understanding, and considers broader market influences. Cite real relevant web page links from the internet using reputable resources."
ENV NEWS_TEMPLATE="Provide a quick impact assessment (bullish, bearish, neutral) of the news headlines for [Company Name], then analyze their significance in detail (100-200 words). Prioritize based on impact to the company's situation and future. Summarize critical news, consider market trends, financial health, and industry developments. Highlight discrepancies in reports, concluding with potential long-term implications and the news' overall impact. Source information from Coindesk, CNBC, Bloomberg, Marketwatch, Investopedia, Yahoo Finance and other relevant sources"
ENV DESC_TEMPLATE="Provide a detailed yet concise pitch of [Company Name] suitable for potential investment analysis, covering key aspects in bullet points (150-300 words). Include Company Name, Year Founded, CEO, Number of Employees, Funding Rounds, Total Amount Raised, Market Capitalization, Recent Acquisitions, Significant Financial Metrics, Performance Trends, Valuation Insights, and Investment Thesis (Bullish, Bearish, or Neutral). Add sections for 'Key Challenges' and 'Recent News Impact' to offer a balanced view and link financial health with current events. Source information from Coindesk, CNBC, Bloomberg, Marketwatch, Investopedia, Yahoo Finance and other relevant sources"
ENV TA_TEMPLATE="Give an in-depth analysis (concluding with bullish, bearish or neutral) of the company stock ticker future price action based on recent stock information and following technical indicators: SMA & EMA, Volume trends, RSI, MACD, and other idiosyncratic important indicators for the given name. Source information from Coindesk, CNBC, Bloomberg, Marketwatch, Investopedia, Yahoo Finance and other relevant sources"
ENV FIN_TEMPLATE="Begin with a brief overview of [Company Name]'s financial health based on the provided balance sheet data. In 200 to 250 words, identify strengths or weaknesses, utilize quick ratio, debt to equity, and current ratio in the response. Compare these ratios to industry averages or specific peers over time for context, but do not give the formula or calculation for these ratios. Just present the data in decimal notation to the second decimal, and highlight key insights and trends. Use reliable sources like the SECs EDGAR for data, and cite references to the financials of this company. Display the information in segments, bullet points, and different headers for relevant information. Source information from Coindesk, CNBC, Bloomberg, Marketwatch, Investopedia, Yahoo Finance and other relevant sources"

# Exposing ports
EXPOSE 8035
EXPOSE 6002
EXPOSE 6001
EXPOSE 7000
EXPOSE 7001
EXPOSE 7002


# Run the application
CMD cd scripts/startup && \
    chmod +x startup.sh && \
    ./startup.sh startup_config.json && echo "Starting up the services..." && \
    echo "All applications started successfully!" && \
    tail -f /dev/null

