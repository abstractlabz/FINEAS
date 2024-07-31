#!/bin/bash

# Function to check command availability
command_exists() {
  command -v "$1" >/dev/null 2>&1
}

# Ensure required commands are available
for cmd in jq wget tar curl; do
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

# Check if a config file path is provided as an argument
if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <path_to_config_file>"
    exit 1
fi

# Assign the first argument to CONFIG_FILE
CONFIG_FILE="$1"

# Your script logic here using the CONFIG_FILE
echo "Using config file: $CONFIG_FILE"

# Function to read a value from the JSON file
get_config_value() {
  local key=$1
  jq -r ".$key" "$CONFIG_FILE"
}

# Read values from the configuration file and assign them to variables
# Necessary to keep environment variables persistent for ssh connections

export PROCESS_URL=$(get_config_value "PROCESS_URL")
export PROCESS_DIR=$(get_config_value "PROCESS_DIR")
export INPUT_SCRIPT=$(get_config_value "INPUT_SCRIPT")
export LOG_FILE=$(get_config_value "LOG_FILE")
export RETRY_LIMIT=$(get_config_value "RETRY_LIMIT")
export RETRY_DELAY=$(get_config_value "RETRY_DELAY")
export AUTH_BEARER=$(get_config_value "AUTH_BEARER")
export COMMAND_PREFIX=$(get_config_value "COMMAND_PREFIX")


echo "Environment variables set."

# Ensure the log file directory exists
mkdir -p "$(dirname "$LOG_FILE")"

# Function to check if the process is up
check_process() {
    response=$(curl -k --write-out '%{http_code}' --silent --output /dev/null -X POST "${PROCESS_URL}" -H "Authorization: Bearer ${AUTH_BEARER}")
    curl_exit_status=$?
    
    if [[ $curl_exit_status -ne 0 ]]; then
        echo "$(date): Process is down (curl failed with exit status $curl_exit_status)" | tee -a "$LOG_FILE"
        return 1
    elif [[ "$response" -ne 400 && "$response" -ne 401 && "$response" -ne 500 && "$response" -ne 524 && "$response" -ne 521 && "$response" -ne 522 && "$response" -ne 523 ]]; then
        echo "$(date): Process is down with response code $response" | tee -a "$LOG_FILE"
        return 1
    elif [[ "$response" -eq 200 ]]; then
        echo "$(date): Process is up with response code $response" | tee -a "$LOG_FILE"
        return 0
    else
        echo "$(date): Process is up with response code $response" | tee -a "$LOG_FILE"
        return 0
    fi
}

# Function to kill the process
kill_process() {
    pid=$(pgrep -f "${INPUT_SCRIPT}")
    if [ -n "$pid" ]; then
        echo "$(date): Sending SIGTERM to process with PID $pid" | tee -a "$LOG_FILE"
        kill $pid
        sleep 5  # Wait for a few seconds to allow graceful shutdown
        
        # Check if the process is still running
        if pgrep -f "${INPUT_SCRIPT}" > /dev/null; then
            echo "$(date): Process did not terminate, sending SIGKILL to PID $pid" | tee -a "$LOG_FILE"
            kill -9 $pid
        fi
    else
        echo "$(date): No process found" | tee -a "$LOG_FILE"
    fi
}

# Function to start the process
start_process() {
    echo "$(date): Starting process" | tee -a "$LOG_FILE"

    # Check if PROCESS_DIR is set and is a valid directory
    if [ -z "$PROCESS_DIR" ]; then
        echo "$(date): PROCESS_DIR is not set." | tee -a "$LOG_FILE"
        return 1
    elif [ ! -d "$PROCESS_DIR" ]; then
        echo "$(date): PROCESS_DIR ($PROCESS_DIR) does not exist." | tee -a "$LOG_FILE"
        return 1
    fi

    # Check if INPUT_SCRIPT is set and is a valid file
    if [ -z "$INPUT_SCRIPT" ]; then
        echo "$(date): INPUT_SCRIPT is not set." | tee -a "$LOG_FILE"
        return 1
    fi

    # Save the current directory
    current_dir=$(pwd)

    # Change to the process directory
    cd "$PROCESS_DIR" || { echo "$(date): Failed to navigate to PROCESS_DIR ($PROCESS_DIR)"; return 1; }

    cmd_script="$COMMAND_PREFIX $INPUT_SCRIPT"
    # Start the process
    nohup $cmd_script &>> "$LOG_FILE" &
    
    # Return to the original directory
    cd "$current_dir" || { echo "$(date): Failed to return to the original directory"; return 1; }

    sleep 10  # Give it some time to start up

    # Recheck if the process started
    pid=$(pgrep -f "${INPUT_SCRIPT}")
    if [ -n "$pid" ]; then
        echo "$(date): Process started successfully with PID $pid" | tee -a "$LOG_FILE"
    else
        echo "$(date): Failed to start process" | tee -a "$LOG_FILE"
        return 1
    fi
}

# Function to restart the process with retry logic
restart_process() {
    for (( i=1; i<=$RETRY_LIMIT; i++ )); do
        kill_process
        sleep 5
        start_process
        
        # Check if the process started successfully
        
        if pgrep -f "${INPUT_SCRIPT}" > /dev/null; then
            echo "$(date): Process restarted successfully on attempt $i" | tee -a "$LOG_FILE"
            return 0
        else
            echo "$(date): Failed to start process on attempt $i" | tee -a "$LOG_FILE"
            sleep $RETRY_DELAY
        fi
    done

    # If we reach here, all retries have failed
    echo "$(date): Process failed to start after $RETRY_LIMIT attempts" | tee -a "$LOG_FILE"
}

# Monitor loop
while true; do
    check_process
    if [ $? -ne 0 ]; then
        restart_process
    fi
    # Wait for 10 minutes before checking again
    sleep 600
done
