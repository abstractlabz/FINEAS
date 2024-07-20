#!/bin/bash

# Function to check command availability
command_exists() {
  command -v "$1" >/dev/null 2>&1
}

# Ensure required commands are available
for cmd in jq wget tar; do
  if ! command_exists "$cmd"; then
    echo "$cmd is not installed. Installing $cmd..."
    apt-get update
    apt-get install -y "$cmd"
    if [ $? -ne 0 ]; then
      echo "Error: Failed to install $cmd." >&2
      exit 1
    fi
  fi
done

# Ensure the config file path is provided as an argument
if [ "$#" -ne 1 ]; then
  echo "Usage: $0 <path_to_config_file>"
  exit 1
fi

CONFIG_FILE="$1"

# Function to read a value from the JSON file
get_config_value() {
  local key=$1
  jq -r ".$key" "$CONFIG_FILE"
}

# Read values from the configuration file and assign them to variables
export API_KEY=$(get_config_value "API_KEY")
export PASS_KEY=$(get_config_value "PASS_KEY")
export MONGO_DB_LOGGER_PASSWORD=$(get_config_value "MONGO_DB_LOGGER_PASSWORD")
export OPEN_AI_API_KEY=$(get_config_value "OPEN_AI_API_KEY")
export KB_WRITE_KEY=$(get_config_value "KB_WRITE_KEY")
export MR_WRITE_KEY=$(get_config_value "MR_WRITE_KEY")
export PINECONE_API_KEY=$(get_config_value "PINECONE_API_KEY")
export STRIPE_ENDPOINT_SECRET=$(get_config_value "STRIPE_ENDPOINT_SECRET")
export STRIPE_SECRET_KEY=$(get_config_value "STRIPE_SECRET_KEY")
export REDIRECT_DOMAIN=$(get_config_value "REDIRECT_DOMAIN")

echo "Environment variables set."

# Install Go
echo "Installing Go..."
GO_VERSION="1.21.6"
GO_TAR="go${GO_VERSION}.linux-amd64.tar.gz"
GO_URL="https://golang.org/dl/${GO_TAR}"

wget "$GO_URL"
if [ $? -ne 0 ]; then
  echo "Failed to download Go." >&2
  exit 1
fi

tar -C /usr/local -xzf "$GO_TAR"
if [ $? -ne 0 ]; then
  echo "Failed to extract Go." >&2
  exit 1
fi

rm "$GO_TAR"
export PATH=$PATH:/usr/local/go/bin
echo "Go installed."

# Make Go path persistent
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile

# Move to the application startup directory
APP_DIR="../../cmd/fineas-app"
cd "$APP_DIR" || { echo "Failed to navigate to the application directory"; exit 1; }

# Start the Go application
echo "Starting the Go application..."
nohup go run main.go > ../../scripts/logs/go_app.log 2>&1 &

# Start the Python application
echo "Starting the Python application..."
nohup python3 main.py > ../../scripts/logs/python_app.log 2>&1 &

# Start automation and cron jobs
echo "Starting cron jobs..."
cd - > /dev/null
cd "../../scripts/automation" || { echo "Failed to navigate to automation script directory"; exit 1; }
nohup ./monitor_process.sh "monitor_query_config.json" > ../logs/monitor_process_query.log 2>&1 &
nohup ./monitor_process.sh "monitor_upgrade_config.json" > ../logs/monitor_process_upgrade.log 2>&1 &

echo "All processes have started up..."
