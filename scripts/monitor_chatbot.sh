#!/bin/bash

# Variables
CHATBOT_URL="https://query.fineasapp.io:443/chat"  # Replace with your actual URL
DUMMY_PROMPT="Hello"
CHATBOT_DIR="../cmd/fineas-app"
MAIN_SCRIPT="main.py"

# Function to check if the chatbot is up
check_chatbot() {
    response=$(curl -k --write-out '%{http_code}' --silent --output /dev/null -X POST "${CHATBOT_URL}?prompt=${DUMMY_PROMPT}" -H "Authorization: Bearer 671b31a4e4d59e1f4e344e91fb343c6988462a0afcf828bcd3f55404058819f2")
    curl_exit_status=$?
    
    if [[ $curl_exit_status -ne 0 ]]; then
        echo "$(date): Chatbot is down (curl failed with exit status $curl_exit_status)"
        return 1
    elif [[ "$response" -ne 400 && "$response" -ne 401 && "$response" -ne 500 && "$response" -ne 524 && "$response" -ne 521 && "$response" -ne 522 && "$response" -ne 523 ]]; then
        echo "$(date): Chatbot is up with response code $response"
        return 0
    else
        echo "$(date): Chatbot is down with response code $response"
        return 1
    fi
}

# Function to kill the chatbot process
kill_chatbot() {
    pid=$(pgrep -f "${MAIN_SCRIPT}")
    if [ -n "$pid" ]; then
        echo "$(date): Killing chatbot process with PID $pid"
        kill -9 $pid
    else
        echo "$(date): No chatbot process found"
    fi
}

# Function to start the chatbot
start_chatbot() {
    echo "$(date): Starting chatbot"
    cd "$CHATBOT_DIR"
    nohup python3 "$MAIN_SCRIPT" &> chatbot.log &
    cd - > /dev/null  # Return to the previous directory
}

# Monitor loop
while true; do
    check_chatbot
    if [ $? -ne 0 ]; then
        kill_chatbot
        sleep 5
        start_chatbot
    fi
    # Wait for 2 minutes before checking again
    sleep 600
done
