const wsClient = new WebSocket('ws://' + window.location.host + "/ws")
wsClient.binaryType = 'arraybuffer'

var ownUserID

if (typeof (Storage) !== "undefined") {
    console.log('Supports storage')
} else {
    console.log('Doesnt support storage')
}

document.addEventListener("DOMContentLoaded", function () {
    addServer(2000, 0, 'Direct Messages', 'hs.svg', 'dm') // add the direct messages button
    createAddServerButton() // add event listener for the add server button

    // add place holder servers depending on how many servers the client was in, will delete on websocket connection
    // purely visual
    const placeholderButtons = createPlaceHolderServers()

    // this will continue when websocket connected
    wsClient.onopen = function (_event) {
        console.log('Connected to WebSocket successfully.')

        const loading = document.getElementById('loading')

        const fadeOut = 0.25 //seconds

        setTimeout(() => {
            loading.remove(); // Remove the element from the DOM
        }, fadeOut * 1000)

        loading.style.transition = `background-color ${fadeOut}s ease`
        loading.style.backgroundColor = '#00000000'
        loading.style.pointerEvents = 'none'

        // remove placeholder servers
        for (let i = 0; i < placeholderButtons.length; i++) {
            console.log(placeholderButtons[i])
            placeholderButtons[i].remove()
        }

        requestServerList()

        // for (i = 0; i < 1000000; i++) {
        //     sendChatMessage(Math.random().toString(), BigInt(1810996904781152256n))
        // }
    }
})



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
        case 0:
            console.warn(json.Reason)
            break
        case 1: // server sent a chat message
            addChatMessage(BigInt(json.IDm), BigInt(json.IDu), json.Msg)
            messages.scrollTo({
                top: messages.scrollHeight,
                behavior: 'smooth'
            })
            break
        case 2: // server sent the requested chat history
            if (json === null) {
                console.log('Chat history is empty')
                return
            }
            for (let i = 0; i < json.length; i++) {
                addChatMessage(BigInt(json[i].IDm), BigInt(json[i].IDu), json[i].Msg) // messageID, userID, Message
            }
            messages.scrollTo({
                top: messages.scrollHeight,
                behavior: 'instant'
            })
            break
        case 3: // server sent which message was deleted
            const messageID = BigInt(json.MessageID)
            console.log('Deleting message id ' + messageID)
            document.getElementById(messageID).remove()
            break
        case 21: // server responded to the add server request
            addServer(BigInt(json.ServerID), BigInt(json.OwnerID), json.Name, json.Picture, 'server', discordGray, discordBlue)
            localStorage.setItem('serverCount', parseInt(localStorage.getItem('serverCount')) + 1)
            break
        case 22: // server sent the requested server list
            if (json == null) { console.log('Not being in any servers'); break }
            for (let i = 0; i < json.length; i++) {
                console.log('Adding server ID', json[i].ServerID)
                addServer(BigInt(json[i].ServerID), BigInt(json[i].OwnerID), json[i].Name, json[i].Picture, 'server', discordGray, discordBlue)
            }
            localStorage.setItem('serverCount', json.length.toString())
            lookForDeletedServersInLastChannels()
            // go to the last server that was selected last time
            // const lastServerID = localStorage.getItem('lastServer')
            // if (lastServerID != null) {
            //     selectServer(BigInt(lastServerID))
            // }
            break
        case 23: // server sent which server was deleted
            const serverID = BigInt(json.ServerID)
            deleteServer(serverID)
            deleteServerFromLastChannels(serverID)
            // removeDeletedLastChannels()
            localStorage.setItem('serverCount', parseInt(localStorage.getItem('serverCount')) - 1)
            if (serverID == currentServerID) {
                selectServer(2000)
            }
            break
        case 31: // server responded to the add channel request
            addChannel(BigInt(json.ChannelID), json.Name)
            break
        case 32: // server sent the requested channel list
            if (json == null) {
                console.warn(`No channels on server ID`)
                break
            }
            for (let i = 0; i < json.length; i++) {
                addChannel(BigInt(json[i].ChannelID), json[i].Name)
            }
            selectLastChannels(BigInt(json[0].ChannelID))
            break
        case 241: // server sent the client's own user ID
            ownUserID = BigInt(json)
            console.log('Received own user ID:', ownUserID)
            break
        default:
            console.log('Server sent unknown message type')
    }
}

function preparePacket(type, bigintID, struct) {
    if (wsClient.readyState === WebSocket.OPEN) {
        // convert the type value into a single byte value that will be the packet type
        const typeByte = new Uint8Array([1])
        typeByte[0] = type

        let json = JSON.stringify(struct)

        // workaround to turn uint64 value in json from string to normal number value
        // since javascript cant serialize BigInt
        if (bigintID != 0) {
            json = json.replace(`"${bigintID}"`, bigintID)
        }

        // serialize the struct into json then convert to byte array
        let jsonBytes
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
    preparePacket(1, channelID, {
        ChannelID: channelID.toString(),
        Message: message
    })
}
function requestChatHistory(channelID) {
    preparePacket(2, channelID, {
        ChannelID: channelID.toString()
    })
}
function requestChatMessageDeletion(messageID) {
    preparePacket(3, messageID, {
        MessageID: messageID.toString()
    })
}
function requestAddServer(serverName) {
    preparePacket(21, 0, {
        Name: serverName
    })
}

function requestRenameServer(serverID) {
    console.log('Requesting to rename server ID:', serverID)
}

function requestDeleteServer(serverID) {
    console.log('Requesting to delete server ID:', serverID)
    preparePacket(23, serverID, {
        ServerID: serverID.toString()
    })
}

function requestServerList() {
    preparePacket(22, 0, null)
}

function requestAddChannel() {
    preparePacket(31, currentServerID, {
        Name: 'Channel',
        ServerID: currentServerID.toString()
    })
}

function requestChannelList() {
    preparePacket(32, currentServerID, {
        ServerID: currentServerID.toString()
    })
}