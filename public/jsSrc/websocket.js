class WebsocketClass {
    static wsClient
    static wsConnected = false
    static reconnectAttempts = 0

    static REJECTION_MESSAGE = 0

    static ADD_CHAT_MESSAGE = 1
    static CHAT_HISTORY = 2
    static DELETE_CHAT_MESSAGE = 3
    static STARTED_TYPING = 4
    static EDIT_CHAT_MESSAGE = 5

    static ADD_SERVER = 21
    static UPDATE_SERVER_PIC = 22
    static DELETE_SERVER = 23
    static SERVER_INVITE_LINK = 24
    static UPDATE_SERVER_DATA = 25

    static ADD_CHANNEL = 31
    static CHANNEL_LIST = 32
    static DELETE_CHANNEL = 33
    static UPDATE_CHANNEL_DATA = 34

    static ADD_SERVER_MEMBER = 41
    static SERVER_MEMBER_LIST = 42
    static DELETE_SERVER_MEMBER = 43
    static UPDATE_MEMBER_DATA = 44
    static UPDATE_MEMBER_PROFILE_PIC = 45

    static UPDATE_STATUS = 53
    static UPDATE_ONLINE = 55

    static ADD_FRIEND = 61
    static BLOCK_USER = 62
    static UNFRIEND = 63

    static OPEN_DM = 71
    static REQUEST_DM_LIST = 72
    static DM_CHAT_HISTORY = 73

    static INITIAL_USER_DATA = 241
    static IMAGE_HOST_ADDRESS = 242
    static UPDATE_USER_DATA = 243
    static UPDATE_USER_PROFILE_PIC = 244

    // static canSendPacket = true
    static timerStage = 0
    static lastSendAttempt

    static async websocketConnected() {
        console.log('Refreshing websocket connections')

        ServerListClass.removePlaceholderServers()

        // waits until server sends user's own ID and display name
        console.log('Waiting for server to send initial data...')
        await MainClass.waitUntilBoolIsTrue(() => MainClass.receivedInitialUserData)
        console.log('Initial data has already arrived')


        // request http address of image hosting server
        // await WebsocketClass.requestImageHostAddress()

        // wait until the address is received
        // console.log('Waiting for server to send image host address..')
        // await MainClass.waitUntilBoolIsTrue(() => main.receivedImageHostAddress)
        // console.log('Image host address has already arrived')

        LoadingClass.fadeOutLoading()
        const lastServer = LocalStorageClass.getLastServer()
        if (lastServer === null) {
            await ServerListClass.selectServer('1')
        } else {
            await ServerListClass.selectServer(lastServer)
        }
    }

    static websocketBeforeConnected() {
        MainClass.setCurrentServerID('0')
        MainClass.setCurrentChannelID('0')
        MainClass.lastChannelID = 0

        MainClass.receivedInitialUserData = false
        MainClass.receivedImageHostAddress = false

        ServerListClass.removeServers()
        ServerListClass.createPlaceHolderServers()
    }

    static async connectToWebsocket() {
        console.log('Connecting to websocket...')

        this.websocketBeforeConnected()

        // check if protocol is http or https
        const protocol = location.protocol === 'https:' ? 'wss://' : 'ws://'
        const endpoint = `${protocol}${window.location.host}/ws`
        WebsocketClass.wsClient = new WebSocket(endpoint)

        // make the websocket work with byte arrays
        WebsocketClass.wsClient.binaryType = 'arraybuffer'

        WebsocketClass.wsClient.onopen = async () => {
            console.log('Connected to WebSocket successfully.')
            WebsocketClass.wsConnected = true
            await this.websocketConnected()
        }

        WebsocketClass.wsClient.onclose = async () => {
            console.log('Connection lost to websocket')
            if (WebsocketClass.reconnectAttempts > 60) {
                console.log('Failed reconnecting to the server')
                LoadingClass.setLoadingText('Failed reconnecting')
                return
            }
            console.log('Reconnection attempt:', WebsocketClass.reconnectAttempts)
            WebsocketClass.reconnectAttempts++

            WebsocketClass.wsConnected = false
            LoadingClass.fadeInLoading()
            await this.connectToWebsocket()
        }

        // wsClient.onerror = async function () {
        // console.log('Error in websocket')
        // wsConnected = false
        // await reconnectToWebsocket()
        // }

        // when server sends a message
        WebsocketClass.wsClient.onmessage = async (event) => {
            let receivedBytes = new Uint8Array(event.data)

            // convert the first 4 bytes into uint32 to get the endIndex,
            // which marks the end of the packet
            const reversedBytes = receivedBytes.slice(0, 4).reverse()
            const endIndex = new DataView(reversedBytes.buffer).getUint32(0)

            // 5th byte is a 1 byte number which states the type of the packet
            const packetType = receivedBytes[4]

            // get the json string from the 6th byte to the end
            // let packetJson = String.fromCharCode.apply(null, receivedBytes.slice(5, endIndex))

            const decoder = new TextDecoder()
            let packetJson = decoder.decode(receivedBytes.slice(5, endIndex))

            console.log('Received packet:', endIndex, packetType, packetJson)

            if (packetType !== WebsocketClass.REJECTION_MESSAGE) {
                packetJson = packetJson.replace(/([\[:])?(\d{16,})([,\}\]])/g, "$1\"$2\"$3");
            }

            const json = JSON.parse(packetJson)
            if (json !== '') {
                console.log(json)
            }

            switch (packetType) {
                case WebsocketClass.REJECTION_MESSAGE: // Server sent rejection message
                    console.warn('Server response:', json.Reason)
                    break
                case WebsocketClass.ADD_CHAT_MESSAGE: // Server sent a chat message
                    await ChatMessageListClass.chatMessageReceived(json)
                    break
                case WebsocketClass.CHAT_HISTORY: // Server sent the requested chat history
                    await ChatMessageListClass.chatHistoryReceived(json)
                    break
                case WebsocketClass.DELETE_CHAT_MESSAGE: // Server sent which message was deleted
                    ChatMessageListClass.deleteChatMessage(json)
                    break
                case WebsocketClass.STARTED_TYPING: // Server sent that someone started typing on given channel
                    ChatMessageListClass.someoneStartedTyping(json.Typing, json.UserID, json.ChannelID)
                    break
                case WebsocketClass.EDIT_CHAT_MESSAGE: // Server sent info about an edited message
                    ChatMessageListClass.editChatMessage(json.MessageID, json.Message)
                    break
                case WebsocketClass.ADD_SERVER: // Server responded to the add server request
                    console.log('Add server request response arrived')
                    ServerListClass.addServer(json.ServerID, json.Owned, json.Name, MainClass.imageHost + json.Picture, 'server')
                    await ServerListClass.selectServer(json.ServerID)
                    break
                case WebsocketClass.UPDATE_SERVER_PIC: // Server sent that a chat server picture was updated
                    ServerListClass.setServerPicture(json.ServerID, json.Pic)
                    break
                case WebsocketClass.DELETE_SERVER: // Server sent which server was deleted
                    console.log(`Server ID [${json.ServerID}] has been deleted`)
                    const serverID = json.ServerID
                    ServerListClass.deleteServer(serverID)
                    LocalStorageClass.removeServerFromLastChannels(serverID)
                    if (serverID === MainClass.getCurrentServerID()) {
                        await ServerListClass.selectServer('1')
                    }
                    break
                case WebsocketClass.SERVER_INVITE_LINK: // Server sent the requested invite link to the chat server
                    console.log('Requested invite link to the chat server arrived, adding to clipboard')
                    const inviteLink = `${window.location.protocol}//${window.location.host}/invite/${json}`
                    console.log(inviteLink)
                    await navigator.clipboard.writeText(inviteLink)
                    break
                case WebsocketClass.UPDATE_SERVER_DATA: // server sent about a server data being updated
                    console.log(`Received updated data of server ID [${json.ServerID}]`)
                    if (json.NewSN) {
                        ServerListClass.setServerName(json.ServerID, json.Name)
                    }
                    if (json.NewSN) {
                        WindowManagerClass.setCurrentUpdateUserDataResponseLabel(true)
                    } else {
                        WindowManagerClass.setCurrentUpdateUserDataResponseLabel(false)
                    }
                    break
                case WebsocketClass.ADD_CHANNEL: // Server responded to the add channel request
                    console.log(`Adding new channel called [${json.Name}]`)
                    ChannelListClass.addChannel(json.ChannelID, json.Name)
                    break
                case WebsocketClass.CHANNEL_LIST: // Server sent the requested channel list
                    console.log('Requested channel list arrived')
                    // if no channels on server
                    if (json.length === 0) {
                        console.warn('No channels on server ID', MainClass.getCurrentServerID())
                        ThirdColumnMainClass.reset()
                        // await ChannelListClass.selectChannel('0')
                        break
                    }
                    // add the channels if there are
                    for (let i = 0; i < json.length; i++) {
                        await ChannelListClass.addChannel(json[i].ChannelID, json[i].Name)
                    }
                    // select the channel saved in local storage, or just select the first one
                    const lastChannelID = LocalStorageClass.selectLastChannel()
                    if (lastChannelID !== null) {
                        await ChannelListClass.selectChannel(lastChannelID)
                    } else {
                        await ChannelListClass.selectChannel(json[0].ChannelID)
                    }
                    break
                case WebsocketClass.DELETE_CHANNEL:
                    console.log(`Channel ID [${json.ChannelID}] has been removed]`)
                    await ChannelListClass.removeChannel(json.ChannelID)
                    break
                case WebsocketClass.UPDATE_CHANNEL_DATA:
                    if (json.NewCN) {
                        ChannelListClass.setChannelName(json.ChannelID, json.Name)
                    }
                    break
                case WebsocketClass.ADD_SERVER_MEMBER: // A user connected to the server
                    console.log('A user connected to the server')
                    if (json.ServerID === MainClass.getCurrentServerID()) {
                        MemberListClass.addMember(json.Data.UserID, json.Data.Name, json.Data.Pic, json.Data.Online, json.Data.Status, json.Data.StatusText)
                    } else {
                        console.warn(`Received that User ID [${json.Data.UserID}] connected to server ID [${json.ServerID}] but the current server ID is [${MainClass.getCurrentServerID()}]`)
                    }
                    break
                case WebsocketClass.SERVER_MEMBER_LIST: // Server sent the requested member list
                    console.log('Requested member list arrived')
                    if (json == null) {
                        console.warn('No members on server ID', MainClass.getCurrentServerID())
                        break
                    }
                    for (let i = 0; i < json.length; i++) {
                        MemberListClass.addMember(json[i].UserID, json[i].Name, json[i].Pic, json[i].Online, json[i].Status, json[i].StatusText)
                    }
                    MainClass.memberListLoaded = true
                    break
                case WebsocketClass.DELETE_SERVER_MEMBER: // a member left the server
                    if (json.UserID === MainClass.getOwnUserID()) {
                        console.log(`Left server ID [${json.ServerID}], deleting it from list`)
                        ServerListClass.deleteServer(json.ServerID)
                        await ServerListClass.selectServer('1')
                    } else {
                        console.log(`User ID [${json.UserID}] left server ID [${json.ServerID}]`)
                        MemberListClass.removeMember(json.UserID)
                    }
                    break
                case WebsocketClass.UPDATE_MEMBER_DATA: // a member changed user data
                    if (json.NewDN) {
                        MemberListClass.setMemberDisplayName(json.UserID, json.DisplayName)
                        ChatMessageListClass.changeDisplayNameInChatMessageList(json.UserID, json.DisplayName)
                    }
                    if (json.NewP) {
                        // TODO set pronouns
                    }
                    if (json.NewST) {
                        MemberListClass.setMemberStatusText(json.UserID, json.StatusText)
                    }
                    break
                case WebsocketClass.UPDATE_MEMBER_PROFILE_PIC: // a member changed their profile pic
                    json.Pic = MainClass.getAvatarFullPath(json.Pic)
                    MemberListClass.setMemberProfilePic(json.UserID, json.Pic)
                    ChatMessageListClass.setChatMessageProfilePic(json.UserID, json.Pic)
                    break
                case WebsocketClass.UPDATE_USER_DATA: // replied to user data change
                    if (json.NewDN) {
                        MainClass.setOwnDisplayName(json.DisplayName)
                    }
                    if (json.NewP) {
                        MainClass.setOwnPronouns(json.Pronouns)
                    }
                    if (json.NewST) {
                        MainClass.setOwnStatusText(json.StatusText)
                    }

                    if (json.NewDN || json.NewP || json.NewST) {
                        WindowManagerClass.setCurrentUpdateUserDataResponseLabel(true)
                    } else {
                        WindowManagerClass.setCurrentUpdateUserDataResponseLabel(false)
                    }

                    break
                case WebsocketClass.UPDATE_STATUS: // Server sent that a user changed their status value
                    if (json.UserID === MainClass.getOwnUserID()) {
                        console.log('My new status:', json.Status)
                    } else {
                        console.log(`User ID [${json.UserID}] changed their status to [${json.Status}]`)
                    }
                    MemberListClass.changeStatusValueInMemberList(json.UserID, json.Status)
                    break
                // case 54: // Server sent that a user changed their status text
                //     if (json.UserID === main.ownUserID) {
                //         console.log('My new status text:', json.StatusText)
                //         setUserPanelStatusText(json.StatusText)
                //     } else {
                //         console.log(`User ID [${json.UserID}] changed their status text to [${json.StatusText}]`)
                //     }
                //     setMemberOnlineStatusText(json.UserID, json.StatusText)
                //     break
                case WebsocketClass.UPDATE_ONLINE: // Server sent that someone went on or offline
                    // if (json.UserID === MainClass.getOwnUserID()) {
                    //
                    // } else {
                    MemberListClass.setMemberOnline(json.UserID, json.Online)
                    // }
                    break
                case WebsocketClass.ADD_FRIEND:
                    if (json.UserID === MainClass.getOwnUserID()) {
                        MainClass.myFriends.push(json.ReceiverID)
                        console.log(`You have added user ID [${json.ReceiverID}] as friend`)
                    } else if (json.ReceiverID === MainClass.getOwnUserID()) {
                        MainClass.myFriends.push(json.UserID)
                        console.log(`User ID [${json.UserID}] has added you as a friend`)
                    }
                    break
                case WebsocketClass.BLOCK_USER:
                    break
                case WebsocketClass.UNFRIEND:
                    if (json.UserID === MainClass.getOwnUserID()) {
                        MainClass.removeFriend(json.ReceiverID)
                        console.log(`You have unfriended user ID [${json.ReceiverID}]`)
                    } else if (json.ReceiverID === MainClass.getOwnUserID()) {
                        MainClass.removeFriend(json.UserID)
                        console.log(`User ID [${json.UserID}] has unfriended you`)
                    }
                    break
                case WebsocketClass.REQUEST_DM_LIST:
                    DirectMessagesClass.addDirectMessages(json)
                    break
                case WebsocketClass.INITIAL_USER_DATA: // Server sent the client's own user ID and display name
                    MainClass.setOwnUserID(json.UserID)
                    MainClass.setOwnProfilePic(json.ProfilePic)
                    MainClass.setOwnDisplayName(json.DisplayName)
                    MainClass.setOwnPronouns(json.Pronouns)
                    MainClass.setOwnStatusText(json.StatusText)
                    MainClass.setMyFriends(json.Friends)
                    MainClass.setBlockedUsers(json.Blocks)

                    if (json.Servers.length !== 0) {
                        for (let i = 0; i < json.Servers.length; i++) {
                            console.log('Adding server ID', json.Servers[i].ServerID)
                            ServerListClass.addServer(json.Servers[i].ServerID, json.Servers[i].Owned, json.Servers[i].Name, MainClass.imageHost + json.Servers[i].Picture, 'server')
                        }
                        // this.localStorage.setServerCount(json.Servers.length)
                    } else {
                        console.log('Not being in any servers')
                    }
                    LocalStorageClass.lookForDeletedServersInLastChannels()

                    MainClass.receivedInitialUserData = true
                    console.log('Received own initial data')
                    break
                case WebsocketClass.IMAGE_HOST_ADDRESS: // Server sent image host address
                    if (json === '') {
                        console.log('Received image host address, server did not set any external')
                    } else {
                        console.log('Received image host address:', json)
                    }
                    MainClass.imageHost = json
                    MainClass.receivedImageHostAddress = true
                    break
                case WebsocketClass.UPDATE_USER_PROFILE_PIC: // replied to profile pic change
                    MainClass.setOwnProfilePic(json.Pic)
                    break
                default:
                    console.log('Server sent unknown message type')
            }
        }
        await MainClass.waitUntilBoolIsTrue(() => WebsocketClass.wsConnected)
    }

    static async preparePacket(type, struct) {
        // wait if websocket is not on yet
        await MainClass.waitUntilBoolIsTrue(() => WebsocketClass.wsConnected)

        // if (WebsocketClass.lastSendAttempt !== undefined) {
        //     const difference = Date.now() - this.lastSendAttempt
        //     if (difference < 200) {
        //         console.log(`too early, last attempt was ${difference} ms ago`)
        //         await new Promise(resolve => setTimeout(resolve, 200 - difference))
        //     }
        // }
        // WebsocketClass.lastSendAttempt = Date.now()

        // convert the type value into a single byte value that will be the packet type
        const typeByte = new Uint8Array([1])
        typeByte[0] = type

        let json = JSON.stringify(struct)

        // workaround to turn uint64 values in json from string to integer type so server can process
        // numbers longer than 16 characters
        // json = json.replace(/'(\d{16,})"/g, "$1");
        json = json.replace(/(?<!\"Message\"\s*:\s*)\"(\d{16,})\"/g, "$1");

        console.log('Json to prepare for sending:', json)


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

        WebsocketClass.wsClient.send(packet)

        // WebsocketClass.canSendPacket = false
        // setTimeout(() => {
        //     WebsocketClass.canSendPacket = true
        // }, 100)
    }

    static async sendChatMessage(message, channelID, attachmentToken) { // type is 1
        console.log('Sending a chat message')
        if (channelID === 0) {
            console.warn('You have no channel selected')
            return
        }
        await WebsocketClass.preparePacket(WebsocketClass.ADD_CHAT_MESSAGE, {
            ChannelID: channelID,
            Message: message,
            AttTok: attachmentToken
        })

    }

    static async requestChatHistory(channelID, lastMessageID) {
        console.log('Requesting chat history for channel ID', channelID)
        await WebsocketClass.preparePacket(WebsocketClass.CHAT_HISTORY, {
            ChannelID: channelID,
            FromMessageID: lastMessageID,
            Older: true // if true it will request older, if false it will request newer messages from the message id
        })
    }

    static async requestDmChatHistory(chatID, lastMessageID) {
        console.log('Requesting chat history for direct message chat ID', chatID)
        await WebsocketClass.preparePacket(WebsocketClass.DM_CHAT_HISTORY, {
            ChannelID: chatID,
            FromMessageID: lastMessageID,
            Older: true // if true it will request older, if false it will request newer messages from the message id
        })
    }

    static async requestDeleteChatMessage(messageID) {
        console.log('Requesting to delete chat message ID', messageID)
        await WebsocketClass.preparePacket(WebsocketClass.DELETE_CHAT_MESSAGE, {
            MessageID: messageID
        })
    }

    static async requestAddServer(serverName) {
        console.log('Requesting to add a new server')
        await WebsocketClass.preparePacket(WebsocketClass.ADD_SERVER, {
            Name: serverName
        })
    }

    static async requestRenameServer(serverID) {
        console.log('Requesting to rename server ID:', serverID)
    }

    static async requestDeleteServer(serverID) {
        console.log('Requesting to delete server ID:', serverID)
        if (document.getElementById(serverID).getAttribute('owned') === 'false') {
            console.warn(`You don't own server ID [${serverID}]`)
            return
        }

        await WebsocketClass.preparePacket(WebsocketClass.DELETE_SERVER, {
            ServerID: serverID
        })
    }

    static async requestInviteLink(serverID) {
        if (document.getElementById(serverID).getAttribute('owned') === 'false') return
        console.log('Requesting invite link creation for server ID:', serverID)
        await WebsocketClass.preparePacket(WebsocketClass.SERVER_INVITE_LINK, {
            ServerID: serverID,
            SingleUse: true,
            Expiration: 7
        })
    }

    static async requestAddChannel() {
        if (document.getElementById(MainClass.getCurrentServerID()).getAttribute('owned') === 'false') return
        console.log('Requesting to add new channel to server ID:', MainClass.getCurrentServerID())
        await WebsocketClass.preparePacket(WebsocketClass.ADD_CHANNEL, {
            Name: 'Channel',
            ServerID: MainClass.getCurrentServerID()
        })
    }

    static async requestRemoveChannel(channelID) {
        console.log(`Requesting to remove channel ID [${channelID}] from server ID [${MainClass.getCurrentServerID()}]`)
        if (document.getElementById(MainClass.getCurrentServerID()).getAttribute('owned') === 'false') return

        // if (document.getElementById(channelID).parentElement.childElementCount <= 1) {
        //     console.warn('You can't remove last channel of a server')
        //     return
        // }

        await WebsocketClass.preparePacket(WebsocketClass.DELETE_CHANNEL, {
            ChannelID: channelID,
        })
    }

    static async requestChannelList() {
        console.log('Requesting channel list for current server ID', MainClass.getCurrentServerID())
        await WebsocketClass.preparePacket(WebsocketClass.CHANNEL_LIST, {
            ServerID: MainClass.getCurrentServerID()
        })
    }

    static async requestMemberList() {
        console.log('Requesting member list for current server ID', MainClass.getCurrentServerID())
        await WebsocketClass.preparePacket(WebsocketClass.SERVER_MEMBER_LIST, {
            ServerID: MainClass.getCurrentServerID()
        })
    }

    static async requestLeaveServer(serverID) {
        console.log('Requesting to leave a server ID', serverID)
        await WebsocketClass.preparePacket(WebsocketClass.DELETE_SERVER_MEMBER, {
            ServerID: serverID
        })
    }

    static async requestStatusChange(newStatus) {
        console.log('Requesting to change status')
        await WebsocketClass.preparePacket(WebsocketClass.UPDATE_STATUS, {
            Status: newStatus
        })
    }

    static async requestAddFriend(userID) {
        if (userID === MainClass.getOwnUserID()) {
            console.warn("You can't be friends with yourself")
            return
        }
        console.log(`Requesting to add user ID [${userID}] as friend`)
        await WebsocketClass.preparePacket(WebsocketClass.ADD_FRIEND, {
            UserID: userID
        })
    }

    static async requestBlockUser(userID) {
        if (userID === MainClass.getOwnUserID()) {
            console.warn(`You can't block yourself`)
            return
        }
        console.log(`Requesting to block user ID [${userID}]`)
        await WebsocketClass.preparePacket(WebsocketClass.BLOCK_USER, {
            UserID: userID
        })
    }

    static async requestUnfriend(userID) {
        if (userID === MainClass.getOwnUserID()) {
            console.warn(`You can't unfriend yourself`)
            return
        }
        console.log(`Requesting to unfriend user ID [${userID}]`)
        await WebsocketClass.preparePacket(WebsocketClass.UNFRIEND, {
            UserID: userID
        })
    }

    static async requestImageHostAddress() {
        console.log('Requesting image host address')
        await WebsocketClass.preparePacket(WebsocketClass.IMAGE_HOST_ADDRESS, {})
    }

    static async requestUpdateUserData(updatedUserData) {
        console.log('Requesting to update account data')
        await WebsocketClass.preparePacket(WebsocketClass.UPDATE_USER_DATA, updatedUserData)
    }

    static async requestUpdateServerData(updatedServerData) {
        console.log(`Requesting to update data of server ID [${updatedServerData.ServerID}]`)
        await WebsocketClass.preparePacket(WebsocketClass.UPDATE_SERVER_DATA, updatedServerData)
    }

    static async requestUpdateChannelData(updatedChannelData) {
        console.log(`Requesting to update data of channel ID [${updatedChannelData.ChannelID}]`)
        await WebsocketClass.preparePacket(WebsocketClass.UPDATE_CHANNEL_DATA, updatedChannelData)
    }

    static async startedTyping(typing) {
        console.log('Started typing in chat input')
        await WebsocketClass.preparePacket(WebsocketClass.STARTED_TYPING, {
            Typing: typing
        })
    }

    static async requestEditChatMessage(messageID, newMessage) {
        console.log(`Requesting to edit chatMessage message [${messageID}]`)
        await WebsocketClass.preparePacket(WebsocketClass.EDIT_CHAT_MESSAGE, {
            MessageID: messageID,
            Message: newMessage
        })
    }

    static async requestOpenDm(userID) {
        console.log(`Requesting to open dm with user ID [${userID}]`)
        await WebsocketClass.preparePacket(WebsocketClass.OPEN_DM, {
            UserID: userID
        })
    }

    static async requestDmList() {
        console.log(`Requesting list of direct messages`)
        await WebsocketClass.preparePacket(WebsocketClass.REQUEST_DM_LIST, {})
    }
}