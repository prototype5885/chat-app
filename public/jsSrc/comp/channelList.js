class ChannelListClass {
    constructor(main, chatMessageList, localStorage, chatInput) {
        this.main = main
        this.chatMessageList = chatMessageList
        this.localStorage = localStorage
        this.chatInput = chatInput

        this.AddChannelButton = document.getElementById("add-channel-button")
        this.ChannelList = document.getElementById("channel-list")
        this.Channels = document.getElementById("channels")
        this.ChannelNameTop = document.getElementById("channel-name-top")

        MainClass.registerHover(this.AddChannelButton, () => {
            BubbleClass.createBubble(this.AddChannelButton, "Create Channel", "up", 0)
        }, () => {
            BubbleClass.deleteBubble()
        })
        this.AddChannelButton.addEventListener("click", () => {
            WebsocketClass.requestAddChannel()
        })


        this.ChannelVisibilityButton = document.getElementById("channels-visibility-button")
        this.ChannelVisibilityButton.addEventListener("click", e => {
            const channels = Array.from(this.ChannelList.children)
            channels.forEach(channel => {
                const svg = document.querySelector("svg")
                // check if channel is visible
                if (channel.style.display === "") {
                    // hide if the channel isn't the current selected one
                    if (channel.id !== main.currentChannelID) {
                        channel.style.display = "none"
                    }
                    svg.setAttribute("transform", `rotate(-90, 0, 0)`);
                } else {
                    // make all channels visible
                    channel.style.display = ""
                    svg.setAttribute("transform", `rotate(0  0 0)`);
                }
            })
        })

    }

    addChannel(channelID, channelName) {
        const button = document.createElement("button")
        button.id = channelID

        const lineWidth = 1.5
        button.innerHTML += `
            <svg width="32" height="32" xmlns="http://www.w3.org/2000/svg">
                <line x1="16" y1="8" x2="13" y2="24" stroke="grey" stroke-width="${lineWidth}"/>
                <line x1="22" y1="8" x2="19" y2="24" stroke="grey" stroke-width="${lineWidth}"/>
                <line x1="10" y1="13" x2="26" y2="13" stroke="grey" stroke-width="${lineWidth}"/>
                <line x1="9" y1="19" x2="25" y2="19" stroke="grey" stroke-width="${lineWidth}"/>
            </svg>`

        const nameContainer = document.createElement("div")
        nameContainer.textContent = channelName

        button.appendChild(nameContainer)

        this.ChannelList.appendChild(button)

        MainClass.registerClick(button, () => {
            this.selectChannel(channelID)
        })
        ContextMenuClass.registerContextMenu(button, (pageX, pageY) => {
            const owned = document.getElementById(this.main.currentServerID).getAttribute("owned")
            ContextMenuClass.channelCtxMenu(channelID, owned, pageX, pageY)
        })
    }

    removeChannel(channelID) {
        console.log(`Remove channel [${channelID}]`)
        const channelButton = document.getElementById(channelID)
        channelButton.remove()

        if (channelID === this.main.currentChannelID) {
            if (this.ChannelList.firstChild !== null) {
                this.selectChannel(this.ChannelList.firstChild.id)
            } else {
                console.warn("There are no channels in current server, disabling chat...")
                this.localStorage.removeServerFromLastChannels(main.currentServerID)
                this.chatMessageList.disableChat()
            }
        }
    }

    selectChannel(channelID) {
        console.log("Selected channel ID:", channelID)
        this.chatMessageList.channelHistoryReceived = false
        main.reachedBeginningOfChannel = false

        if (main.currentChannelID !== "0" && main.currentChannelID === channelID) {
            console.log("Channel selected is already the current one")
            return
        }

        // get the selected channel
        const channelButton = document.getElementById(channelID)

        // if selected channel doesn't exist
        if (channelButton === null) {
            console.warn(`Selected channel ID [${channelID}] doesn't exist`)
            this.localStorage.removeServerFromLastChannels(main.currentServerID)
            this.chatMessageList.disableChat()
            return
        }

        // this will set the selection style to new channel
        channelButton.style.backgroundColor = ColorsClass.mainColor

        // this will remove selection style from previous channel
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
        this.chatInput.enableChatInput()
    }

    resetChannels() {
        this.ChannelList.innerHTML = ""
    }

    static getChannelName(channelID) {
        console.log(`Getting name of channel ID [${channelID}]`)
        const channel = document.getElementById(channelID)
        if (channel !== null) {
            return channel.querySelector("div").textContent
        } else {
            console.error(`Couldn't get name of channel ID [${channelID}] because channel was not found`)
            return ""
        }
    }

    static setChannelName(channelID, channelName) {
        console.log(`Setting name of channel ID [${channelID}] to [${channelName}]`)
        const channel = document.getElementById(channelID)
        if (channel !== null) {
            channel.querySelector("div").textContent = channelName
        } else {
            console.error(`Couldn't set name of channel ID [${channelID}] because channel was not found`)
        }
    }

    setCurrentChannel(channelID) {
        console.log(`Changing current channel ID to [${channelID}]`)
        main.currentChannelID = channelID
        this.localStorage.updateLastChannelsStorage()
    }
}