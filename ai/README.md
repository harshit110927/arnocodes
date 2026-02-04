# AI Service (Python)

## Description
AI service with placeholder RAG (Retrieval-Augmented Generation) implementation.

## Structure
- `src/main.py` - Flask application with RAG service endpoints

## Setup

### Virtual Environment
```bash
# Create virtual environment
python3 -m venv venv

# Activate virtual environment
source venv/bin/activate  # On Linux/Mac
# OR
venv\Scripts\activate  # On Windows

# Install dependencies
pip install -r requirements.txt
```

## Running Locally

```bash
# Make sure virtual environment is activated
source venv/bin/activate

# Run the service
python src/main.py
```

## Environment Variables
Copy `.env.example` to `.env` and update the values:
- `PORT` - Service port (default: 5000)
- `OPENAI_API_KEY` - OpenAI API key for LLM integration
- `ENVIRONMENT` - Environment name (development/production)

## API Endpoints
- `GET /health` - Health check endpoint
- `POST /query` - Query RAG service with a question
- `POST /index` - Index a document into the vector database

## Future Enhancements
This is a placeholder implementation. Future improvements include:
- Actual vector database integration (ChromaDB, Pinecone, etc.)
- LLM integration (OpenAI, Anthropic, etc.)
- Document processing and chunking
- Embedding generation
- Retrieval logic
