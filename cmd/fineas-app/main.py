import subprocess
import os

if __name__ == '__main__':
    # Specify the directory where you want the subprocess to run
    current_directory = os.getcwd()
    desired_directory = os.path.abspath(os.path.join(current_directory, os.pardir, os.pardir))
    working_directory = os.path.join(desired_directory, "api", "llm")
    
    # Specify the command to run the Flask server
    llm = ["python3", "llm.py"]
    dataingestor = ["python3", "chatbotdataingestor.py"]
    chatbot = ["python3", "chatbotquery.py"]

    # Use subprocess to run the Flask server with the specified working directory
    subprocess.Popen(llm, cwd=working_directory)
    subprocess.Popen(dataingestor, cwd=working_directory)
    subprocess.Popen(chatbot, cwd=working_directory)


    print("LLM services are running...")