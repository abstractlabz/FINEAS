# Use a base image
FROM ubuntu:latest

# Set the working directory
WORKDIR /app

# Copy the application files to the app directory
COPY . /app

# Install dependencies (if any) golang
RUN go mod download

# Install dependencies (if any) python
RUN pip install --no-cache-dir -r requirements.txt

# Exposing port 8080 for the entry point to the  golang backend
EXPOSE 8080

# Exposing port 5432 for llm inference server
EXPOSE 5432

# Exposing port 6002 for query inference server
EXPOSE 6002

# Go to cmd\fineas-app directory and run go run . to run golang backend
RUN cd cmd/fineas-app && go run .

# Go to cmd\fineas-app directory and run python main.py to run inference servers
RUN cd cmd/fineas-app && python main.py
