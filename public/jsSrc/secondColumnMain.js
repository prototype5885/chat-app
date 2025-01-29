class SecondColumnMainClass {
    static secondColumnMain = document.getElementById('second-column-main')

    static reset() {
        document.getElementById('second-column-main').innerHTML = ''
    }

    static selectButton(buttonID) {
        const button = document.getElementById(buttonID)
        if (button === null) {
            console.error(`Button ID ${buttonID} was not found, could not select it`)
            return
        }
        // this will remove selection style from all buttons
        const buttonGroup = document.getElementById('second-column-main').querySelectorAll('.second-column-buttons')
        for (let i = 0; i < buttonGroup.length; i++) {
            const buttons = buttonGroup[i].querySelectorAll('button')
            for (let b = 0; b < buttons.length; b++) {
                buttons[b].removeAttribute('style')
            }
        }

        // this will set the selection style to the new button
        button.style.backgroundColor = ColorsClass.selectedColor
    }
}

class ChannelListClass extends SecondColumnMainClass {
    static create() {
        this.secondColumnMain.innerHTML = `  
            <div id="channels-visible-or-add-new">
                <button id="channels-visibility-button">
                    <svg id="channels-visibility-arrow" width="10px" height="10px">
                        <line x1="0" y1="2" x2="5" y2="8" stroke="grey" stroke-width="1.5"/>
                        <line x1="5" y1="8" x2="10" y2="2" stroke="grey" stroke-width="1.5"/>
                    </svg>
                    <label>${Translation.get('textChannels')}</label>
                </button>
                <button id="add-channel-button">+</button>
            </div>
            <div id="channel-list" class="second-column-buttons"></div>`

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
            await this.selectChannel(channelID)
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
                await this.selectChannel(channelList.firstChild.id)
            } else {
                console.warn('There are no channels in current server, disabling chat...')
                LocalStorageClass.removeServerFromLastChannels(MainClass.getCurrentServerID())
                ThirdColumnMainClass.reset()
            }
        }
    }

    static async selectChannel(channelID) {
        console.log('Selected channel ID:', channelID)

        if (MainClass.getCurrentChannelID() === channelID) {
            console.log('Channel selected is already the current one')
            return
        }

        // get the selected channel
        const channelButton = document.getElementById(channelID)

        // if selected channel doesn't exist
        if (channelButton === null) {
            console.warn(`Selected channel ID [${channelID}] doesn't exist`)
            return
        }

        this.selectButton(channelID)

        this.setCurrentChannel(channelID)

        ChatMessageListClass.create()
        // sets the placeholder text in the area where you enter the chat message
        this.setChannelName(channelID, channelButton.querySelector('div').textContent)
        ChatMessageListClass.setLoadingChatMessagesIndicator(true)


        // ChatMessageListClass.resetChatMessages()
        await WebsocketClass.requestChatHistory(channelID, 0)

        // ChatMessageListClass.enableChat()
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

class DirectMessagesClass extends SecondColumnMainClass {
    static create() {
        this.secondColumnMain.innerHTML = `  

                <div id="dm-buttons" class="second-column-buttons">
                    <button id="dm-friends-button"> - Friends</button>
                </div>
                <div id="dm-chat-visible-or-add-new">
                    <button id="dm-chat-visibility-button">
                        <label>direct messages</label>
                    </button>
                    <button id="create-dm-button">+</button>
                </div>
                <div id="dm-chat-list" class="second-column-buttons"></div>`


        const dmFriendsButton = document.getElementById('dm-friends-button')
        MainClass.registerClick(dmFriendsButton, async () => {
            console.log('clicked dm friends')
            FriendListClass.create()
            // await ChannelListClass.selectChannel(chatID, true)
        })
    }

    static addDirectMessages(json) {
        const dmChatIDs = json
        for (let i = 0, len = dmChatIDs.length; i < len; i++) {
            this.addDirectMessage(dmChatIDs[i])
        }
    }

    static addDirectMessage(chatID) {
        const dmChatList = document.getElementById('dm-chat-list')

        // const dmButton = document.createElement('button')
        // dmButton.id = chatID
        // dmButton.textContent = chatID
        // dmChatList.appendChild(dmButton)

        dmChatList.innerHTML += `
            <button id="${chatID}">
                <div>${chatID}</div>   
            </button>`

        const dmButton = document.getElementById(chatID)

        MainClass.registerClick(dmButton, async () => {
            await ChannelListClass.selectChannel(chatID)
        })
    }
}