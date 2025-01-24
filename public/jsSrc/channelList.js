class ChannelListClass {
    static AddChannelButton = document.getElementById('add-channel-button')
    static ChannelList = document.getElementById('channel-list')
    static Channels = document.getElementById('channels')
    static ChannelNameTop = document.getElementById('channel-name-top')


    static init() {
        MainClass.registerHover(this.AddChannelButton, () => {
            BubbleClass.createBubble(this.AddChannelButton, 'Create Channel', 'up', 0)
        }, () => {
            BubbleClass.deleteBubble()
        })
        this.AddChannelButton.addEventListener('click', async () => {
            await WebsocketClass.requestAddChannel()
        })


        document.getElementById('channels-visibility-button').addEventListener('click', e => {
            const channels = Array.from(this.ChannelList.children)
            channels.forEach(channel => {
                const svg = document.querySelector('svg')
                // check if channel is visible
                if (channel.style.display === '') {
                    // hide if the channel isn't the current selected one
                    if (channel.id !== MainClass.getCurrentChannelID()) {
                        channel.style.display = 'none'
                    }
                    svg.setAttribute('transform', `rotate(-90, 0, 0)`);
                } else {
                    // make all channels visible
                    channel.style.display = ''
                    svg.setAttribute('transform', `rotate(0  0 0)`);
                }
            })
        })
    }

    static addChannel(channelID, channelName) {
        const button = document.createElement('button')
        button.id = channelID

        const lineWidth = 1.5
        button.innerHTML += `
            <svg width='32' height='32' xmlns='http://www.w3.org/2000/svg'>
                <line x1='16' y1='8' x2='13' y2='24' stroke='grey' stroke-width='${lineWidth}'/>
                <line x1='22' y1='8' x2='19' y2='24' stroke='grey' stroke-width='${lineWidth}'/>
                <line x1='10' y1='13' x2='26' y2='13' stroke='grey' stroke-width='${lineWidth}'/>
                <line x1='9' y1='19' x2='25' y2='19' stroke='grey' stroke-width='${lineWidth}'/>
            </svg>`

        const nameContainer = document.createElement('div')
        nameContainer.textContent = channelName

        button.appendChild(nameContainer)

        this.ChannelList.appendChild(button)

        MainClass.registerClick(button, async () => {
            await this.selectChannel(channelID, false)
        })
        ContextMenuClass.registerContextMenu(button, (pageX, pageY) => {
            const owned = document.getElementById(MainClass.getCurrentServerID()).getAttribute('owned')
            ContextMenuClass.channelCtxMenu(channelID, owned, pageX, pageY)
        })
    }

    static async removeChannel(channelID) {
        console.log(`Remove channel [${channelID}]`)
        const channelButton = document.getElementById(channelID)
        channelButton.remove()

        if (channelID === MainClass.getCurrentChannelID()) {
            if (this.ChannelList.firstChild !== null) {
                await this.selectChannel(this.ChannelList.firstChild.id, false)
            } else {
                console.warn('There are no channels in current server, disabling chat...')
                LocalStorageClass.removeServerFromLastChannels(MainClass.getCurrentServerID())
                ChatMessageListClass.disableChat()
            }
        }
    }

    static async selectChannel(channelID) {
        console.log('Selected channel ID:', channelID)

        if (MainClass.getCurrentChannelID() !== '0' && MainClass.getCurrentChannelID() === channelID) {
            console.log('Channel selected is already the current one')
            return
        }

        ChatMessageListClass.channelHistoryReceived = false
        MainClass.reachedBeginningOfChannel = false

        // get the selected channel
        const channelButton = document.getElementById(channelID)

        // if selected channel doesn't exist
        if (channelButton === null) {
            console.warn(`Selected channel ID [${channelID}] doesn't exist`)
            LocalStorageClass.removeServerFromLastChannels(MainClass.getCurrentServerID())
            ChatMessageListClass.disableChat()
            return
        }

        // this will remove selection style from all channels
        const allChannelButtons = this.ChannelList.querySelectorAll('button')
        for (let i = 0; i < allChannelButtons.length; i++) {
            allChannelButtons[i].removeAttribute('style')
        }


        // this will set the selection style to new channel
        channelButton.style.backgroundColor = ColorsClass.mainColor

        // sets the placeholder text in the area where you enter the chat message
        const channelName = channelButton.querySelector('div').textContent
        const chatInput = document.getElementById('chat-input')
        chatInput.placeholder = `Message #${channelName}`

        this.setCurrentChannel(channelID)

        ChatMessageListClass.resetChatMessages()
        await WebsocketClass.requestChatHistory(channelID, 0)
        ChatMessageListClass.setLoadingChatMessagesIndicator(true)
        this.ChannelNameTop.textContent = channelButton.querySelector('div').textContent
        ChatMessageListClass.enableChat()
    }

    static resetChannels() {
        this.ChannelList.innerHTML = ''
    }

    static getChannelName(channelID) {
        console.log(`Getting name of channel ID [${channelID}]`)
        const channel = document.getElementById(channelID)
        if (channel !== null) {
            return channel.querySelector('div').textContent
        } else {
            console.error(`Couldn't get name of channel ID [${channelID}] because channel was not found`)
            return ''
        }
    }

    static setChannelName(channelID, channelName) {
        console.log(`Setting name of channel ID [${channelID}] to [${channelName}]`)
        const channel = document.getElementById(channelID)
        if (channel !== null) {
            channel.querySelector('div').textContent = channelName
        } else {
            console.error(`Couldn't set name of channel ID [${channelID}] because channel was not found`)
        }
    }

    static setCurrentChannel(channelID) {
        console.log(`Changing current channel ID to [${channelID}]`)
        MainClass.setCurrentChannelID(channelID)
        LocalStorageClass.updateLastChannelsStorage()
    }
}