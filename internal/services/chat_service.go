package services

import (
	"fmt"

	"github.com/krishcpatel/my-llm-project/internal/models"
	"github.com/krishcpatel/my-llm-project/internal/repositories"
)

// RetrieveRelevantContext fetches relevant messages via PGVector.
func RetrieveRelevantContext(conversationID int, query string, topN int) (string, error) {
	// 1. Compute the embedding
	queryEmbedding, err := ComputeEmbedding(query)
	if err != nil {
		return "", fmt.Errorf("ComputeEmbedding: %w", err)
	}
	embeddingStr := FormatVector(queryEmbedding)

	// 2. Retrieve relevant messages from repository
	msgs, err := repositories.GetRelevantMessages(models.GetDBConn(), conversationID, embeddingStr, topN)
	if err != nil {
		return "", fmt.Errorf("GetRelevantMessages: %w", err)
	}

	// 3. Build a context string
	context := ""
	for _, msg := range msgs {
		context += fmt.Sprintf("%s: %s\n", msg.Role, msg.Content)
	}
	return context, nil
}

// BuildRAGPrompt wraps the retrieved context plus new user query.
func BuildRAGPrompt(conversationID int, newQuery string, topN int) (string, error) {
	context, err := RetrieveRelevantContext(conversationID, newQuery, topN)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Context:\n%s\nUser: %s\nAssistant:", context, newQuery), nil
}

// BuildFullPrompt just transforms the entire conversation (if you want).
func BuildFullPrompt(msgs []models.Message) string {
	prompt := ""
	for _, m := range msgs {
		prompt += fmt.Sprintf("%s: %s\n", m.Role, m.Content)
	}
	return prompt + "\nAssistant:"
}

// BuildCombinedPrompt returns a prompt that includes:
// 1) The last 'recentLimit' messages of the conversation (short-term memory).
// 2) Top 'ragLimit' relevant older messages (PGVector) for the new user query.
// Then appends the new user query at the end.
func BuildCombinedPrompt(conversationID int, userQuery string, recentLimit, ragLimit int) (string, error) {
	dbConn := models.GetDBConn()

	// 1. Get the last N (short-term memory)
	shortTermMsgs, err := repositories.GetLastMessages(dbConn, conversationID, recentLimit)
	if err != nil {
		return "", fmt.Errorf("failed to fetch last messages: %w", err)
	}

	// 2. Retrieve relevant context (RAG) from the entire conversation
	relevantContext, err := RetrieveRelevantContext(conversationID, userQuery, ragLimit)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve RAG context: %w", err)
	}

	// 3. Build a short-term memory prompt from the last N messages
	shortTermPrompt := BuildFullPrompt(shortTermMsgs)

	// 4. Combine everything into one text:
	//    a) short-term conversation memory
	//    b) "Context" from RAG
	//    c) "User: {query}\nAssistant:"
	combined := fmt.Sprintf(
		`You are a chat bot that answers questions based on the conversation history and relevant context.
Short-Term Memory (This is your memory the previous chats we have had):
%s

---

Relevant Context (This is the relevant context found from embeddings):
%s

---

Main Query (This what you need to respond to):

User: %s
Assistant:`,
		shortTermPrompt,
		relevantContext,
		userQuery,
	)

	return combined, nil
}
