/**
 * Opens the chat window and starts a WebSocket connection with the given user.
 * @param {number} userID - The ID of the user to chat with.
 */
function openChat(userID) {
    document.getElementById("chatWindow").style.display = "block";
    startWebSocket(userID);
}

/** @type {WebSocket | null} */
let ws = null;

/**
 * Initializes a WebSocket connection for chatting.
 * @param {number} userID - The ID of the user to chat with.
 */
function startWebSocket(userID) {
    ws = new WebSocket(`ws://localhost:8080/chat/${userID}`);

    ws.onmessage = async function(event) {
        let data = JSON.parse(event.data);
        let chatMessages = document.getElementById("chatMessages");
        let msgDiv = document.createElement("div");

        let senderUsername = await getUsernameById(data.sender_id);

        msgDiv.innerHTML = (data.sender_id === userID)
            ? `<b>${senderUsername}</b>: ${data.content}`
            : `<b>You</b>: ${data.content}`;
        
        msgDiv.classList.add(data.sender_id === userID ? "receiver" : "sender");
        chatMessages.appendChild(msgDiv);
    };
}

/**
 * Sends a chat message to the specified receiver.
 * @param {number} receiverID - The ID of the recipient user.
 */
async function sendMessage(receiverID) {
    let chatInput = document.getElementById("chatInput");
    let message = chatInput.value;
    let currentUserID = await getCurrentUserID();

    if (!ws || ws.readyState !== WebSocket.OPEN) {
        console.error("WebSocket is not connected.");
        return;
    }

    ws.send(JSON.stringify({
        sender_id: currentUserID,
        receiver_id: receiverID,
        content: message
    }));

    chatInput.value = ""; 
}
