import subprocess
import os

if __name__ == '__main__':
    # Base directories
    current_directory = os.getcwd()
    desired_directory = os.path.abspath(os.path.join(current_directory, os.pardir, os.pardir))

    # Function to get SSL paths for a given process
    def get_ssl_paths(process_name):
        base_path = os.path.join(desired_directory, 'utils', 'keys', process_name)
        certfile = os.path.join(base_path, 'fullchain.pem')
        keyfile = os.path.join(base_path, 'privkey.pem')
        return certfile, keyfile

    # LLM services working directory
    llm_working_directory = os.path.join(desired_directory, "api", "llm")

    # LLM services commands (without SSL)
    llm_command = [
        "gunicorn", "-w", "4", "-b", "0.0.0.0:5432", 
        "--max-requests", "200", "--limit-request-line", "8190", "--timeout", "120", "llm:app"
    ]
    dataingestor_command = [
        "gunicorn", "-w", "4", "-b", "0.0.0.0:6001", 
        "--max-requests", "200", "--limit-request-line", "8190", "--timeout", "120", "chatbotdataingestor:app"
    ]

    # Start LLM services without SSL
    subprocess.Popen(llm_command, cwd=llm_working_directory)
    subprocess.Popen(dataingestor_command, cwd=llm_working_directory)

    # Chatbotquery (query process) with SSL
    certfile_query, keyfile_query = get_ssl_paths('query')
    chatbot_command = [
        "gunicorn", "--certfile", certfile_query, "--keyfile", keyfile_query,
        "-w", "4", "-b", "0.0.0.0:6002", 
        "--max-requests", "200", "--limit-request-line", "8190", "--timeout", "120", "chatbotquery:app"
    ]
    subprocess.Popen(chatbot_command, cwd=llm_working_directory)

    # Discord bot (run using python3)
    bot_working_directory = os.path.join(desired_directory, "api", "bot")
    discordbot_command = ["python3", "bot-cli.py"]
    subprocess.Popen(discordbot_command, cwd=bot_working_directory)

    # User services working directory
    user_working_directory = os.path.join(desired_directory, "api", "user")

    # Upgrade-webhook with SSL
    certfile_webhook, keyfile_webhook = get_ssl_paths('webhook')
    upgrade_webhook_command = [
        "gunicorn", "--certfile", certfile_webhook, "--keyfile", keyfile_webhook,
        "-w", "4", "-b", "0.0.0.0:7000", 
        "--max-requests", "200", "--limit-request-line", "8190", "--timeout", "120", "upgrade-webhook:app"
    ]
    subprocess.Popen(upgrade_webhook_command, cwd=user_working_directory)

    # Upgrade process with SSL
    certfile_upgrade, keyfile_upgrade = get_ssl_paths('upgrade')
    upgrade_command = [
        "gunicorn", "--certfile", certfile_upgrade, "--keyfile", keyfile_upgrade,
        "-w", "4", "-b", "0.0.0.0:7002", 
        "--max-requests", "200", "--limit-request-line", "8190", "--timeout", "120", "upgrade:app"
    ]
    subprocess.Popen(upgrade_command, cwd=user_working_directory)

    print("LLM services are running...")
