import os

from dotenv import load_dotenv
from flask import Flask, jsonify, request

from errors import AIServiceError, ProviderError
from providers.gemini import GeminiProvider
from service import AIOrchestrator

load_dotenv()

app = Flask(__name__)
orchestrator = AIOrchestrator(provider=GeminiProvider())


@app.route('/health', methods=['GET'])
def health():
    return jsonify({"status": "healthy", "service": "ai-service"})


@app.route('/query', methods=['POST'])
def query():
    data = request.get_json(silent=True) or {}

    user_id = data.get('userId')
    mode = data.get('mode', 'chatbot')
    text = data.get('text') or data.get('query')

    result = orchestrator.process(user_id=user_id, mode=mode, text=text)
    return jsonify(result)


@app.route('/index', methods=['POST'])
def index_document():
    # Deliberately left as no-op placeholder; integration should be implemented by
    # consumers outside this module if indexing is required.
    return jsonify({
        "status": "not_implemented",
        "message": "Document indexing is not part of this AI mode service."
    }), 501


@app.errorhandler(AIServiceError)
def handle_ai_error(err: AIServiceError):
    return jsonify({
        "error": {
            "code": err.code,
            "message": err.message,
        }
    }), err.status_code


@app.errorhandler(Exception)
def handle_unexpected_error(_err: Exception):
    fallback = ProviderError("Internal AI processing failure.")
    return jsonify({
        "error": {
            "code": fallback.code,
            "message": fallback.message,
        }
    }), 500


if __name__ == '__main__':
    port = int(os.getenv('PORT', 5000))
    debug = os.getenv('ENVIRONMENT', 'development') == 'development'
    app.run(host='0.0.0.0', port=port, debug=debug)
