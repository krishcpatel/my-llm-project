# My LLM Project

A Go-based web application that integrates **Ollama** (local LLM inference) with **PostgreSQL + PGVector** for vector storage and advanced AI functionality. Built with an MVC-like architecture in mind and featuring a simple web-based chat interface.

---

## Features

1. **Local LLM (Ollama) Integration**  
   - Streams token-by-token responses via Server-Sent Events (SSE).  
   - Uses a local Ollama server on `http://localhost:11411/api/generate`.

2. **Postgres + PGVector**  
   - Optionally store embeddings or chat logs in a PostgreSQL database with the [PGVector Extension](https://github.com/pgvector/pgvector).

3. **MVC-Like Structure**  
   - Clear separation of concerns with `cmd/`, `internal/ai`, `internal/db`, and `internal/web`.

4. **Live Chat Interface**  
   - HTML/JS page at `/chat` streams real-time responses from your local LLM server.  
   - Keep-alive pings are used to prevent SSE timeouts.

---

## Project Layout

```
my-llm-project/ 
  ├── cmd/ │ 
    └── web/ │ 
      └── main.go // Entry point (web server) 
  ├── internal/ │ 
    ├── db/ │ 
    │ └── db.go // Database logic (PGVector) │ 
    ├── ai/ │ 
    │ └── ai.go // Ollama streaming integration │ 
    └── web/ │ 
      └── handlers.go // Handlers for SSE, chat, etc. 
  ├── go.mod 
  ├── go.sum 
  ├── README.md
```

Feel free to customize or extend the structure.

---

## Requirements

- **Go** 1.18+  
- **PostgreSQL** 14+ (with the [PGVector extension](https://github.com/pgvector/pgvector))  
- **Ollama** installed (or running via Docker)  

---

## Quick Start

1. **Clone this Repo**

   ```bash
   git clone https://github.com/<your-username>/my-llm-project.git
   cd my-llm-project
   ```
2. Install Dependencies

```bash
go mod tidy
```

3. Configure Postgres (if using PGVector)
  
    Ensure you have a running Postgres instance with PGVector installed.
    Set environment variables if needed (e.g., DB_HOST, DB_PORT, etc.) or edit db.go.

4. Run Ollama (if not already running)
```bash
ollama serve
```

  By default, listens on localhost:11411.

5. Run the Go App
```bash
go run ./cmd/web
```
or build a binary:
```bash
go build -o my-llm-app ./cmd/web
./my-llm-app
```
The server will start on http://localhost:4000. Go to http://localhost:4000/chat in your browser.

6. Chat

    Type a prompt and watch partial tokens stream in real-time.
    For longer sessions, you can adapt the code to store conversation history in memory or a database.

### Usage

  1. Navigate to http://localhost:4000/chat in your web browser.
  2. Enter a prompt, e.g. “Why is the sky blue?”
  3. Observe real-time streaming of the LLM’s response.
  4. Modify internal/web/handlers.go or internal/ai/ai.go to adjust how prompts and responses are handled.
