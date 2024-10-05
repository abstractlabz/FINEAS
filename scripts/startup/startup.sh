#!/bin/bash

# Function to check command availability
command_exists() {
  command -v "$1" >/dev/null 2>&1
}

# Ensure required commands are available
for cmd in jq wget tar python3 python3-pip gunicorn; do
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

# Ensure pip is available
if ! command_exists pip3; then
  echo "pip is not installed. Installing pip..."
  apt-get install -y python3-pip
  if [ $? -ne 0 ]; then
    echo "Error: Failed to install pip." >&2
    exit 1
  fi
fi

# Ensure Gunicorn is installed via pip
if ! pip3 show gunicorn >/dev/null 2>&1; then
  echo "Gunicorn is not installed. Installing Gunicorn..."
  pip3 install gunicorn
  if [ $? -ne 0 ]; then
    echo "Error: Failed to install Gunicorn." >&2
    exit 1
  fi
fi

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
cd ~ || { echo "Failed to navigate to the home directory"; exit 1; }
cd .. || { echo "Failed to navigate to the parent directory"; exit 1; }
rm -rf /usr/local/go
cd /tmp || { echo "Failed to navigate to the temp directory"; exit 1; }

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

source ~/.profile

# Move to the application startup directory
APP_DIR="../app/cmd/fineas-app"
cd "$APP_DIR" || { echo "Failed to navigate to the application directory"; exit 1; }

# Start the Go application
echo "Starting the Go application..."
go clean -cache -modcache -i -r
nohup go run main.go > "go_app.log" 2>&1 &

# Start the Python application
echo "Starting the Python application..."
nohup python3 main.py > "python_app.log" 2>&1 &

# Start automation and cron jobs
echo "Starting cron jobs..."
#cd "../../scripts/automation" || { echo "Failed to navigate to automation script directory"; exit 1; }
#nohup ./monitor_process.sh "monitor_query_config.json" > ../logs/monitor_process_query.log 2>&1 &
#nohup ./monitor_process.sh "monitor_upgrade_config.json" > ../logs/monitor_process_upgrade.log 2>&1 &
#nohup ./monitor_process.sh "monitor_data_config.json" > ../logs/monitor_process_data.log 2>&1 &

echo "All processes have started up..."
