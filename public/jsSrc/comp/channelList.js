class ChannelListServer {
    constructor(main, chatMessageList, localStorage, contextMenu) {
        this.main = main
        this.chatMessageList = chatMessageList
        this.localStorage = localStorage
        this.contextMenu = contextMenu

        // if a chat message was received in previous channel while switching to new channel,
        // this should prevent it from being added to new one


        this.AddChannelButton = document.getElementById("add-channel-button")
        this.ChannelList = document.getElementById("channel-list")
        this.Channels = document.getElementById("channels")
        this.ChannelNameTop = document.getElementById("channel-name-top")

        // add channel button

        MainClass.registerHover(this.AddChannelButton, () => {
            BubbleClass.createBubble(this.AddChannelButton, "Create Channel", "up", 0)
        }, () => {
            BubbleClass.deleteBubble()
        })
        this.AddChannelButton.addEventListener("click", () => {
            WebsocketClass.requestAddChannel()
        })
    }

    addChannel(channelID, channelName) {
        const button = document.createElement("button")
        button.id = channelID

        const nameContainer = document.createElement("div")
        nameContainer.textContent = channelName

        button.appendChild(nameContainer)

        this.ChannelList.appendChild(button)

        MainClass.registerClick(button, () => {
            this.selectChannel(channelID)
        })
        this.contextMenu.registerContextMenu(button, (pageX, pageY) => {
            this.contextMenu.channelCtxMenu(channelID, pageX, pageY)
        })

    }

    selectChannel(channelID) {
        console.log("Selected channel ID:", channelID)
        this.chatMessageList.channelHistoryReceived = false
        main.reachedBeginningOfChannel = false

        if (channelID === main.currentChannelID) {
            console.log("Channel selected is already the current one")
            return
        }

        const channelButton = document.getElementById(channelID)
        channelButton.style.backgroundColor = ColorsClass.mainColor

        const previousChannel = document.getElementById(main.currentChannelID)
        if (previousChannel != null) {
            document.getElementById(main.currentChannelID).removeAttribute("style")
        }

        // sets the placeholder text in the area where you enter the chat message
        const channelName = channelButton.querySelector("div").textContent
        const chatInput = document.getElementById("chat-input")
        chatInput.placeholder = `Message ${channelName}`

        this.setCurrentChannel(channelID)

        this.chatMessageList.resetChatMessages()
        WebsocketClass.requestChatHistory(channelID, 0)
        this.chatMessageList.setLoadingChatMessagesIndicator(true)
        this.ChannelNameTop.textContent = channelButton.querySelector("div").textContent
    }

    toggleChannelsVisibility() {
        const channels = Array.from(this.ChannelList.children)

        channels.forEach(channel => {
            // check if channel is visible
            if (channel.style.display === "") {
                // hide if the channel isn't the current selected one
                if (channel.id !== main.currentChannelID) {
                    channel.style.display = "none"
                }
            } else {
                // make all channels visible
                channel.style.display = ""
            }
        })
    }

    resetChannels() {
        this.ChannelList.innerHTML = ""
    }

    getChannelname(channelID) {
        return document.getElementById(channelID).querySelector("div").textContent
    }

    setChannelname(channelID, channelName) {
        document.getElementById(channelID).querySelector("div").textContent = channelName
    }

    setCurrentChannel(channelID) {
        main.currentChannelID = channelID
        this.localStorage.updateLastChannelsStorage()
    }
}