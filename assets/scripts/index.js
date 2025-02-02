const ws = new WebSocket("ws://localhost:8080/ws");

ws.onopen = function() {
    console.log("Connected to WebSocket");
};

ws.onmessage = function(event) {
    console.log("Message received:", event.data);
    let chatBox = document.getElementById("chat-box");
    let messageDiv = document.createElement("div");
    messageDiv.textContent = event.data;
    chatBox.appendChild(messageDiv);
};

ws.onerror = function(error) {
    console.error("WebSocket error:", error);
};

ws.onclose = function() {
    console.warn("WebSocket closed");
};

document.getElementById("chat-form").addEventListener("submit", function(event) {
    event.preventDefault();
    let input = document.getElementById("message");
    if (ws.readyState === WebSocket.OPEN) {
        ws.send(input.value);
        console.log("Message sent:", input.value);
    } else {
        console.warn("WebSocket is not open");
    }
    input.value = "";
});