#!/bin/bash

# Start the Go application
cd ../cmd/fineas-app && go run . 

echo "Waiting for the Go application to start..."

# Start the Python application
cd ../cmd/fineas-app && python3 main.py 

echo "Waiting for the Python application to start..."

# Run the data population script
cd ../scripts/populatedata && go run populatedata.go 

echo "Waiting for the data population script to finish..."

# Wait for all background processes to finish
wait

echo "All processes have started up..."
```


