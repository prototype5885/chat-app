let wsClient
let wsConnected

function refreshWebsocketContent() {
    document.querySelectorAll('.server').forEach(server => {
        server.remove();
    })

    requestServerList()
    selectServer("2000")
    fadeOutLoading()
}

async function connectToWebsocket() {
    console.log("Connecting to websocket...")
    // check if protocol is http or https
    // if (location.protocol === "https:") {
    //     wsClient = new WebSocket("wss://" + window.location.host + "/wss")
    // } else {
    //     wsClient = new WebSocket("ws://" + window.location.host + "/ws")
    // }

    // check if protocol is http or https
    const protocol = location.protocol === "https:" ? "wss://" : "ws://";
    const endpoint = `${protocol}${window.location.host}/ws`;
    wsClient = new WebSocket(endpoint);

    // make the websocket work with byte arrays
    wsClient.binaryType = "arraybuffer"

    wsClient.onopen = async function (_event) {
        console.log("Connected to WebSocket successfully.")
        wsConnected = true
        if (currentChannelID != null) {
            currentServerID = 0
            currentChannelID = 0
            lastChannelID = 0
            refreshWebsocketContent()
        }
    }

    wsClient.onclose = async function (_event) {
        console.log("Connection lost to websocket")
        wsConnected = false
        fadeInLoading()
        await connectToWebsocket()
    }

    // wsClient.onerror = async function (_event) {
        // console.log("Error in websocket")
        // wsConnected = false
        // await reconnectToWebsocket()
    // }

    // when server sends a message
    wsClient.onmessage = async function (event) {
        let receivedBytes = new Uint8Array(event.data)

        // convert the first 4 bytes into uint32 to get the endIndex,
        // which marks the end of the packet
        const reversedBytes = receivedBytes.slice(0, 4).reverse()
        const endIndex = new DataView(reversedBytes.buffer).getUint32(0)

        // 5th byte is a 1 byte number which states the type of the packet
        const packetType = receivedBytes[4]

        // get the json string from the 6th byte to the end
        const packetJson = String.fromCharCode.apply(null, receivedBytes.slice(5, endIndex))

        // console.log("Received packet:", endIndex, packetType, packetJson)
        console.log(`Received packet size: [${receivedBytes.length} bytes] index: [${endIndex}] packetType: [${packetType}] json: ${packetJson}`)

        const msg = JSON.parse(packetJson)
        switch (packetType) {
            case 0: // Server sent rejection message
                console.warn(msg.Reason)
                break
            case 1: // Server sent a chat message
                await chatMessageReceived(msg)
                break
            case 2: // Server sent the requested chat history
                await chatHistoryReceived(msg)
                break
            case 3: // Server sent which message was deleted
                deleteChatMessage(msg)
                break
            case 21: // Server responded to the add server request
                console.log("Add server request response arrived")
                addServer(msg.ServerID, msg.OwnerID, msg.Name, imageHost + "content/avatars/" + msg.Picture, "server")
                selectServer(msg.ServerID)
                break
            case 22: // Server sent the requested server list
                console.log("Requested server list arrived")
                if (msg != null) {
                    for (let i = 0; i < msg.length; i++) {
                        console.log("Adding server ID", msg[i].ServerID)
                        addServer(msg[i].ServerID, msg[i].OwnerID, msg[i].Name, imageHost + "content/avatars/" + msg[i].Picture, "server")
                    }
                } else {
                    console.log("Not being in any servers")
                }
                lookForDeletedServersInLastChannels()
                break
            case 23: // Server sent which server was deleted
                console.log(`Server ID [${msg.ServerID}] has been deleted`)
                const serverID = msg.ServerID
                deleteServer(serverID)
                removeServerFromLastChannels(serverID)
                if (serverID === currentServerID) {
                    selectServer("2000")
                }
                break
            case 24: // Server sent the requested invite link to the chat server
                console.log("Requested invite link to the chat server arrived, adding to clipboard")
                const inviteLink = `${window.location.protocol}//${window.location.host}/invite/${msg}`
                console.log(inviteLink)
                await navigator.clipboard.writeText(inviteLink)
                break
            case 31: // Server responded to the add channel request
                console.log(`Adding new channel called [${msg.Name}]`)
                addChannel(msg.ChannelID, msg.Name)
                break
            case 32: // Server sent the requested channel list
                console.log("Requested channel list arrived")
                if (msg == null) {
                    console.warn("No channels on server ID", currentServerID)
                    break
                }
                for (let i = 0; i < msg.length; i++) {
                    addChannel(msg[i].ChannelID, msg[i].Name)
                }
                selectLastChannels(msg[0].ChannelID)
                break
            case 42: // Server sent the requested member list
                console.log("Requested member list arrived")
                if (msg == null) {
                    console.warn("No members on server ID", currentServerID)
                    break
                }
                for (let i = 0; i < msg.length; i++) {
                    addMember(msg[i].UserID, msg[i].Name, msg[i].Picture, msg[i].Status)
                }
                memberListLoaded = true
                break
            case 43: // Server sent user which user left a server
                if (msg.UserID === ownUserID) {
                    console.log(`Left server ID [${msg.ServerID}], deleting it from list`)
                    deleteServer(msg.ServerID)
                    selectServer("2000")
                } else {
                    console.log(`User ID [${msg.UserID}] left server ID [${msg.ServerID}]`)
                    removeMember(msg.UserID)
                }
                break
            case 51: // Server sent that a user changed display name
                if (userID === ownUserID) {
                    console.log("New display name:", msg.newName)
                } else {
                    console.log(`User ID [${msg.UserID}] changed their name to [${msg.NewName}]`)
                }
                changeDisplayNameInChatMessageList(userID, newDisplayName)
                changeDisplayNameInMemberList(userID, newDisplayName)
                break

            case 241: // Server sent the client's own user ID
                ownUserID = msg
                console.log("Received own user ID:", ownUserID)
                UserPanelName.textContent = ownUserID
                receivedOwnUserID = true
                break
            case 242: // Server sent image host address
                if (msg === "") {
                    console.log("Received image host address, server did not set any external")
                } else {
                    console.log("Received image host address:", msg)
                }
                imageHost = msg
                receivedImageHostAddress = true
                break
            default:
                console.log("Server sent unknown message type")
        }
    }
    await waitUntilBoolIsTrue(() => wsConnected)
}

class ReceivedChatMessage {
    constructor(messageID, userID, message) {
        this.messageID = messageID;
        this.userID = userID;
        this.message = message;
    }

    static fromJSON(jsonString) {
        const data = JSON.parse(jsonString);
        return new ReceivedChatMessage(data.IDm, data.IDu, this.Msg);
    }
}

async function preparePacket(type, bigIntIDs, struct) {
    await waitUntilBoolIsTrue(() => wsConnected)

    // convert the type value into a single byte value that will be the packet type
    const typeByte = new Uint8Array([1])
    typeByte[0] = type

    let json = JSON.stringify(struct)

    // workaround to turn uint64 value in json from string to normal number value
    // since javascript cant serialize BigInt
    for (i = 0; i < bigIntIDs.length; i++) {
        if (bigIntIDs[i] !== 0) {
            json = json.replace(`"${bigIntIDs[i]}"`, bigIntIDs[i])
        }
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

    console.log("Prepared packet:", endIndex, packet[4], json)

    wsClient.send(packet)
}

function sendChatMessage(message, channelID) { // type is 1
    console.log("Sending a chat message")
    preparePacket(1, [channelID], {
        ChannelID: channelID,
        Message: message
    })
}
function requestChatHistory(channelID, lastMessageID) {
    console.log("Requesting chat history for channel ID", channelID)
    preparePacket(2, [channelID, lastMessageID], {
        ChannelID: channelID,
        FromMessageID: lastMessageID,
        Older: true // if true it will request older, if false it will request newer messages from the message id
    })
}
function requestDeleteChatMessage(messageID) {
    console.log("Requesting to delete chat message ID", messageID)
    preparePacket(3, [messageID], {
        MessageID: messageID
    })
}
function requestAddServer(serverName) {
    console.log("Requesting to add a new server")
    preparePacket(21, [0], {
        Name: serverName
    })
}

function requestRenameServer(serverID) {
    console.log("Requesting to rename server ID:", serverID)
}

function requestDeleteServer(serverID) {
    if (document.getElementById(serverID).getAttribute("owned") == "false") return
    console.log("Requesting to delete server ID:", serverID)
    preparePacket(23, [serverID], {
        ServerID: serverID
    })
}

function requestInviteLink(serverID) {
    if (document.getElementById(serverID).getAttribute("owned") == "false") return
    console.log("Requesting invite link creation for server ID:", serverID)
    preparePacket(24, [serverID], {
        ServerID: serverID,
        SingleUse: false,
        Expiration: 7
    })
}

function requestServerList() {
    console.log("Requesting server list")
    preparePacket(22, [0], null)
}

function requestAddChannel() {
    if (document.getElementById(currentServerID).getAttribute("owned") == "false") return
    console.log("Requesting to add new channel to server ID:", currentServerID)
    preparePacket(31, [currentServerID], {
        Name: "Channel",
        ServerID: currentServerID
    })
}

function requestChannelList() {
    console.log("Requesting channel list for current server ID", currentServerID)
    preparePacket(32, [currentServerID], {
        ServerID: currentServerID
    })
}

function requestMemberList() {
    console.log("Requesting member list for current server ID", currentServerID)
    preparePacket(42, [currentServerID], {
        ServerID: currentServerID
    })
}

function requestLeaveServer(serverID) {
    console.log("Requesting to leave a server ID", serverID)
    preparePacket(43, [serverID], {
        ServerID: serverID
    })
}

function requestChangeDisplayName(newName) {
    console.log("Requesting to change display name to:", newName)
    preparePacket(51, [], {
        NewName: newName
    })
}

function requestImageHostAddress() {
    console.log("Requesting image host address")
    preparePacket(242, [], {})
}