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
ENV YTD_TEMPLATE="Based on [Company Name]'s stock data for this year, provide a simple explanation of its performance. Mention the percentage change in the stock price, any notable trends, and key factors that influenced its performance. Conclude by categorizing the performance into one of these five categories, specifying the percentage ranges for each: Very Positive (over +15%), Positive (+5% to +15%), Moderate (-5% to +5%), Negative (-15% to -5%), Very Negative (below -15%), Keep the summary brief and easy to understand. Use reliable sources like CNBC, Bloomberg, or Yahoo Finance to cite relevant information."
ENV NEWS_TEMPLATE="Provide a quick assessment of recent news headlines about [Company Name], indicating their impact using terms like Positive, Negative, or Moderate. Then, in simple terms (100-150 words), explain why these news items are significant. Focus on how they might affect the company's situation and future. Summarize the most important news, considering market trends, the company's financial health, and industry developments. Conclude with potential long-term implications and the overall impact of the news. Use reputable sources like CNBC, Bloomberg, or Yahoo Finance for information."
ENV DESC_TEMPLATE="Present a concise overview of [Company Name] for potential investors, covering key points in bullet form (150-200 words). Include all: Company Name, What the Company Does (in simple terms), Year Founded, CEO, Number of Employees, Recent Achievements or Milestones, Basic Financial Metrics (e.g., revenue growth) in bulleted format with necessary information. Add sections for: Key Challenges and Recent News Impact. Aim to provide a balanced view that links the company's financial health and market position with current events. Use easy-to-understand language and reputable sources like CNBC or Yahoo Finance."
ENV TA_TEMPLATE="Provide an understandable analysis of [Company Name]'s stock price movements based on recent data, using simple explanations of technical indicators like Moving Averages (SMA & EMA), Volume Trends, RSI, and MACD. Conclude with an overall outlook categorized as Positive, Negative, or Moderate. Use straightforward language to explain what these indicators suggest about future price action. Reference reputable sources like Investopedia for definitions and Yahoo Finance for data."
ENV FIN_TEMPLATE="Provide a brief overview of [Company Name]'s financial health based on recent balance sheet data. In 150-200 words, highlight key strengths or weaknesses. Mention important financial ratios like the Current Ratio and Debt-to-Equity Ratio, explaining what they mean in simple terms without detailed calculations. Compare these ratios to industry averages for context. Present the information using bullet points and clear headers. Use reliable sources like Yahoo Finance or Investopedia for data."

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

