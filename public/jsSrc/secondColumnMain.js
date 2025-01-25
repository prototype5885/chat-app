class SecondColumnMainClass {
    static reset() {
        document.getElementById('second-column-main').innerHTML = ''
    }
}

class ChannelListClass extends SecondColumnMainClass {
    static createChannelList() {
        const secondColumnMain = document.getElementById('second-column-main')

        secondColumnMain.innerHTML = `  
            <div id="channels-visible-or-add-new">
                <button id="channels-visibility-button">
                    <svg id="channels-visibility-arrow" width="10px" height="10px">
                        <line x1="0" y1="2" x2="5" y2="8" stroke="grey" stroke-width="1.5"/>
                        <line x1="5" y1="8" x2="10" y2="2" stroke="grey" stroke-width="1.5"/>
                    </svg>
                    <label>text channels</label>
                </button>
                <button id="add-channel-button">+</button>
            </div>
            <div id="channel-list"></div>`

        const addChannelButton = document.getElementById('add-channel-button')
        const channelList = document.getElementById('channel-list')

        MainClass.registerHover(addChannelButton, () => {
            BubbleClass.createBubble(addChannelButton, 'Create Channel', 'up', 0)
        }, () => {
            BubbleClass.deleteBubble()
        })
        addChannelButton.addEventListener('click', async () => {
            await WebsocketClass.requestAddChannel()
        })


        document.getElementById('channels-visibility-button').addEventListener('click', e => {
            const channels = Array.from(channelList.children)
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
        const channelList = document.getElementById('channel-list')
        if (channelList === null) {
            console.warn(`Channel list is not loaded, can't add channel`)
            return
        }

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


        channelList.appendChild(button)

        MainClass.registerClick(button, async () => {
            await this.selectChannel(channelID, false)
        })
        ContextMenuClass.registerContextMenu(button, (pageX, pageY) => {
            const owned = document.getElementById(MainClass.getCurrentServerID()).getAttribute('owned')
            ContextMenuClass.channelCtxMenu(channelID, owned, pageX, pageY)
        })
    }

    static async removeChannel(channelID) {
        console.log(`Removing channel [${channelID}]`)
        const channelButton = document.getElementById(channelID)
        if (channelButton === null) {
            console.warn(`Can't remove channel [${channelID}], the channel button doesn't exist`)
            return
        }

        channelButton.remove()

        if (channelID === MainClass.getCurrentChannelID()) {
            const channelList = document.getElementById('channel-list')
            if (channelList.firstChild !== null) {
                await this.selectChannel(channelList.firstChild.id, false)
            } else {
                console.warn('There are no channels in current server, disabling chat...')
                LocalStorageClass.removeServerFromLastChannels(MainClass.getCurrentServerID())
                ChatMessageListClass.disableChat()
            }
        }
    }

    static async selectChannel(channelID) {
        console.log('Selected channel ID:', channelID)

        if (MainClass.getCurrentChannelID() === channelID) {
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
            // LocalStorageClass.removeServerFromLastChannels(MainClass.getCurrentServerID())
            // ChatMessageListClass.disableChat()
            return
        }

        // this will remove selection style from all channels
        const allChannelButtons = document.getElementById('channel-list').querySelectorAll('button')
        for (let i = 0; i < allChannelButtons.length; i++) {
            allChannelButtons[i].removeAttribute('style')
        }

        // this will set the selection style to new channel
        channelButton.style.backgroundColor = ColorsClass.mainColor

        // sets the placeholder text in the area where you enter the chat message
        this.setChannelName(channelID, channelButton.querySelector('div').textContent)

        this.setCurrentChannel(channelID)

        ChatMessageListClass.resetChatMessages()
        await WebsocketClass.requestChatHistory(channelID, 0)
        ChatMessageListClass.setLoadingChatMessagesIndicator(true)
        ChatMessageListClass.enableChat()
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
            document.getElementById('channel-name-top').textContent = channelName
            ChatInputClass.setChatInputPlaceHolderText(channelName)

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

class DirectMessagesClass {
    static DirectMessages = document.getElementById('direct-messages')
    static DmChatList = document.getElementById('dm-chat-list')

    static init() {
        // const dmFriendsButton = document.getElementById('dm-friends-button')
        // MainClass.registerClick(dmFriendsButton, async () => {
        //     await ChannelListClass.selectChannel(chatID, true)
        // })
    }

    static addDirectMessages(json) {
        const dmChatIDs = json
        for (let i = 0, len = dmChatIDs.length; i < len; i++) {
            this.addDirectMessage(dmChatIDs[i])
        }
        console.log(LocalStorageClass.selectLastChannel())
        // ChannelListClass.selectChannel(LocalStorageClass.selectLastChannel())
    }

    static addDirectMessage(chatID) {
        const dmButton = document.createElement('button')
        dmButton.id = chatID
        dmButton.textContent = chatID
        this.DmChatList.appendChild(dmButton)

        MainClass.registerClick(dmButton, async () => {
            await ChannelListClass.selectChannel(chatID, true)
        })
    }
}