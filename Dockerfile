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
ENV YTD_TEMPLATE="Conduct a technical analysis of [COMPANY_NAME]'s recent stock price movements. Your response should include: \ 
Stock movement information: YTD, MOM change etc. \ 
Trend Analysis: \ 
Analysis: Identify and describe the current price trend (uptrend, downtrend, sideways). Discuss the strength and duration of the trend. \ 
Support and Resistance Levels: \
Analysis: Determine key support and resistance levels and explain their significance. \ 
Price Levels (present numbers as bullet points): \ 
Support Levels: \ 
Level 1: $[SUPPORT_LEVEL_1] \ 
Level 2: $[SUPPORT_LEVEL_2] \ 
Resistance Levels: \ 
Level 1: $[RESISTANCE_LEVEL_1] \ 
Level 2: $[RESISTANCE_LEVEL_2] \ 
Technical Indicators: \ 
Analysis: Evaluate technical indicators such as RSI, MACD, moving averages, and other relevant indicators. Interpret what these indicators suggest about future price movements. \ 
Indicator Values (present numbers as bullet points): Use a minimum 5. \ 
Outlook: \ 
Analysis: Based on the above factors, provide an outlook on whether the stock is likely to be bullish or bearish in the near term. " 
ENV NEWS_TEMPLATE="Provide a comprehensive analysis of recent news articles related to [COMPANY_NAME]. Your response should include: \ 
Overall Sentiment Analysis: \ 
Analysis: Determine the overall sentiment (bullish, bearish, neutral) of the news affecting [COMPANY_NAME]. Discuss key themes and implications for the company's stock performance. \ 
Key News Highlights: \ 
Analysis: Summarize the most impactful news items, explaining their significance in the context of the company's operations and industry trends. \ 
News Items (present details as bullet points): \ 
[NEWS_HEADLINE_1]: Brief description and analysis of its impact on the company. Sentiment Scale (Highly Bearish to Highly Bullish). \ 
[NEWS_HEADLINE_2]: Brief description and analysis of its impact on the company. Sentiment Scale (Highly Bearish to Highly Bullish). \ 
[NEWS_HEADLINE_3]: Brief description and analysis of its impact on the company. Sentiment Scale (Highly Bearish to Highly Bullish). \ 
[NEWS_HEADLINE_4]: Brief description and analysis of its impact on the company. Sentiment Scale (Highly Bearish to Highly Bullish). \ 
[NEWS_HEADLINE_5]: Brief description and analysis of its impact on the company.  Sentiment Scale (Highly Bearish to Highly Bullish). \ 
(Include additional news items as necessary.) \ 
Market Reaction: \ 
Analysis: Evaluate how the market has reacted to these news items, considering stock price movements and trading volumes. \ 
Market Data (present numbers as bullet points): \ 
Stock Price Change: [PRICE_CHANGE]% since the news release. \ 
Trading Volume Change: [VOLUME_CHANGE]% compared to average volume. \ 
Macro and Micro Implications: \ 
Analysis: Discuss the broader macroeconomic themes ([MACRO_THEMES]) and microeconomic factors ([MICRO_FACTORS]) highlighted by the news, and how they affect [COMPANY_NAME]."  
ENV DESC_TEMPLATE=" Business Description: \ 
Analysis: In 100 words, explain the business and its key business lines and how they contribute to the company's value chain. \ 
Bulleted Information about the Industry, Year Founded, CEO, MarketCap, Employee Count, Headquarters, and other pertinent information if available (if not disregard, don't return N/A) \ 
Market Positioning: \ 
Analysis: Describe the company's target markets and geographical presence in paragraph form, detailing where the company operates and its main consumer demographics. Include how these factors align with its strategic goals and market dominance. \ 
Market Share: Provide details about the company’s market share in key sectors it dominates. \ 
Business Line Segmentation & Revenue Generation: Provide a breakdown of the company’s key business lines, showing their contribution to total revenue, and how this compares to overall industry trends. Discuss strengths, weaknesses, opportunities, and threats. \ 
Competitive Advantages: \ 
Analysis: Identify and discuss the company's unique selling propositions, core competencies, and barriers to entry. Rank these in order of significance, from most important to least, and clarify whether each advantage is a big or small factor in protecting or enhancing the company’s market position. \ 
Industry Context: \ 
Analysis: Place the company or asset within the broader industry context, considering current trends such as growth, regulatory challenges, and technological advancements. Explain how these factors affect the company’s strategic direction and operations, particularly in comparison to key competitors." 
ENV TA_TEMPLATE="“Perform an in-depth technical analysis of [COMPANY_NAME]'s stock. Your response should include: \ 
1. Chart Patterns: \ 
Analysis: Identify any significant chart patterns (e.g., head and shoulders, double bottom, triangles) and discuss their implications. \ 
2. Volume Analysis: \ 
Analysis: Examine trading volumes to assess the strength of price movements. \ 
Volume Data (present numbers as bullet points): \ 
Average Trading Volume: [AVERAGE_VOLUME] \ 
Current Trading Volume: [CURRENT_VOLUME] \ 
3. Technical Indicators: \ 
Analysis: Evaluate key technical indicators and oscillators. \ 
Indicator Values (present numbers as bullet points): \ 
Bollinger Bands: Current position of price within the bands. \ 
Stochastic Oscillator: [STOCHASTIC_VALUE] \ 
Fibonacci Retracement Levels: Key levels at [LEVELS] \ 
4. Momentum and Trend Strength: \ 
Analysis: Assess the momentum and strength of the current trend using indicators like ADX and other relevant indicators. \ 
Momentum Data (present numbers as bullet points): \ 
5. Support and Resistance Levels: \ 
Analysis: Identify key support and resistance zones based on historical price action, compared to the current price level. \ 
Support Levels: [Level 1], [Level 2], [Level 3] \ 
Resistance Levels: [Level 1], [Level 2], [Level 3] \ 
6. Moving Averages: \ 
Analysis: Assess the position of key moving averages (e.g., 50-day, 200-day) relative to the stock price. \ 
Moving Average Values (present numbers as bullet points): \ 
50-day Moving Average: [VALUE] \ 
200-day Moving Average: [VALUE] \ 
7. Relative Strength Index (RSI): \ 
Analysis: Evaluate whether the stock is overbought or oversold based on the RSI. \ 
RSI Value: [VALUE] (Range 0-100) \ 
8. MACD (Moving Average Convergence Divergence): \ 
Analysis: Interpret the MACD line, signal line, and histogram. Note any crossovers and discuss implications for bullish or bearish sentiment. \ 
MACD Values: \ 
MACD Line: [VALUE] \ 
Signal Line: [VALUE] \ 
Histogram: [VALUE] \ 
9. Candlestick Patterns: \ 
Analysis: Identify key candlestick patterns such as doji, hammer, engulfing patterns, etc. Discuss their significance in predicting reversals or continuation. \ 
Key Patterns Observed: [PATTERNS] \ 
10. Sentiment Analysis: \ 
Analysis: Gauge market sentiment using tools like the Put/Call Ratio, Fear & Greed Index, or other sentiment indicators. \ 
Put/Call Ratio: [VALUE] \ 
Fear & Greed Index: [VALUE] \ 
11. Market Correlation: \ 
Analysis: Examine how the stock moves relative to the broader market or industry. \ 
Correlation Coefficient: [VALUE] (Range -1 to 1) \ 
12. Earnings and News Impact: \ 
Analysis: Assess how upcoming earnings releases or recent news might affect the stock’s technical outlook. \ 
Earnings Date: [DATE] \ 
Recent News Events: [EVENTS] \ 
13. Historical Volatility: \ 
Analysis: Evaluate the stock's historical volatility to understand its risk profile and potential price swings. \ 
Volatility Data (present numbers as bullet points): \ 
1-Month Historical Volatility: [VALUE] \ 
3-Month Historical Volatility: [VALUE] \ 
14. Seasonality and Cyclical Trends: \ 
Analysis: Discuss any seasonal or cyclical patterns that may impact the stock’s performance. \ 
Historical Seasonal Trends: [TREND DESCRIPTION] \ 
15. Risk Management Metrics: \ 
Analysis: Use risk metrics such as Sharpe Ratio, Beta, and Maximum Drawdown to assess the risk-reward profile of the stock. \ 
Risk Data: \ 
Sharpe Ratio: [VALUE] \ 
Beta (vs S&P 500): [VALUE] \ 
Maximum Drawdown: [VALUE] \ 
16. Outlook: \ 
Analysis: Based on the technical factors above, provide a forecast on the stock's potential movement, indicating if signals are bullish, bearish, or neutral.”"

ENV FIN_TEMPLATE="Provide a detailed analysis of [COMPANY_NAME], which operates in [KEY_BUSINESS_LINES]. Your response should include (if available if not disregard, do not return N/A, or choose a close alternative). If the asset is a crypto currency this section should focus on tokenomics: \  
Overview of Business Segments: \ 
Analysis: Discuss all business segments, including core segments like [LARGEST_SEGMENT], and their respective contributions to revenue. Connect macro themes such as [MACRO_THEME] and micro themes like [MICRO_THEME]. \ 
Segments and Revenue Contributions (present numbers as bullet points): \ 
[LARGEST_SEGMENT]: [SHORT_DESCRIPTION_1] – contributes [PERCENTAGE]% of revenue. \ 
[SEGMENT_2]: [SHORT_DESCRIPTION_2] – contributes [PERCENTAGE_2]% of revenue. \ 
[SEGMENT_3]: [SHORT_DESCRIPTION_3] – contributes [PERCENTAGE_3]% of revenue. \ 
(Include additional segments as necessary.) \ 
Growth Analysis: \ 
Analysis: Discuss the fastest-growing segment, [GROWING_SEGMENT], and how it's influenced by market trends. \ 
Growth Data (present numbers as bullet points): \ 
Growth Rate: [GROWTH_RATE]% \ 
Market Trend Driving Growth: [MARKET_TREND] \ 
Profitability Metrics: \ 
Analysis: Evaluate the company's profitability compared to industry averages. \ 
Profitability Data (present numbers as bullet points): \ 
Net Margin: [NET_PROFIT_MARGIN]% (Industry average: [INDUSTRY_AVG_NET_MARGIN]%) \ 
Operating Margin: [OPERATING_MARGIN]% (Industry average: [INDUSTRY_AVG_OP_MARGIN]%) \ 
Earnings Per Share (EPS): $[EPS] \ 
Financial Health Indicators: \ 
Analysis: Assess the company's financial stability, liquidity, and leverage. \ 
Financial Data (present numbers as bullet points): \ 
Debt-to-Equity Ratio: [D_E_RATIO] (Industry average: [INDUSTRY_AVG_DE_RATIO]) \ 
Free Cash Flow (FCF): $[FCF_AMOUNT] \ 
Current Ratio: [CURRENT_RATIO] (Industry average: [INDUSTRY_AVG_CURRENT_RATIO]) \ 
Derived Analysis: \ 
Summarize key insights, highlighting potential risks and opportunities."

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
