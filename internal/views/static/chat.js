// Conversation array to store messages (client-side memory)
let conversation = [];

// Get DOM elements
const chatArea = document.getElementById('chatArea');
const userInput = document.getElementById('userInput');
const sendBtn = document.getElementById('sendBtn');
const conversationIdInput = document.getElementById('conversationId');
const loadConversationBtn = document.getElementById('loadConversationBtn');
const newConversationBtn = document.getElementById('newConversationBtn');
const scrollBtn = document.getElementById('scrollBtn');

// When "Load Conversation" is clicked, reload the page with that conversation_id
loadConversationBtn.addEventListener('click', () => {
  const convId = conversationIdInput.value.trim() || "1";
  window.location.href = `/chat?conversation_id=${convId}`;
});

// When "New Conversation" is clicked, call the new conversation endpoint
newConversationBtn.addEventListener('click', () => {
  fetch('/chat/create')
    .then(response => response.json())
    .then(data => {
      // Update conversationId input with new conversation id and reload the page.
      conversationIdInput.value = data.conversation_id;
      window.location.href = `/chat?conversation_id=${data.conversation_id}`;
    })
    .catch(err => console.error("Error creating new conversation:", err));
});

// Send button: add new message and start SSE request
sendBtn.addEventListener('click', () => {
  const text = userInput.value.trim();
  if (!text) return;
  userInput.value = "";

  // Append user message to conversation array and UI
  conversation.push({ role: "user", content: text });
  appendMessage("user", text);

  // Start SSE request with entire conversation as JSON.
  startChatRequest(conversation);
});

function startChatRequest(conv) {
  // Encode the conversation as JSON and URL-encode it.
  const convJSON = encodeURIComponent(JSON.stringify(conv));
  const conversationId = conversationIdInput.value.trim() || "1";
  const url = `/chat/stream?conversation_id=${conversationId}&conv=${convJSON}`;
  const eventSource = new EventSource(url);

  let botMessageBuffer = "";
  eventSource.onmessage = (event) => {
    botMessageBuffer += event.data;
    updateBotMessage(botMessageBuffer);
  };

  eventSource.addEventListener("done", () => {
    console.log("Done event received. Closing SSE.");
    eventSource.close();
    conversation.push({ role: "assistant", content: botMessageBuffer });
  });

  eventSource.onerror = (err) => {
    console.error("SSE Error:", err);
    eventSource.close();
  };
}

function appendMessage(role, text) {
  const div = document.createElement("div");
  div.className = role === "user" ? "userMsg" : "botMsg";
  if (role === "user") {
    div.textContent = "User: " + text;
  } else {
    div.innerHTML = "Bot: " + text;
  }
  chatArea.appendChild(div);
  scrollChatAreaToBottom();
}

function updateBotMessage(text) {
  const lastChild = chatArea.lastElementChild;
  if (!lastChild || lastChild.className === "userMsg") {
    const div = document.createElement("div");
    div.className = "botMsg";
    div.innerHTML = "Bot: " + text;
    chatArea.appendChild(div);
  } else {
    lastChild.innerHTML = "Bot: " + text;
  }
  scrollChatAreaToBottom();
}

function scrollChatAreaToBottom() {
  chatArea.scrollTop = chatArea.scrollHeight;
}

// Scroll-to-bottom button event
scrollBtn.addEventListener('click', () => {
  scrollChatAreaToBottom();
});
