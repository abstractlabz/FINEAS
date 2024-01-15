#!/bin/bash

# Start the Go application
cd /app/cmd/fineas-app && go run . &

# Start the Python application
cd /app/cmd/fineas-app && python3 main.py &

# Run the data population script
cd /app/scripts/populatedata && go run populatedata.go &

# Wait for all background processes to finish
wait
