# Use a base image
FROM ubuntu:latest

# Install necessary tools and libraries
RUN apt-get update && apt-get install -y \
    curl \
    git \
    python3 \
    python3-pip \
    dos2unix \
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

# Convert line endings for startup.sh
RUN dos2unix scripts/startup/startup.sh  # Convert line endings to LF

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
ENV STRIPE_ENDPOINT_SECRET=STRIPE_ENDPOINT_SECRET
ENV STRIPE_SECRET_KEY=STRIPE_SECRET_KEY
ENV REDIRECT_DOMAIN=REDIRECT_DOMAIN

# Set environment variables for templates
ENV STK_SERVICE_URL=http://0.0.0.0:8081
ENV FIN_SERVICE_URL=http://0.0.0.0:8082
ENV NEWS_SERVICE_URL=http://0.0.0.0:8083
ENV DESC_SERVICE_URL=http://0.0.0.0:8084
ENV LLM_SERVICE_URL=http://0.0.0.0:5432
ENV TA_SERVICE_URL=http://0.0.0.0:8089

ENV YTD_TEMPLATE="# Conduct an analysis of [ASSET_NAME]'s recent and historical price movements. Provide the annotation information throught your response including only the avaliable links, search headers, and other information to provide more context to the yearly price information report. Only base your response on the information avaliable to you. *If any information is not available, please ignore it. Don't even include it in your response. Represent all numbers to the second decimal point .00 [Units] after*.\ 
 Header: Stock Movement Information: - YTD, Week over Week Change, MOM change, and 1-year change - Provide a brief summary of notable price milestones or shifts over the years.\ 
 Header: Historical Price Trend Analysis: - Describe the historical trends of [ASSET_NAME] over different periods (e.g., monthly, yearly). - Highlight key price movements, including peaks and troughs, and discuss potential factors that influenced these movements.\ 
 Header: Price Levels: - Recent High: $[RECENT_HIGH]\ 
 - Recent Low: $[RECENT_LOW]\ 
 - All-Time High: $[ALL_TIME_HIGH]\ 
 - All-Time Low: $[ALL_TIME_LOW]\ 
 Header: Volatility Overview: - Describe the recent volatility trends and provide context on how the asset's price has fluctuated in both the short and long term.\ 
 Header: Outlook: - Based on the historical price movements and recent trends, provide an outlook on the asset's potential price direction in the near term. *If any information is not available, please ignore it. Represent all numbers to the second decimal point .00 [Units] after*."

ENV NEWS_TEMPLATE="# Provide a comprehensive analysis of recent news articles related to [ASSET_NAME].\ 
 Header: Overall Sentiment Analysis: - Determine the overall sentiment (bullish, bearish, neutral) of the news affecting [ASSET_NAME].\ 
 Header: Key News Highlights: - Summarize the most impactful news items.\ 
 Header: News Items (5 news items): \
  [NEWS_HEADLINE_1]: Provide a brief description of the news story.\ 
 - Provide an impact analysis to further contexualize the news and its impact on business, and social expectations. \
 - Evaluate how the market has reacted to these news items. Determine the overall sentiment (highly bullish to highly bearish) of the news affecting [ASSET_NAME].\ 
 [NEWS_HEADLINE_2]: Provide a brief description of the news story.\ 
 - Provide an impact analysis to further contexualize the news and its impact on business, and social expectations.\ 
 - Evaluate how the market has reacted to these news items. Determine the overall sentiment (highly bullish to highly bearish) of the news affecting [ASSET_NAME].\ 
 [NEWS_HEADLINE_3]: Provide a brief description of the news story.\ 
 - Provide an impact analysis to further contexualize the news and its impact on business, and social expectations.\ 
 - Evaluate how the market has reacted to these news items. Determine the overall sentiment (highly bullish to highly bearish) of the news affecting [ASSET_NAME].\ 
 [NEWS_HEADLINE_4]: Provide a brief description of the news story.\ 
 - Provide an impact analysis to further contexualize the news and its impact on business, and social expectations.\ 
 - Evaluate how the market has reacted to these news items. Determine the overall sentiment (highly bullish to highly bearish) of the news affecting [ASSET_NAME].\ 
 [NEWS_HEADLINE_5]: Provide a brief description of the news story.\ 
 - Provide an impact analysis to further contexualize the news and its impact on business, and social expectations.\ 
 - Evaluate how the market has reacted to these news items. Determine the overall sentiment (highly bullish to highly bearish) of the news affecting [ASSET_NAME]. *If any information is not available, please ignore it. Represent all numbers to the second decimal point .00 [Units] after*."

ENV DESC_TEMPLATE="# Business Description:\ 
 Header: Company Description: - In 50-100 words, explain the business and its key business focus and business lines, including their composition of revenue.\ 
 Header: Company Information: - Fill in the following values from available data and sources:\ 
 - CEO: [CEO]\ 
 - Year Founded: [YEAR_FOUNDED]\ 
 - Sector/Industry: [SECTOR_INDUSTRY]\ 
 - Employees Count: [EMPLOYEES_COUNT]\ 
 - Market Cap: $[MARKET_CAP]\ 
 - HQ: [HQ]\ 
 - Include more nuanced information about the company.\ 
 Header: Market Share: - Provide a breakdown of the company's market share:\ 
 - **By Region (% Composition)**:\ 
 - North America: [MARKET_SHARE_NA]%\ 
 - Europe: [MARKET_SHARE_EU]%\ 
 - Asia-Pacific: [MARKET_SHARE_APAC]%\ 
 - Other Regions: [MARKET_SHARE_OTHER]%\ 
 - **By Business Line (% Composition)**:\ 
 - [BUSINESS_LINE_1]: [MARKET_SHARE_LINE_1]%\ 
 - [BUSINESS_LINE_2]: [MARKET_SHARE_LINE_2]%\ 
 - [BUSINESS_LINE_3]: [MARKET_SHARE_LINE_3]%\ 
 - Highlight where the company ranks among competitors in its key business segments.\ 
 Header: Market Positioning: - Describe the company's target markets and geographical presence in detail:\ 
 - Key geographic areas of focus: [KEY_REGIONS]\ 
 - Market demographics or customer segments: [CUSTOMER_SEGMENTS]\ 
 - Current strategic focus or areas of opportunity, based on Management Discussion and Analysis (MD&A):\ 
 - [BIGGEST_OPPORTUNITY]\ 
 - Explain why this is a significant growth driver or differentiator.\ 
 Header: Competitive Advantages: - Identify and discuss the company's unique selling propositions:\ 
 - [ADVANTAGE_1]\ 
 - [ADVANTAGE_2]\ 
 - [ADVANTAGE_3]\ 
 Header: Sentiment: - Determine the overall sentiment (highly bullish to highly bearish) for [ASSET_NAME], supported by the above analysis. *If any information is not available, please ignore it. Represent all numbers to the second decimal point .00 [Units] after*."

ENV TA_TEMPLATE="# Perform an in-depth technical analysis of [ASSET_NAME]'s stock.\ 
 Header: Chart Patterns: - Identify any significant chart patterns.\ 
 Header: Volume Analysis: - Examine trading volumes to assess the strength of price movements.\ 
 Header: Technical Indicators: - Evaluate key technical indicators and oscillators. Determine the overall sentiment (highly bullish to highly bearish) of the technicals for [ASSET_NAME]. *If any information is not available, please ignore it. Represent all numbers to the second decimal point .00 [Units] after*."

ENV FIN_TEMPLATE="# Business Description:\ 
 Header: Company Description: - In 50-100 words, explain the business and its key business focus and business lines, including their composition of revenue.\ 
 Header: Company Information: - Fill in the following values from available data and sources:\ 
 - CEO: [CEO]\ 
 - Year Founded: [YEAR_FOUNDED]\ 
 - Sector/Industry: [SECTOR_INDUSTRY]\ 
 - Employees Count: [EMPLOYEES_COUNT]\ 
 - Market Cap: $[MARKET_CAP]\ 
 - HQ: [HQ]\ 
 - Include more nuanced information about the company.\ 
 Header: Market Share: - Provide a breakdown of the company's market share:\ 
 - **By Region (% Composition)**:\ 
 - North America: [MARKET_SHARE_NA]%\ 
 - Europe: [MARKET_SHARE_EU]%\ 
 - Asia-Pacific: [MARKET_SHARE_APAC]%\ 
 - Other Regions: [MARKET_SHARE_OTHER]%\ 
 - **By Business Line (% Composition)**:\ 
 - [BUSINESS_LINE_1]: [MARKET_SHARE_LINE_1]%\ 
 - [BUSINESS_LINE_2]: [MARKET_SHARE_LINE_2]%\ 
 - [BUSINESS_LINE_3]: [MARKET_SHARE_LINE_3]%\ 
 - Highlight where the company ranks among competitors in its key business segments.\ 
 Header: Market Positioning: - Describe the company's target markets and geographical presence in detail:\ 
 - Key geographic areas of focus: [KEY_REGIONS]\ 
 - Market demographics or customer segments: [CUSTOMER_SEGMENTS]\ 
 - Current strategic focus or areas of opportunity, based on Management Discussion and Analysis (MD&A):\ 
 - [BIGGEST_OPPORTUNITY]\ 
 - Explain why this is a significant growth driver or differentiator.\ 
 Header: Competitive Advantages: - Identify and discuss the company's unique selling propositions:\ 
 - [ADVANTAGE_1]\ 
 - [ADVANTAGE_2]\ 
 - [ADVANTAGE_3]\ 
 Header: Sentiment: - Determine the overall sentiment (highly bullish to highly bearish) for [ASSET_NAME], supported by the above analysis. *If any information is not available, please ignore it. Represent all numbers to the second decimal point .00 [Units] after*."

# Exposing ports
EXPOSE 8035
EXPOSE 6002
EXPOSE 6001
EXPOSE 7000
EXPOSE 7001
EXPOSE 7002

# Run the application
CMD cd scripts/startup && bash ./startup.sh startup_config.json && \
    echo "Starting up the services..." && \
    echo "All applications started successfully!" && \
    tail -f /dev/null