
const messages = document.getElementById('chat-message-list')

const wsClient = new WebSocket('wss://' + window.location.host + "/wss");
wsClient.binaryType = 'arraybuffer';

wsClient.onopen = function (_event) {
    console.log('Connected to WebSocket successfully.');
    requestChatHistory('2002')
};

// when server sends a message
wsClient.onmessage = function (event) {
    const receivedBytes = new Uint8Array(event.data);

    // convert the first 4 bytes into uint32 to get the endIndex,
    // which marks the end of the packet
    const reversedBytes = receivedBytes.slice(0, 4).reverse()
    const endIndex = new DataView(reversedBytes.buffer).getUint32(0)

    // 5th byte is a 1 byte number which states the type of the packet
    const packetType = receivedBytes[4]

    // get the json string from the 6th byte to the end
    const packetJson = String.fromCharCode.apply(null, receivedBytes.slice(5, endIndex));

    console.log('Received packet:', endIndex, packetType, packetJson)

    switch (packetType) {
        case 1: // server sent a chat message
            const msg = JSON.parse(packetJson)
            addChatMessage(BigInt(msg.MessageID), BigInt(msg.ChannelID), BigInt(msg.UserID), msg.Username, msg.Message)
            messages.scrollTo({
                top: messages.scrollHeight,
                behavior: 'smooth'
            })
            break
        case 2: // server sent the requested chat history
            const history = JSON.parse(packetJson)
            if (history.Messages === null) {
                console.log('Chat history is empty')
                return
            }
            for (let i = 0; i < history.Messages.length; i++) {
                addChatMessage(BigInt(history.Messages[i].MessageID), BigInt(history.Messages[i].ChannelID), BigInt(history.Messages[i].UserID), history.Messages[i].Username, history.Messages[i].Message)
            }
            messages.scrollTo({
                top: messages.scrollHeight,
                behavior: 'instant'
            })
            break
        case 3: // server sent which message was deleted
            const messageToDelete = JSON.parse(packetJson)
            deleteChatMessage(messageToDelete.MessageID)
            break
        case 21: // server sent information of newly added chat server
            const server = JSON.parse(packetJson)
            addServer(BigInt(server.ServerID), BigInt(server.ServerOwnerID), server.ServerName)
            break
    }
}

function preparePacket(type, struct) {
    if (wsClient.readyState === WebSocket.OPEN) {
        // convert the type value into a single byte value that will be the packet type
        const typeByte = new Uint8Array([1])
        typeByte[0] = type

        // serialize the struct into json then convert to byte array
        const json = JSON.stringify(struct)
        const jsonBytes = new TextEncoder().encode(json)

        // convert the end index uint32 value into 4 bytes
        const endIndex = jsonBytes.length + 5
        const buffer = new ArrayBuffer(4)
        new DataView(buffer).setUint32(0, endIndex, true)
        const endIndexBytes = new Uint8Array(buffer)

        // merge them into a single packet
        const packet = new Uint8Array(4 + 1 + jsonBytes.length)
        packet.set(endIndexBytes, 0) // first 4 bytes will be the length
        packet.set(typeByte, 4) // 5. byte will be the packet type
        packet.set(jsonBytes, 5) // rest will be the json byte array

        console.log('Prepared packet:', endIndex, packet[4], json)

        wsClient.send(packet);
    }
    else {
        console.log("Websocket is not open")
    }
}

function sendChatMessage(message, channelID) { // type is 1
    preparePacket(1, {
        ChannelID: channelID.toString(),
        Message: message
    })
}
function requestChatHistory(channelID) {
    preparePacket(2, {
        ChannelID: channelID
    })
}
function requestChatMessageDeletion(messageID) {
    preparePacket(3, {
        MessageID: messageID.toString()
    })
}
function requestAddServer(serverName) {
    preparePacket(21, {
        Name: serverName
    })
}