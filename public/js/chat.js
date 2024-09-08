
const wsClient = new WebSocket('wss://' + window.location.host + "/wss");
wsClient.binaryType = 'arraybuffer';
const messages = document.getElementById('chat-message-list')

wsClient.onopen = function (_event) {
    console.log('Connected to WebSocket successfully.');
};

// when server sends a message
wsClient.onmessage = function (event) {
    const receivedBytes = new Uint8Array(event.data);

    const packetType = Number(receivedBytes[0])
    const packetJson = new TextDecoder('utf-8').decode(receivedBytes.slice(1));

    console.log('Packet type:', packetType)
    console.log('Packet json:', packetJson)

    switch (packetType) {
        case 1: // ServerChatMsg
            addChatMessage(packetJson)
            messages.scrollTo({
                top: messages.scrollHeight,
                behavior: 'smooth'
            })
            break
    }
}

function sendPacket(typeStr, json) {
    if (wsClient.readyState === WebSocket.OPEN) {
        let typeByte
        switch (typeStr) {
            case "clientChatMsg":
                typeByte = new Uint8Array([1])
                break
        }
        let jsonBytes = new TextEncoder().encode(json)
        let packet = new Uint8Array(typeByte.length + jsonBytes.length)
        packet.set(typeByte) // adds the 1 byte type in the beginning
        packet.set(jsonBytes, typeByte.length) // adds the json binary after the type
        console.log('sending packet')
        wsClient.send(packet);
    }
    else {
        console.log("Websocket is not open")
    }
}

// socket.on('user connected', () => {
//     const messageData = {
//         userID: 0,
//         msgID: 0,
//         msg: 'New user has joined'
//     }

//     addChatMessage(messageData)
// })

// function main() {
//     // for (i = 0; i < 20; i++) {
//     // socket.emit('chat message', 'test')
//     // addMember(5)
//     // addFriend(5)
//     // }
// }

// main()