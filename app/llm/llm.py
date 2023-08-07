from flask import Flask, request, jsonify
import openai
import os
import urllib.parse

app = Flask(__name__)

# Set your OpenAI API key here or use an environment variable
OPENAI_API_KEY = "sk-"
openai.api_key = OPENAI_API_KEY

openai.Engine.list()
@app.route('/llm', methods=['GET'])
def generate_response():
    prompt = request.args.get('prompt')

    if not prompt:
        return jsonify({'error': 'Missing prompt parameter'}), 400

    decoded_prompt = urllib.parse.unquote(prompt)  # Decode URL-encoded prompt
    print(decoded_prompt)
    try:
        response = openai.Completion.create(
            engine="text-davinci-001",
            prompt=decoded_prompt,
            max_tokens=225  # Adjust as needed
        )

        generated_text = response.choices[0].text.strip()
        print(generated_text)

        return jsonify({'response': generated_text})

    except openai.error.OpenAIError as e:
        print(f"OpenAI error: {e}")
        return jsonify({'error': str(e)}), 500
    except Exception as e:
        print(f"Error: {e}")
        return jsonify({'error': 'Failed to generate response'}), 500

if __name__ == '__main__':
    app.run(debug=True)
