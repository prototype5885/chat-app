let wsClient
let wsConnected
let reconnectAttempts = 0

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
        if (reconnectAttempts > 10) {
            console.log("Failed reconnecting to the server")
            setLoadingText("Failed reconnecting")
            return
        }
        console.log("Reconnection attempt:", reconnectAttempts)
        reconnectAttempts++

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

        console.log("Received packet:", endIndex, packetType, packetJson)
        console.log(`Received packet size: [${receivedBytes.length} bytes] index: [${endIndex}] packetType: [${packetType}] json: ${packetJson}`)

        const json = fixJson(packetJson)
        // const json = JSON.parse(packetJson)
        switch (packetType) {
            case 0: // Server sent rejection message
                console.warn(json.Reason)
                break
            case 1: // Server sent a chat message
                await chatMessageReceived(json)
                break
            case 2: // Server sent the requested chat history
                await chatHistoryReceived(json)
                break
            case 3: // Server sent which message was deleted
                deleteChatMessage(json)
                break
            case 21: // Server responded to the add server request
                console.log("Add server request response arrived")
                addServer(json.ServerID, json.UserID, json.Name, imageHost + json.Picture, "server")
                selectServer(json.ServerID)
                break
            case 22: // Server sent the requested server list
                console.log("Requested server list arrived")
                if (json != null) {
                    for (let i = 0; i < json.length; i++) {
                        console.log("Adding server ID", json[i].ServerID)
                        addServer(json[i].ServerID, json[i].UserID, json[i].Name, imageHost + json[i].Picture, "server")
                    }
                } else {
                    console.log("Not being in any servers")
                }
                lookForDeletedServersInLastChannels()
                break
            case 23: // Server sent which server was deleted
                console.log(`Server ID [${json.ServerID}] has been deleted`)
                const serverID = json.ServerID
                deleteServer(serverID)
                removeServerFromLastChannels(serverID)
                if (serverID === currentServerID) {
                    selectServer("2000")
                }
                break
            case 24: // Server sent the requested invite link to the chat server
                console.log("Requested invite link to the chat server arrived, adding to clipboard")
                const inviteLink = `${window.location.protocol}//${window.location.host}/invite/${json}`
                console.log(inviteLink)
                await navigator.clipboard.writeText(inviteLink)
                break
            case 31: // Server responded to the add channel request
                console.log(`Adding new channel called [${json.Name}]`)
                addChannel(json.ChannelID, json.Name)
                break
            case 32: // Server sent the requested channel list
                console.log("Requested channel list arrived")
                if (json == null) {
                    console.warn("No channels on server ID", currentServerID)
                    break
                }
                for (let i = 0; i < json.length; i++) {
                    addChannel(json[i].ChannelID, json[i].Name)
                }
                selectLastChannels(json[0].ChannelID)
                break
            case 41: // A user connected to the server
                console.log("A user connected to the server")
                if (json.UserID !== ownUserID) {
                    addMember(json.UserID, json.Name, json.Picture, json.Status, json.StatusText)
                }
                break
            case 42: // Server sent the requested member list
                console.log("Requested member list arrived")
                if (json == null) {
                    console.warn("No members on server ID", currentServerID)
                    break
                }
                for (let i = 0; i < json.length; i++) {
                    addMember(json[i].UserID, json[i].Name, json[i].Pic, json[i].Online, json[i].Status, json[i].StatusText)
                }
                memberListLoaded = true
                break
            case 43: // a member left the server
                if (json.UserID === ownUserID) {
                    console.log(`Left server ID [${json.ServerID}], deleting it from list`)
                    deleteServer(json.ServerID)
                    selectServer("2000")
                } else {
                    console.log(`User ID [${json.UserID}] left server ID [${json.ServerID}]`)
                    removeMember(json.UserID)
                }
                break
            case 44: // a member changed their display name
                setMemberDisplayName(json.UserID, json.DisplayName)
                break
            case 45: // a member changed their profile pic
                setMemberProfilePic(json.UserID, json.Pic)
                break
            case 243: // replied to user data change
                if (json.NewDN) {
                    setOwnDisplayName(json.DisplayName)
                }
                if (json.NewP) {
                    setOwnPronouns(json.Pronouns)
                }
                if (json.NewST) {
                    setOwnStatusText(json.StatusText)
                }
                break
            case 244: // replied to profile pic change
                setOwnProfilePic(json.Pic)
                break
            case 53: // Server sent that a user changed their status value
                if (json.UserID === ownUserID) {
                    console.log("My new status:", json.Status)
                } else {
                    console.log(`User ID [${json.UserID}] changed their status to [${json.Status}]`)
                }
                changeStatusValueInMemberList(json.UserID, json.Status)
                break
            // case 54: // Server sent that a user changed their status text
            //     if (json.UserID === ownUserID) {
            //         console.log("My new status text:", json.StatusText)
            //         setUserPanelStatusText(json.StatusText)
            //     } else {
            //         console.log(`User ID [${json.UserID}] changed their status text to [${json.StatusText}]`)
            //     }
            //     setMemberOnlineStatusText(json.UserID, json.StatusText)
            //     break
            case 55: // Server sent that someone went on or offline
                if (json.UserID === ownUserID) {

                } else {
                    setMemberOnline(json.UserID, json.Online)
                }
                break
            case 56: // Server sent new pronouns
                setOwnPronouns(json.Pronouns)
            case 241: // Server sent the client's own user ID and display name
                ownUserID = json.UserID
                setOwnProfilePic(json.ProfilePic)
                setOwnDisplayName(json.DisplayName)
                setOwnPronouns(json.Pronouns)
                setOwnStatusText(json.StatusText)
                receivedInitialUserData = true
                console.log(`Received own user ID [${ownUserID}] and display name: [${ownDisplayName}]:`)
                break
            case 242: // Server sent image host address
                if (json === "") {
                    console.log("Received image host address, server did not set any external")
                } else {
                    console.log("Received image host address:", json)
                }
                imageHost = json
                receivedImageHostAddress = true
                break
            default:
                console.log("Server sent unknown message type")
        }
    }
    await waitUntilBoolIsTrue(() => wsConnected)
}

// class ReceivedChatMessage {
//     constructor(messageID, userID, message) {
//         this.messageID = messageID;
//         this.userID = userID;
//         this.message = message;
//     }
//
//     static fromJSON(jsonString) {
//         const data = JSON.parse(jsonString);
//         return new ReceivedChatMessage(data.IDm, data.IDu, this.Msg);
//     }
// }

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

async function sendChatMessage(message, channelID, attachmentToken) { // type is 1
    console.log("Sending a chat message")
    await preparePacket(1, [channelID], {
        ChannelID: channelID,
        Message: message,
        AttTok: attachmentToken
    })
}
async function requestChatHistory(channelID, lastMessageID) {
    console.log("Requesting chat history for channel ID", channelID)
    preparePacket(2, [channelID, lastMessageID], {
        ChannelID: channelID,
        FromMessageID: lastMessageID,
        Older: true // if true it will request older, if false it will request newer messages from the message id
    })
}
async function requestDeleteChatMessage(messageID) {
    console.log("Requesting to delete chat message ID", messageID)
    preparePacket(3, [messageID], {
        MessageID: messageID
    })
}
async function requestAddServer(serverName) {
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

function requestStatusChange(newStatus) {
    console.log("Requesting to change status")
    preparePacket(53, [], {
        Status: newStatus
    })
}

function requestImageHostAddress() {
    console.log("Requesting image host address")
    preparePacket(242, [], {})
}

function requestUpdateUserData(updatedUserData) {
    console.log("Requesting to update account data")
    preparePacket(243, [], updatedUserData)
}


