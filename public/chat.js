var wsClient
if (location.protocol === 'https:') {
    wsClient = new WebSocket('wss://' + window.location.host + "/wss")
} else {
    wsClient = new WebSocket('ws://' + window.location.host + "/ws")
}
wsClient.binaryType = 'arraybuffer'

var ownUserID
var receivedOwnUserID = false

if (typeof (Storage) !== "undefined") {
    console.log('Supports storage')
} else {
    console.log('Doesnt support storage')
}

function waitUntilBoolIsTrue(checkFunction, interval = 10) {
    return new Promise((resolve) => {
        const intervalId = setInterval(() => {
            if (checkFunction()) {
                clearInterval(intervalId);
                resolve();
            }
        }, interval);
    });
}

// this runs after webpage was loaded
document.addEventListener("DOMContentLoaded", function () {
    addServer(2000, 0, 'Direct Messages', 'hs.svg', 'dm') // add the direct messages button

    // add place holder servers depending on how many servers the client was in, will delete on websocket connection
    // purely visual
    const placeholderButtons = createPlaceHolderServers()
    serversSeparatorVisibility()
    console.log("Placeholder buttons:", placeholderButtons.length)

    // this will continue when websocket connected
    wsClient.onopen = async function (_event) {
        console.log('Connected to WebSocket successfully.')

        // waits until server sends user's own ID
        console.log("Waiting for server to send own user ID...")
        await waitUntilBoolIsTrue(() => receivedOwnUserID);

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
            placeholderButtons[i].remove()
        }

        requestServerList()

        // for (i = 0; i < 1000000; i++) {
        //     sendChatMessage(Math.random().toString(), BigInt(1810996904781152256n))
        // }
    }

    registerClickListeners() // add event listener for clicking
    registerHoverListeners() // add event listeners for hovering
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

    const json = JSON.parse(packetJson)
    switch (packetType) {
        case 0: // Server sent rejection message
            console.log("Server sent rejection message:", json.Reason)
            break
        case 1: // Server sent a chat message
            console.log("Server sent a chat message")
            addChatMessage(BigInt(json.IDm), BigInt(json.IDu), json.Msg)
            ChatMessagesList.scrollTo({
                top: ChatMessagesList.scrollHeight,
                behavior: 'smooth'
            })
            break
        case 2: // Server sent the requested chat history
            console.log("Server sent the requested chat history")
            if (json !== null) {
                for (let i = 0; i < json.length; i++) {
                    addChatMessage(BigInt(json[i].IDm), BigInt(json[i].IDu), json[i].Msg) // messageID, userID, Message
                }
                ChatMessagesList.scrollTo({
                    top: ChatMessagesList.scrollHeight,
                    behavior: 'instant'
                })
            } else {
                console.log('Current channel has no chat history')
            }
            break
        case 3: // Server sent which message was deleted
            console.log("Server sent which message was deleted")
            const messageID = BigInt(json.MessageID)
            console.log('Deleting message id ' + messageID)
            document.getElementById(messageID).remove()
            break
        case 21: // Server responded to the add server request
            console.log("Server responded to the add server request")
            addServer(BigInt(json.ServerID), BigInt(json.OwnerID), json.Name, json.Picture, 'server', discordGray, discordBlue)
            selectServer(BigInt(json.ServerID))
            serversSeparatorVisibility()
            break
        case 22: // Server sent the requested server list
            console.log("Server sent the requested server list")
            if (json != null) {
                for (let i = 0; i < json.length; i++) {
                    console.log('Adding server ID', json[i].ServerID)
                    addServer(BigInt(json[i].ServerID), BigInt(json[i].OwnerID), json[i].Name, json[i].Picture, 'server', discordGray, discordBlue)
                }
            } else {
                console.log('Not being in any servers')
            }

            serversSeparatorVisibility()
            lookForDeletedServersInLastChannels()
            break
        case 23: // Server sent which server was deleted
            console.log("Server sent which server was deleted")
            const serverID = BigInt(json.ServerID)
            deleteServer(serverID)
            deleteServerFromLastChannels(serverID)
            // removeDeletedLastChannels()
            if (serverID == currentServerID) {
                selectServer(2000)
            }
            serversSeparatorVisibility()
            break
        case 24: // Server sent the requested invite link to the chat server
            console.log("Server sent the requested invite link to the chat server")
            const inviteID = BigInt(json).toString()
            const inviteLink = `${window.location.protocol}//${window.location.host}/invite/${inviteID}`
            console.log(inviteLink)
            navigator.clipboard.writeText(inviteLink)
            break
        case 31: // Server responded to the add channel request
            console.log("Server responded to the add channel request")
            addChannel(BigInt(json.ChannelID), json.Name)
            break
        case 32: // Server sent the requested channel list
            console.log("Server sent the requested channel list")
            if (json == null) {
                console.warn("No channels on server ID", currentServerID)
                break
            }
            for (let i = 0; i < json.length; i++) {
                addChannel(BigInt(json[i].ChannelID), json[i].Name)
            }
            selectLastChannels(BigInt(json[0].ChannelID))
            break
        case 42: // Server sent the requested member list
            console.log("Server sent the requsted member list")
            if (json == null) {
                console.warn("No members on server ID", currentServerID)
                break
            }
            for (let i = 0; i < json.length; i++) {
                console.log(BigInt(json[i].UserID))
            }
            break
        case 241: // Server sent the client's own user ID
            ownUserID = BigInt(json)
            console.log('Received own user ID:', ownUserID)
            UserPanelName.textContent = ownUserID
            receivedOwnUserID = true
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
    console.log("Sending a chat message")
    preparePacket(1, channelID, {
        ChannelID: channelID.toString(),
        Message: message
    })
}
function requestChatHistory(channelID) {
    console.log("Requesting chat history for channel ID", channelID)
    preparePacket(2, channelID, {
        ChannelID: channelID.toString()
    })
}
function requestDeleteChatMessage(messageID) {
    console.log("Requesting to delete chat message ID", messageID)
    preparePacket(3, messageID, {
        MessageID: messageID.toString()
    })
}
function requestAddServer(serverName) {
    console.log("Requesting to add a new server")
    preparePacket(21, 0, {
        Name: serverName
    })
}

function requestRenameServer(serverID) {
    console.log('Requesting to rename server ID:', serverID)
}

function requestDeleteServer(serverID) {
    if (document.getElementById(serverID).getAttribute('owned') == 'false') return
    console.log('Requesting to delete server ID:', serverID)
    preparePacket(23, serverID, {
        ServerID: serverID.toString()
    })
}

function requestInviteLink(serverID) {
    if (document.getElementById(serverID).getAttribute('owned') == 'false') return
    console.log("Requesting invite link creation for server ID:", serverID)
    preparePacket(24, serverID, {
        ServerID: serverID.toString(),
        SingleUse: false,
        Expiration: 7
    })
}

function requestServerList() {
    console.log("Requesting server list")
    preparePacket(22, 0, null)
}

function requestAddChannel() {
    if (document.getElementById(currentServerID).getAttribute('owned') == 'false') return
    console.log('Requesting to add new channel to server ID:', currentServerID)
    preparePacket(31, currentServerID, {
        Name: 'Channel',
        ServerID: currentServerID.toString()
    })
}

function requestChannelList() {
    console.log("Requesting channel list for current server ID", currentServerID)
    preparePacket(32, currentServerID, {
        ServerID: currentServerID.toString()
    })
}

function requestMemberList() {
    console.log("Requesting member list for current server ID", currentServerID)
    preparePacket(42, currentServerID, {
        ServerID: currentServerID.toString()
    })
}