# My LLM Project

A Go-based web application that integrates **Ollama** (local LLM inference) with **PostgreSQL + PGVector** for vector storage and advanced AI functionality. The project uses an MVC-like architecture and features a web-based chat interface that supports multi-turn conversations and a basic retrieval-augmented generation (RAG) system to handle long chats.

---

## Features

1. **Local LLM (Ollama) Integration**  
   - Streams token-by-token responses via Server-Sent Events (SSE).  
   - Uses a local Ollama server at `http://localhost:11411/api/generate` for chat generation.

2. **Embeddings via Ollama**  
   - Computes text embeddings using an Ollama endpoint (e.g., `http://localhost:11434/api/embed`) with a configurable embedding model.  
   - Embeddings are stored in PostgreSQL using the PGVector extension for efficient similarity searches.

3. **PostgreSQL + PGVector**  
   - Stores chat messages (with computed embeddings) and conversation metadata.
   - Uses PGVector similarity (via the `<->` operator) to retrieve context for long conversations.

4. **Retrieval-Augmented Generation (RAG)**  
   - When a new message arrives, the system computes its embedding using the model defined in `EMBEDDING_MODEL`.
   - The RAG engine retrieves the top-N most relevant past messages using PGVector.
   - The retrieved context is combined with the new user query to build a prompt for the LLM.
   - If the RAG retrieval fails, the full conversation context is used as a fallback.

5. **Multi-Conversation Support**  
   - Users can create a new conversation using the `/chat/create` endpoint.
   - The chat page displays past messages for the selected conversation.
   - A visible conversation ID input and buttons allow the user to load or create a conversation.

6. **Configurable Models via Environment Variables**  
   - `CHAT_MODEL`: Specifies the model used for chat generation (default: `"deepseek-r1:14b"`).
   - `EMBEDDING_MODEL`: Specifies the model used for generating embeddings (default: `"all-minilm"`).
   - `EMBEDDING_DIM`: Specifies the expected embedding dimension (e.g., `5120`).

7. **User Interface Enhancements**  
   - The chat area has a fixed height and is scrollable, while the rest of the page remains static.
   - A "Scroll to Bottom" button lets users quickly jump to the latest messages.
   - The conversation selector is visible, allowing users to change or create conversations.

---

## Requirements

- **Go** 1.18+  
- **PostgreSQL** 14+ with the [PGVector extension](https://github.com/pgvector/pgvector) installed  
- **Ollama** installed locally (or via Docker)  
- (Optional) Docker for containerized development

---

## Quick Start

1. **Clone the Repository**

   ```bash
   git clone https://github.com/<your-username>/my-llm-project.git
   cd my-llm-project
   ```
2. Install Dependencies
   ```bash
   go mod tidy
   ```
3. Configure Environment Variables
   ### Create a .env file or set environment variables directly:
   ```bash
   # Database configuration
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=your_db_user
   DB_PASSWORD=your_db_password
   DB_NAME=your_db_name

   # Model configuration
   CHAT_MODEL=deepseek-r1:14b
   EMBEDDING_MODEL=all-minilm
   EMBEDDING_DIM=5120
   ```
4. Set Up Your Database
   ### Ensure PostgreSQL is running and create the necessary tables. For example:
   ```bash
   -- Enable PGVector extension
   CREATE EXTENSION IF NOT EXISTS vector;

   -- Create conversations table
   DROP TABLE IF EXISTS conversations;
   CREATE TABLE conversations (
       id SERIAL PRIMARY KEY,
       user_id INT,         -- Can be NULL for now
       created_at TIMESTAMP DEFAULT NOW()
   );
   
   -- Create chat_messages table
   DROP TABLE IF EXISTS chat_messages;
   CREATE TABLE chat_messages (
       id SERIAL PRIMARY KEY,
       conversation_id INT REFERENCES conversations(id),
       role TEXT NOT NULL,
       content TEXT NOT NULL,
       embedding vector(5120),  -- Adjust the dimension based on EMBEDDING_DIM
       created_at TIMESTAMP DEFAULT NOW()
   );
   ```
5. Run the Application
   ### Build and run your application:
   ```bash
   go run ./cmd/server
   ```
   ### Or build a binary:
   ```bash
   go build -o my-llm-app ./cmd/server
   ./my-llm-app
   ```
   ### The server will start on http://localhost:4000.
6. Interact with the Chat
   - Navigate to http://localhost:4000/chat in your browser. 
   - Use the conversation selector to load or create a conversation.
   - Use the "Scroll to Bottom" button to jump to the latest messages.

---

## How It Works

## Conversation & Message Storage

### Conversations:
- Conversations are stored in the `conversations` table.
- A new conversation can be created via the `/chat/create` endpoint.

### Messages:
- Each message (user or assistant) is stored in the `chat_messages` table along with its embedding.
- The embedding is computed via the Ollama embedding API using the model specified by the `EMBEDDING_MODEL` environment variable.

## RAG Engine Integration

### Embedding Computation:
- When a new user message is received, the system computes its embedding by calling the embedding endpoint.
- The computed vector is then formatted to a string literal suitable for PGVector.

### Context Retrieval:
- The RAG engine uses this embedding to retrieve the top-N most similar past messages from the same conversation (using PGVector's similarity operator).
- The number of messages to retrieve (N) is configurable via the chat UI.

### Prompt Building:
- The retrieved messages are combined with the new user query to build a RAG prompt.
- This prompt is then sent to the LLM for generating a response.

### Fallback:
- If the RAG retrieval fails, the system falls back to using the full conversation context.

## Live Chat via SSE

- The chat interface uses Server-Sent Events (SSE) to stream partial responses in real time.
- Keep-alive pings are used to maintain the connection.
- The chat area is scrollable with a fixed height, and a "Scroll to Bottom" button is provided.

## Model Configuration

- `CHAT_MODEL` is used for chat generation.
- `EMBEDDING_MODEL` is used for computing embeddings.
