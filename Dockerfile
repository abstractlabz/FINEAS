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
ENV CLAUDE_API_KEY=CLAUDE_API_KEY
ENV KB_WRITE_KEY=KB_WRITE_KEY
ENV MR_WRITE_KEY=MR_WRITE_KEY
ENV PINECONE_API_KEY=PINECONE_API_KEY 
ENV STRIPE_ENDPOINT_SECRET=STRIPE_ENDPOINT_SECRET
ENV STRIPE_SECRET_KEY=STRIPE_SECRET_KEY
ENV REDIRECT_DOMAIN=REDIRECT_DOMAIN
ENV PINECONE_HOST=PINECONE_HOST
ENV OPENAI_API_KEY=OPENAI_API_KEY

# Set environment variables for templates
ENV STK_SERVICE_URL=http://0.0.0.0:8081
ENV FIN_SERVICE_URL=http://0.0.0.0:8082
ENV NEWS_SERVICE_URL=http://0.0.0.0:8083
ENV DESC_SERVICE_URL=http://0.0.0.0:8084
ENV LLM_SERVICE_URL=http://0.0.0.0:5432
ENV TA_SERVICE_URL=http://0.0.0.0:8089
ENV YTD_TEMPLATE="# Conduct an analysis of [ASSET_NAME]'s recent price movements. \
When referring to [ASSET_NAME], replace that with the ticker symbol or the full name of the asset. \
Provide the annotation information throughout your response, including only the available links, search headers, and other information to provide more context to the yearly price information report. \
Only base your response on the information available to you and only present the response in a header and bullet point format: \
Represent all numbers to two decimal places .00 [Units], using numerical short scales (thousand, M for million, B for billion, T for trillion, etc.). \
*Fill in the following values from available data and sources *Omit the following placeholders in the brackets from the response when the values replace them*:\
*Omit any calculations or special mathematical notation in your response*:\
*Include high-level signals such as 'highly bearish,' 'bearish,' 'neutral,' 'bullish,' or 'highly bullish' where relevant.* \
## Current Position: \
- $[LAST_CLOSING_PRICE] \
- [YEAR_OVER_YEAR_CHANGE]% \
## Key Insights Analysis: \
*Include high-level signals such as 'highly bearish,' 'bearish,' 'neutral,' 'bullish,' or 'highly bullish' where relevant.* \
- Indicate what the price value along with the year over year change might suggest about market conditions. \
*Omit any calculations or special mathematical notation in your response* \
*Include all relevant annotation URLs in the format [TITLE]url with no spaces or line breaks. \
If a URL comes with a title or description, place the title inside brackets [TITLE_HERE] immediately followed by the URL, for example [SomeArticleTitle]https://example.com. \  "

ENV NEWS_TEMPLATE="# Provide a comprehensive analysis of recent news articles related to [ASSET_NAME]. When referring to [ASSET_NAME], replace that with the ticker symbol or the full name of the asset. Provide the annotation information throught your response including only the avaliable links, search headers, and other information to provide more context to the yearly price information report. Only base your response on the information avaliable to you. \
*If any information is not available, please ignore it. Don't even include it in your response. Represent all numbers to the second decimal point .00 [Units] after and display numbers using numerical short scales (thousand, million, billion, trillion, etc)*.\ 
## Overall Sentiment Analysis:
- Determine the overall sentiment (bullish, bearish, neutral) of the news affecting [ASSET_NAME].\ 
## Key News Highlights: Summarize the most impactful news items.\ 
## News Items (5 news items): \
 *If it is avaliable, provide all relevant annotation url information throughout your response. If a url comes with a title or description, provide the title inside brackets [TITLE_HERE] and the url next to the title with no spaces in between and on the same line. Never leave a space or line break between the title and the url. \
 - *Fill in the following values from available data and sources *Omit the following placeholders in the brackets from the response when the values replace them*:\
  [FIRST_TITLE_HERE]url:\
 - Provide a brief description of the news story as a bullet point on a new line.\ 
 - Provide an impact analysis to further contexualize the news and its impact on business, and social expectations as a bullet point on a new line. \
 - Evaluate how the market has reacted to these news items as a bullet point on a new line. Determine the overall sentiment (highly bullish to highly bearish) of the news affecting [ASSET_NAME].\ 
*If it is avaliable, provide all relevant annotation url information throughout your response. If a url comes with a title or description, provide the title inside brackets [TITLE_HERE] and the url next to the title with no spaces in between and on the same line. Never leave a space or line break between the title and the url. \
 [SECOND_TITLE_HERE]url:\
 - Provide a brief description of the news story as a bullet point on a new line.\ 
 - Provide an impact analysis to further contexualize the news and its impact on business, and social expectations as a bullet point on a new line.\ 
 - Evaluate how the market has reacted to these news items as a bullet point on a new line. Determine the overall sentiment (highly bullish to highly bearish) of the news affecting [ASSET_NAME].\ 
*If it is avaliable, provide all relevant annotation url information throughout your response. If a url comes with a title or description, provide the title inside brackets [TITLE_HERE] and the url next to the title with no spaces in between and on the same line. Never leave a space or line break between the title and the url. \
 [THIRD_TITLE_HERE]url:\
 - Provide a brief description of the news story as a bullet point on a new line.\ 
 - Provide an impact analysis to further contexualize the news and its impact on business, and social expectations as a bullet point on a new line.\ 
 - Evaluate how the market has reacted to these news items as a bullet point on a new line. Determine the overall sentiment (highly bullish to highly bearish) of the news affecting [ASSET_NAME].\ 
*If it is avaliable, provide all relevant annotation url information throughout your response. If a url comes with a title or description, provide the title inside brackets [TITLE_HERE] and the url next to the title with no spaces in between and on the same line. Never leave a space or line break between the title and the url. \
 [FOURTH_TITLE_HERE]url:\
 - Provide a brief description of the news story as a bullet point on a new line.\ 
 - Provide an impact analysis to further contexualize the news and its impact on business, and social expectations as a bullet point on a new line.\ 
 - Evaluate how the market has reacted to these news items as a bullet point on a new line. Determine the overall sentiment (highly bullish to highly bearish) of the news affecting [ASSET_NAME].\ 
*If it is avaliable, provide all relevant annotation url information throughout your response. If a url comes with a title or description, provide the title inside brackets [TITLE_HERE] and the url next to the title with no spaces in between and on the same line. Never leave a space or line break between the title and the url. \
 [FIFTH_TITLE_HERE]url:\
 - Provide a brief description of the news story as a bullet point on a new line.\ 
 - Provide an impact analysis to further contexualize the news and its impact on business, and social expectations as a bullet point on a new line.\ 
 - Evaluate how the market has reacted to these news items as a bullet point on a new line. Determine the overall sentiment (highly bullish to highly bearish) of the news affecting [ASSET_NAME]. *If any information is not available, please ignore it. Represent all numbers to the second decimal point .00 [Units] after and display numbers using numerical short scales (thousand, M for million, B for billion, T for trillion, etc)*."

ENV DESC_TEMPLATE="# Business Description:\ 
When referring to [ASSET_NAME], replace that with the ticker symbol or the full name of the asset. Provide the annotation information throught your response including only the avaliable links, search headers, and other information to provide more context to the yearly price information report. Only base your response on the information avaliable to you. All information avaliable to you is accurate and relevant, omit any references questioning the accuracy of the information in your response. *If any information is not available, please ignore it. Don't even include it in your response. Represent all numbers to the second decimal point .00 [Units] after and display numbers using numerical short scales (thousand, M for million, B for billion, T for trillion, etc)*.\ 
 ## Company Description: - In 50-100 words, explain the business and its key business focus and business lines, including their composition of revenue.\ 
 ## Company Information: - Fill in the following values from available data and sources *Omit the following placeholders in the brackets from the response when the values replace them*:\ 
 - CEO: [CEO]\ 
 - Year Founded: [YEAR_FOUNDED]\ 
 - Market Cap: $[MARKET_CAP]\ 
 - Sector/Industry: [SECTOR_INDUSTRY]\ 
 - Employees Count: [EMPLOYEES_COUNT]\ 
 - HQ: [HQ]\ 
 - Include more nuanced information about the company.\ 
 ## Market Share: - Provide a breakdown of the company's market share:\ 
 - *If this information is not available, please entirely omit this section. Don't even include it in your response. \
 - North America: [MARKET_SHARE_NA]%\ 
 - Europe: [MARKET_SHARE_EU]%\ 
 - Asia-Pacific: [MARKET_SHARE_APAC]%\ 
 - Other Regions: [MARKET_SHARE_OTHER]%\ 
 - **By Business Line (% Composition)**:\ 
 - [BUSINESS_LINE_1]: [MARKET_SHARE_LINE_1]%\ 
 - [BUSINESS_LINE_2]: [MARKET_SHARE_LINE_2]%\ 
 - [BUSINESS_LINE_3]: [MARKET_SHARE_LINE_3]%\ 
 - Highlight where the company ranks among competitors in its key business segments. \
 ## Market Positioning: - Describe the company's target markets and geographical presence in detail: \
 - Key geographic areas of focus: [KEY_REGIONS] \
 - Market demographics or customer segments: [CUSTOMER_SEGMENTS] \
 - Current strategic focus or areas of opportunity, based on Management Discussion and Analysis (MD&A): \
 - [BIGGEST_OPPORTUNITY] \
 - Explain why this is a significant growth driver or differentiator. \
## Competitive Advantages: - Identify and discuss the company's unique selling propositions: \
*Fill in the following placeholders with response analysis about the [ASSET_NAME]'s unique selling propositions on a new bullet point line for each advantage. *Omit the following placeholders from the response when the analysis replaces them*:\
 - ADVANTAGE_1 \
 - ADVANTAGE_2 \
 - ADVANTAGE_3 \
*If it is avaliable, provide all relevant annotation url information throughout your response. If a url comes with a title or description, provide the title inside brackets [TITLE_HERE] and the url next to the title with no spaces in between and on the same line. Never leave a space or line break between the title and the url. \
## Sentiment: 
- Determine the overall sentiment (highly bullish to highly bearish) for [ASSET_NAME], supported by the above analysis. *If any information is not available, please ignore it. Represent all numbers to the second decimal point .00 [Units] after and display numbers using numerical short scales (thousand, million, billion, trillion, etc)*."

ENV TA_TEMPLATE="Perform an in-depth technical analysis of [ASSET_NAME]'s stock. When referring to [ASSET_NAME], replace that with the ticker symbol or the full name of the asset. Provide the annotation information throught your response including only the avaliable links, search headers, and other information to provide more context to the yearly price information report. Only base your response on the information avaliable to you. All information avaliable to you is accurate and relevant, omit any references questioning the accuracy of the information in your response. *If any information is not available, please ignore it. Don't even include it in your response. Represent all numbers to the second decimal point .00 [Units] after and display numbers using numerical short scales (thousand, million, billion, trillion, etc)*.\ 
## Chart Patterns: - Identify any significant chart patterns.\ 
## Volume Analysis: - Examine trading volumes to assess the strength of price movements.\ 
## Technical Indicators: - Evaluate key technical indicators and oscillators. Determine the overall sentiment (highly bullish to highly bearish) of the technicals for [ASSET_NAME]. *If any information is not available, please ignore it. Represent all numbers to the second decimal point .00 [Units] after and display numbers using numerical short scales (thousand, M for million, B for billion, T for trillion, etc)*.
"

ENV FIN_TEMPLATE="# Provide a detailed analysis of [ASSET_NAME] financial health and performance. When referring to [ASSET_NAME], replace that with the ticker symbol or the full name of the asset. Provide the annotation information throught your response including only the avaliable links, search headers, and other information to provide more context to the financial health information report. Only base your response on the information avaliable to you. All information avaliable to you is accurate and relevant, omit any references questioning the accuracy of the information in your response. \ 
*If any information is not available, please ignore it. Don't even include it in your response. Represent all numbers to the second decimal point .00 [Units] after and display numbers using numerical short scales (thousand, M for million, B for billion, T for trillion, etc)*. \ 
Never include any calculations nor special mathematical notation in your response. *Omit the following placeholders in the brackets from the response when the values replace them* \ 
## Financial Position Analysis: \ 
## Assets and Liabilities: \
- Total Assets: $[TOTAL_ASSETS] \ 
- Current Assets: $[CURRENT_ASSETS] \ 
- Total Liabilities: $[TOTAL_LIABILITIES] \ 
- Current Liabilities: $[CURRENT_LIABILITIES] \ 
- ACCOUNTS PAYABLE: $[ACCOUNTS_PAYABLE] \ 
- Total Equity: $[TOTAL_EQUITY] \ 
## Health Analysis: \  
- Working Capital: [CURRENT_ASSETS] - [CURRENT_LIABILITIES], *DONT INCLUDE ANY CALCULATIONS OR SPECIAL MATHEMATICAL NOTATION IN YOUR RESPONSE, ONLY INCLUDE THE VALUE* \ 
- Current Ratio: [CURRENT_ASSETS] / [CURRENT_LIABILITIES], *DONT INCLUDE ANY CALCULATIONS OR SPECIAL MATHEMATICAL NOTATION IN YOUR RESPONSE, ONLY INCLUDE THE VALUE* \ 
- Using both ratios interpret the company's liquidity position and its ability to meet short-term obligations as highly bearish, bearish, neutral, bullish, or highly bullish. \ 
## Growth Synopsis: \ 
- Explain how [ASSET_NAME] finances are positioned for investment relative to market conditions. \
*If it is avaliable, provide all relevant annotation url information throughout your response. If a url comes with a title or description, provide the title inside brackets [TITLE_HERE] and the url next to the title with no spaces in between and on the same line. Never leave a space or line break between the title and the url. \
*If any information is not available, please ignore it. Represent all numbers to the second decimal point .00 [Units] after and display numbers using numerical short scales (thousand, M for million, B for billion, T for trillion, etc)*."


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