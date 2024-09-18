const wsClient = new WebSocket('wss://' + window.location.host + "/wss")
wsClient.binaryType = 'arraybuffer'

wsClient.onopen = function (_event) {
    console.log('Connected to WebSocket successfully.')
    requestServerList()
    requestChannelList(1917)
    requestChatHistory(2002)
}

// when server sends a message
wsClient.onmessage = function (event) {
    const receivedBytes = new Uint8Array(event.data)

    // convert the first 4 bytes into uint32 to get the endIndex,
    // which marks the end of the packet
    const reversedBytes = receivedBytes.slice(0, 4).reverse()
    const endIndex = new DataView(reversedBytes.buffer).getUint32(0)

    // 5th byte is a 1 byte number which states the type of the packet
    const packetType = receivedBytes[4]

    // get the json string from the 6th byte to the end
    const packetJson = String.fromCharCode.apply(null, receivedBytes.slice(5, endIndex))

    console.log('Received packet:', endIndex, packetType, packetJson)

    const messages = document.getElementById('chat-message-list')
    const json = JSON.parse(packetJson)
    switch (packetType) {
        case 1: // server sent a chat message
            addChatMessage(BigInt(json.MessageID), BigInt(json.ChannelID), BigInt(json.UserID), json.Username, json.Message)
            messages.scrollTo({
                top: messages.scrollHeight,
                behavior: 'smooth'
            })
            break
        case 2: // server sent the requested chat history
            if (json.Messages === null) {
                console.log('Chat history is empty')
                return
            }
            for (let i = 0; i < json.Messages.length; i++) {
                addChatMessage(BigInt(json.Messages[i].MessageID), BigInt(json.Messages[i].ChannelID), BigInt(json.Messages[i].UserID), json.Messages[i].Username, json.Messages[i].Message)
            }
            messages.scrollTo({
                top: messages.scrollHeight,
                behavior: 'instant'
            })
            break
        case 3: // server sent which message was deleted
            deleteChatMessage(json.MessageID)
            break
        case 21: // server responded to the add server request
            addServer(BigInt(json.ServerID), json.Name, json.Picture)
            break
        case 22: // server sent the requested server list
            if (json.Servers == null) {
                console.log('Not being in any servers')
                break
            }
            for (let i = 0; i < json.Servers.length; i++) {
                addServer(BigInt(json.Servers[i].ServerID), json.Servers[i].Name, json.Servers[i].Picture)
            }
            break
        case 31: // server responded to the add channel request
            addChannel(json.ChannelID, json.Name)
            break
        case 32: // server sent the requested channel list
            if (json.Channels == null) {
                console.log(`No channels on server ID`)
                break
            }
            for (let i = 0; i < json.Channels.length; i++) {
                // addChannel
                addChannel(BigInt(json.Channels[i].ChannelID), json.Channels[i].Name)
            }
            break
    }
}

function preparePacket(type, struct) {
    if (wsClient.readyState === WebSocket.OPEN) {
        const json = JSON.stringify(struct)
        // convert the type value into a single byte value that will be the packet type
        const typeByte = new Uint8Array([1])
        typeByte[0] = type

        // serialize the struct into json then convert to byte array
        var jsonBytes
        if (struct != null) {
            jsonBytes = new TextEncoder().encode(json)
        } else {
            jsonBytes = new Uint8Array([0])
        }

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

        wsClient.send(packet)
    }
    else {
        console.log('Websocket is not ope')
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
        ChannelID: channelID.toString()
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

function requestServerList() {
    preparePacket(22, null)
}

function requestAddChannel() {
    const id = 1917
    preparePacket(31, {
        Name: 'Test Channel',
        ServerID: id.toString()
    })
}

function requestChannelList(serverID) {
    preparePacket(32, {
        ServerID: serverID.toString()
    })
}