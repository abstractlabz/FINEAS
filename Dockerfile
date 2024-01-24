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
RUN pip3 install --no-cache-dir -r utils/requirements.txt

# Set environment variables for keys
ENV API_KEY=API_KEY
ENV PASS_KEY=PASS_KEY
ENV MONGO_DB_LOGGER_PASSWORD=MONGO_DB_LOGGER_PASSWORD
ENV OPEN_AI_API_KEY=OPEN_AI_API_KEY
ENV KB_WRITE_KEY=KB_WRITE_KEY
ENV MR_WRITE_KEY=MR_WRITE_KEY
ENV PINECONE_API_KEY=PINECONE_API_KEY 

# Set environment variables for templates
ENV STK_SERVICE_URL=http://0.0.0.0:8081
ENV FIN_SERVICE_URL=http://0.0.0.0:8082
ENV NEWS_SERVICE_URL=http://0.0.0.0:8083
ENV DESC_SERVICE_URL=http://0.0.0.0:8084
ENV LLM_SERVICE_URL=http://0.0.0.0:5432
ENV TA_SERVICE_URL=http://0.0.0.0:8089
ENV YTD_TEMPLATE="Give a simple explanation of the companys stock performance based on the following year to date stock data in 15 to 25 words.Make sure to explicitly state the returns and mention the company name. At the end of the summary categorize this stock perfomance into only one of five categories, Very Bad, Bad, Neutral, Good, and Very Good: \n"
ENV NEWS_TEMPLATE="Give a simple explanation of these news headlines based on the following news headline data in 100 to 200 words. Explain what this means for the present and future for the company."
ENV DESC_TEMPLATE="Give a response of the description of this company. Ignore any seemingly random strings, just respond with the informative description."
ENV TA_TEMPLATE="Give an in depth analysis of the company stock ticker future price action based off of the recent stock information and following technical indicators."

# Exposing ports
EXPOSE 8035
EXPOSE 6002


# Run the application
CMD cd cmd/fineas-app && \
    go run . && echo "Waiting for the Go application to start..." && \
    python3 main.py && echo "Waiting for the Python application to start..." && \
    cd ../../scripts/populatedata && \
    go run . && echo "Starting up data populator..." && \
    echo "All applications started successfully!"

