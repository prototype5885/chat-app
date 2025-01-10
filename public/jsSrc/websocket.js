const REJECTION_MESSAGE = 0

const ADD_CHAT_MESSAGE = 1
const CHAT_HISTORY = 2
const DELETE_CHAT_MESSAGE = 3

const ADD_SERVER = 21
const UPDATE_SERVER_PIC = 22
const DELETE_SERVER = 23
const SERVER_INVITE_LINK = 24
const UPDATE_SERVER_DATA = 25

const ADD_CHANNEL = 31
const CHANNEL_LIST = 32
const DELETE_CHANNEL = 33

const ADD_SERVER_MEMBER = 41
const SERVER_MEMBER_LIST = 42
const DELETE_SERVER_MEMBER = 43
const UPDATE_MEMBER_DISPLAY_NAME = 44
const UPDATE_MEMBER_PROFILE_PIC = 45

const UPDATE_STATUS = 53
const UPDATE_ONLINE = 55

const ADD_FRIEND = 61
const BLOCK_USER = 62
const UNFRIEND = 63

const INITIAL_USER_DATA = 241
const IMAGE_HOST_ADDRESS = 242
const UPDATE_USER_DATA = 243
const UPDATE_USER_PROFILE_PIC = 244

let wsClient
let wsConnected = false
let reconnectAttempts = 0


async function websocketConnected() {
    console.log("Refreshing websocket connections")

    removePlaceholderServers()

    // waits until server sends user's own ID and display name
    console.log("Waiting for server to send initial data...")
    await waitUntilBoolIsTrue(() => receivedInitialUserData)
    console.log("Initial data has already arrived")


    // request http address of image hosting server
    requestImageHostAddress()

    // wait until the address is received
    console.log("Waiting for server to send image host address..")
    await waitUntilBoolIsTrue(() => receivedImageHostAddress)
    console.log("Image host address has already arrived")

    registerHoverListeners() // add event listeners for hovering

    fadeOutLoading()
    const lastServer = getLastServer()
    if (lastServer === null) {
        selectServer("2000")
    } else {
        selectServer(getLastServer())
    }

}

function websocketBeforeConnected() {
    currentServerID = 0
    currentChannelID = 0
    lastChannelID = 0

    receivedInitialUserData = false
    receivedImageHostAddress = false

    removeServers()
    createPlaceHolderServers()
}

async function connectToWebsocket() {
    console.log("Connecting to websocket...")

    websocketBeforeConnected()

    // check if protocol is http or https
    const protocol = location.protocol === "https:" ? "wss://" : "ws://";
    const endpoint = `${protocol}${window.location.host}/ws`;
    wsClient = new WebSocket(endpoint);

    // make the websocket work with byte arrays
    wsClient.binaryType = "arraybuffer"

    wsClient.onopen = async function (_event) {
        console.log("Connected to WebSocket successfully.")
        wsConnected = true
        // if (currentChannelID != null) {

        websocketConnected()
        // }
    }

    wsClient.onclose = async function (_event) {
        console.log("Connection lost to websocket")
        if (reconnectAttempts > 60) {
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
        let packetJson = String.fromCharCode.apply(null, receivedBytes.slice(5, endIndex))

        console.log("Received packet:", endIndex, packetType, packetJson)

        if (packetType !== REJECTION_MESSAGE) {
            packetJson = packetJson.replace(/([\[:])?(\d{16,})([,\}\]])/g, "$1\"$2\"$3");
        }

        json = JSON.parse(packetJson)
        console.log(json)

        switch (packetType) {
            case REJECTION_MESSAGE: // Server sent rejection message
                console.warn("Server response:", json.Reason)
                break
            case ADD_CHAT_MESSAGE: // Server sent a chat message
                await chatMessageReceived(json)
                break
            case CHAT_HISTORY: // Server sent the requested chat history
                await chatHistoryReceived(json)
                break
            case DELETE_CHAT_MESSAGE: // Server sent which message was deleted
                deleteChatMessage(json)
                break
            case ADD_SERVER: // Server responded to the add server request
                console.log("Add server request response arrived")
                addServer(json.ServerID, json.UserID, json.Name, imageHost + json.Picture, "server")
                selectServer(json.ServerID)
                break
            case UPDATE_SERVER_PIC: // Server sent that a chat server picture was updated
                setServerPicture(json.ServerID, json.Pic)
                break
            case DELETE_SERVER: // Server sent which server was deleted
                console.log(`Server ID [${json.ServerID}] has been deleted`)
                const serverID = json.ServerID
                deleteServer(serverID)
                removeServerFromLastChannels(serverID)
                if (serverID === currentServerID) {
                    selectServer("2000")
                }
                break
            case SERVER_INVITE_LINK: // Server sent the requested invite link to the chat server
                console.log("Requested invite link to the chat server arrived, adding to clipboard")
                const inviteLink = `${window.location.protocol}//${window.location.host}/invite/${json}`
                console.log(inviteLink)
                await navigator.clipboard.writeText(inviteLink)
                break
            case UPDATE_SERVER_DATA: // server sent about a server data being updated
                console.log(`Received updated data of server ID [${json.ServerID}]`)
                if (json.NewSN) {
                    setServerName(json.ServerID, json.Name)
                }
                break
            case ADD_CHANNEL: // Server responded to the add channel request
                console.log(`Adding new channel called [${json.Name}]`)
                addChannel(json.ChannelID, json.Name)
                break
            case CHANNEL_LIST: // Server sent the requested channel list
                console.log("Requested channel list arrived")
                if (json.length === 0) {
                    console.warn("No channels on server ID", currentServerID)
                    break
                }
                for (let i = 0; i < json.length; i++) {
                    addChannel(json[i].ChannelID, json[i].Name)
                }
                selectLastChannels(json[0].ChannelID)
                break
            case ADD_SERVER_MEMBER: // A user connected to the server
                console.log("A user connected to the server")
                if (json.UserID !== ownUserID) {
                    addMember(json.UserID, json.Name, json.Picture, json.Status, json.StatusText)
                }
                break
            case SERVER_MEMBER_LIST: // Server sent the requested member list
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
            case DELETE_SERVER_MEMBER: // a member left the server
                if (json.UserID === ownUserID) {
                    console.log(`Left server ID [${json.ServerID}], deleting it from list`)
                    deleteServer(json.ServerID)
                    selectServer("2000")
                } else {
                    console.log(`User ID [${json.UserID}] left server ID [${json.ServerID}]`)
                    removeMember(json.UserID)
                }
                break
            case UPDATE_MEMBER_DISPLAY_NAME: // a member changed their display name
                setMemberDisplayName(json.UserID, json.DisplayName)
                break
            case UPDATE_MEMBER_PROFILE_PIC: // a member changed their profile pic
                setMemberProfilePic(json.UserID, json.Pic)
                break
            case UPDATE_USER_DATA: // replied to user data change
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
            case UPDATE_STATUS: // Server sent that a user changed their status value
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
            case UPDATE_ONLINE: // Server sent that someone went on or offline
                if (json.UserID === ownUserID) {

                } else {
                    setMemberOnline(json.UserID, json.Online)
                }
                break
            case ADD_FRIEND:
                if (json.UserID === ownUserID) {
                    ownFriends.push(json.ReceiverID)
                    console.log(`You have added user ID [${json.ReceiverID}] as friend`)
                } else if (json.ReceiverID == ownUserID) {
                    ownFriends.push(json.UserID)
                    console.log(`User ID [${json.UserID}] has added you as a friend`)
                }
                break
            case BLOCK_USER:
                break
            case UNFRIEND:
                if (json.UserID === ownUserID) {
                    removeFriend(json.ReceiverID)
                    console.log(`You have unfriended user ID [${json.ReceiverID}]`)
                } else if (json.ReceiverID == ownUserID) {
                    removeFriend(json.UserID)
                    console.log(`User ID [${json.UserID}] has unfriended you`)
                }
                break
            case INITIAL_USER_DATA: // Server sent the client's own user ID and display name
                setOwnUserID(json.UserID)
                setOwnProfilePic(json.ProfilePic)
                setOwnDisplayName(json.DisplayName)
                setOwnPronouns(json.Pronouns)
                setOwnStatusText(json.StatusText)
                setOwnFriends(json.Friends)
                setBlockedUsers(json.Blocks)

                if (json.Servers.length !== 0) {
                    for (let i = 0; i < json.Servers.length; i++) {
                        console.log("Adding server ID", json.Servers[i].ServerID)
                        addServer(json.Servers[i].ServerID, json.Servers[i].UserID, json.Servers[i].Name, imageHost + json.Servers[i].Picture, "server")
                    }
                } else {
                    console.log("Not being in any servers")
                }
                lookForDeletedServersInLastChannels()

                receivedInitialUserData = true
                console.log("Received own initial data")
                break
            case IMAGE_HOST_ADDRESS: // Server sent image host address
                if (json === "") {
                    console.log("Received image host address, server did not set any external")
                } else {
                    console.log("Received image host address:", json)
                }
                imageHost = json
                receivedImageHostAddress = true
                break
            case UPDATE_USER_PROFILE_PIC: // replied to profile pic change
                setOwnProfilePic(json.Pic)
                break
            default:
                console.log("Server sent unknown message type")
        }
    }
    await waitUntilBoolIsTrue(() => wsConnected)
}

async function preparePacket(type, struct) {
    // wait if websocket is not on yet
    await waitUntilBoolIsTrue(() => wsConnected)

    // convert the type value into a single byte value that will be the packet type
    const typeByte = new Uint8Array([1])
    typeByte[0] = type

    let json = JSON.stringify(struct)

    // workaround to turn uint64 values in json from string to integer type so server can process
    // numbers longer than 16 characters
    // json = json.replace(/"(\d{16,})"/g, "$1");
    json = json.replace(/(?<!\"Message\"\s*:\s*)\"(\d{16,})\"/g, "$1");

    console.log("Json to prepare for sending:", json)


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
    await preparePacket(ADD_CHAT_MESSAGE, {
        ChannelID: channelID,
        Message: message,
        AttTok: attachmentToken
    })
}
async function requestChatHistory(channelID, lastMessageID) {
    console.log("Requesting chat history for channel ID", channelID)
    preparePacket(CHAT_HISTORY, {
        ChannelID: channelID,
        FromMessageID: lastMessageID,
        Older: true // if true it will request older, if false it will request newer messages from the message id
    })
}
async function requestDeleteChatMessage(messageID) {
    console.log("Requesting to delete chat message ID", messageID)
    preparePacket(DELETE_CHAT_MESSAGE, {
        MessageID: messageID
    })
}
async function requestAddServer(serverName) {
    console.log("Requesting to add a new server")
    preparePacket(ADD_SERVER, {
        Name: serverName
    })
}

function requestRenameServer(serverID) {
    console.log("Requesting to rename server ID:", serverID)
}

// function requestServerList() {
//     console.log("Requesting server list")
//     preparePacket(SERVER_LIST, null)
// }

function requestDeleteServer(serverID) {
    if (document.getElementById(serverID).getAttribute("owned") == "false") return
    console.log("Requesting to delete server ID:", serverID)
    preparePacket(DELETE_SERVER, {
        ServerID: serverID
    })
}

function requestInviteLink(serverID) {
    if (document.getElementById(serverID).getAttribute("owned") == "false") return
    console.log("Requesting invite link creation for server ID:", serverID)
    preparePacket(SERVER_INVITE_LINK, {
        ServerID: serverID,
        SingleUse: false,
        Expiration: 7
    })
}

function requestAddChannel() {
    if (document.getElementById(currentServerID).getAttribute("owned") == "false") return
    console.log("Requesting to add new channel to server ID:", currentServerID)
    preparePacket(ADD_CHANNEL, {
        Name: "Channel",
        ServerID: currentServerID
    })
}

function requestChannelList() {
    console.log("Requesting channel list for current server ID", currentServerID)
    preparePacket(CHANNEL_LIST, {
        ServerID: currentServerID
    })
}

function requestMemberList() {
    console.log("Requesting member list for current server ID", currentServerID)
    preparePacket(SERVER_MEMBER_LIST, {
        ServerID: currentServerID
    })
}

function requestLeaveServer(serverID) {
    console.log("Requesting to leave a server ID", serverID)
    preparePacket(DELETE_SERVER_MEMBER, {
        ServerID: serverID
    })
}

function requestStatusChange(newStatus) {
    console.log("Requesting to change status")
    preparePacket(UPDATE_STATUS, {
        Status: newStatus
    })
}

function requestAddFriend(userID) {
    if (userID === ownUserID) {
        console.warn("You can't be friends with yourself")
        return
    }
    console.log(`Requesting to add user ID [${userID}] as friend`)
    preparePacket(ADD_FRIEND, {
        UserID: userID
    })
}

function requestBlockUser(userID) {
    if (userID === ownUserID) {
        console.warn("You can't block yourself")
        return
    }
    console.log(`Requesting to block user ID [${userID}]`)
    preparePacket(BLOCK_USER, {
        UserID: userID
    })
}

function requestUnfriend(userID) {
    if (userID === ownUserID) {
        console.warn("You can't unfriend yourself")
        return
    }
    console.log(`Requesting to unfriend user ID [${userID}]`)
    preparePacket(UNFRIEND, {
        UserID: userID
    })
}

function requestImageHostAddress() {
    console.log("Requesting image host address")
    preparePacket(IMAGE_HOST_ADDRESS, {})
}

function requestUpdateUserData(updatedUserData) {
    console.log("Requesting to update account data")
    preparePacket(UPDATE_USER_DATA, updatedUserData)
}

function requestUpdateServerData(updatedServerData) {
    console.log(`Requesting to update data of server ID [${updatedServerData.ServerID}]`)
    preparePacket(UPDATE_SERVER_DATA, updatedServerData)
}