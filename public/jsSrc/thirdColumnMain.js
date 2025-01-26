class ThirdColumnMainClass {
    static reset() {
        console.log('Resetting third column main')
        document.getElementById('third-column-main').innerHTML = ''
        MainClass.setCurrentChannelID('0')
    }
}

class ChatMessageListClass {
    static #amountOfMessagesLoaded = 0
    static #channelHistoryReceived = false
    static #reachedBeginningOfChannel = false

    static #peopleTyping = []

    static locale = navigator.language
    static dateHourShort = {timeStyle: 'short'}
    static dateOptionsDay = {year: 'numeric', month: 'long', day: 'numeric'}
    static dateOptionsLong = {year: 'numeric', month: 'long', day: 'numeric', hour: '2-digit', minute: '2-digit'}


    static create() {
        this.resetChatMessages()

        // the div in list makes sure there will be a little gap between chat input box
        // and the chat messages when user is viewing the latest message at the bottom
        document.getElementById('third-column-main').innerHTML = `
            <div id="chat-loading-indicator">Loading</div>
            <ul id="chat-message-list" onscroll="ChatMessageListClass.messageListScrolled(event)">
               <div></div>
            </ul>`

        ChatInputClass.create()
    }

    static async messageListScrolled(event) {
        if (!this.#reachedBeginningOfChannel && event.currentTarget.scrollTop < 200) {
            console.log('check if needs history')
            await this.checkIfNeedsHistory()
        }
    }


    static resetChatMessages() {
        const chatMessageList = document.getElementById('chat-message-list')
        if (chatMessageList === null) {
            console.error(`Can't reset chat messages as chat message list isn't loaded`)
            return
        }
        this.#channelHistoryReceived = false
        this.#reachedBeginningOfChannel = false
        this.#amountOfMessagesLoaded = 0
    }

    static updateDaySeparatorsInChat() {
        const chatMessageList = document.getElementById('chat-message-list')
        if (chatMessageList === null) {
            console.error(`Can't update day/date separators in chat because chat message list isnt loaded`)
            return
        }

        // insert a separator that separates message each day
        const messages = Array.from(chatMessageList.querySelectorAll('li.msg'))

        // remove previous day separators
        const daySeparators = Array.from(chatMessageList.querySelectorAll('li.date-between-msgs'))
        daySeparators.forEach(daySeparator => {
            daySeparator.remove()
        })

        let lastDate = ''
        for (let i = 0; i < messages.length; i++) {
            // extract date from message id
            const date = new Date(Number((BigInt(messages[i].id) >> BigInt(22)))).toLocaleDateString(this.locale, this.dateOptionsDay)

            if (lastDate !== '' && lastDate !== date) {
                const dateBetweenMsgs = document.createElement('li')
                dateBetweenMsgs.className = 'date-between-msgs'

                const leftLine = document.createElement('div')

                const dateText = document.createElement('span')
                dateText.textContent = date

                const rightLine = document.createElement('div')

                dateBetweenMsgs.appendChild(leftLine)
                dateBetweenMsgs.appendChild(dateText)
                dateBetweenMsgs.appendChild(rightLine)

                chatMessageList.insertBefore(dateBetweenMsgs, messages[i])
            }
            lastDate = date
        }
    }

    static removeGhostMessages() {
        const chatMessageList = document.getElementById('chat-message-list')
        if (chatMessageList === null) {
            console.error(`Can't remove ghost messages because chat message list isn't loaded`)
            return
        }

        const ghostMessages = chatMessageList.querySelectorAll('.ghost-msg')
        if (ghostMessages.length === 0) {
            return
        }
        for (let i = 0; i < ghostMessages.length; i++) {
            ghostMessages[i].remove()
        }
    }

    // adds the new chat message into html
    static addChatMessage(messageID, userID, message, attachments, edited, ghost) {
        if (document.getElementById(messageID) !== null) {
            console.error(`This message already exists in chat list with same ID, won't add it again: ${messageID}`)
            return
        }

        const chatMessageList = document.getElementById('chat-message-list')
        if (chatMessageList === null) {
            console.error(`Can't add chat message because chat message list isnt loaded`)
            return
        }

        this.removeUserFromTypingList(userID)

        // extract the message date from messageID
        let msgDate
        if (ghost) {
            msgDate = new Date()
        } else {
            msgDate = MainClass.extractDateFromId(messageID)
        }

        let msgDateStr = ''

        const today = new Date()

        const yesterday = new Date()
        yesterday.setDate(yesterday.getDate() - 1)

        if (msgDate.toLocaleDateString() === today.toLocaleDateString()) {
            msgDateStr = Translation.get('today') + ' ' + msgDate.toLocaleTimeString(this.locale, this.dateHourShort)
        } else if (msgDate.toLocaleDateString() === yesterday.toLocaleDateString()) {
            msgDateStr = Translation.get('yesterday') + ' ' + msgDate.toLocaleTimeString(this.locale, this.dateHourShort)
        } else {
            msgDateStr = msgDate.toLocaleString(this.locale, this.dateOptionsLong)
        }

        const userInfo = MemberListClass.getUserInfo(userID)

        // create a <li> that holds the message
        const li = document.createElement('li')
        if (ghost) {
            li.className = 'msg ghost-msg'
        } else {
            li.className = 'msg'
        }
        li.id = messageID
        li.setAttribute('user-id', userID)

        let owner = false
        if (userID === MainClass.getOwnUserID()) {
            owner = true
        }

        ContextMenuClass.registerContextMenu(li, (pageX, pageY) => {
            ContextMenuClass.messageCtxMenu(messageID, owner, pageX, pageY)
        })

        // create a <img> that shows profile pic on the left
        const img = document.createElement('img')
        img.className = 'msg-profile-pic'

        if (userInfo.pic !== '') {
            img.src = userInfo.pic
        } else {
            img.src = '/content/static/discord.webp'
        }

        img.width = 40
        img.height = 40


        // MainClass.registerClick(img, () => {
        //     const pictureViewerContainer = document.getElementById('picture-viewer-container')
        //     const pictureViewer = document.getElementById('picture-viewer')
        //
        //     pictureViewerContainer.style.display = 'block'
        //     pictureViewer.src = img.src
        //
        //     ContextMenuClass.registerContextMenu(pictureViewer, (pageX, pageY) => {
        //         ContextMenuClass.pictureCtxMenu(img.src, img.src, pageX, pageY)
        //     })
        //
        //     MainClass.registerClick(pictureViewerContainer, () => {
        //         pictureViewerContainer.style.display = 'none'
        //         pictureViewer.src = ''
        //     })
        // })


        ContextMenuClass.registerContextMenu(img, (pageX, pageY) => {
            ContextMenuClass.userCtxMenu(userID, pageX, pageY)
        })

        // create a nested <div> that will contain sender name, message and date
        const msgDataDiv = document.createElement('div')
        msgDataDiv.className = 'msg-data'

        // inside that create a sub nested <div> that contains sender name and date
        const msgNameAndDateDiv = document.createElement('div')
        msgNameAndDateDiv.className = 'msg-name-and-date'

        // and inside that create a <div> that displays the sender's name on the left
        const msgNameDiv = document.createElement('div')
        msgNameDiv.className = 'msg-user-name'
        msgNameDiv.textContent = userInfo.username

        ContextMenuClass.registerContextMenu(msgNameDiv, (pageX, pageY) => {
            ContextMenuClass.userCtxMenu(userID, pageX, pageY)
        })

        // and next to it create a <div> that displays the date of msg on the right
        const msgDateDiv = document.createElement('div')
        msgDateDiv.className = 'msg-date'
        msgDateDiv.textContent = msgDateStr

        // append name and date to msgNameAndDateDiv
        msgNameAndDateDiv.appendChild(msgNameDiv)
        msgNameAndDateDiv.appendChild(msgDateDiv)

        msgDataDiv.appendChild(msgNameAndDateDiv)

        const msgLeftRightContainer = document.createElement('div')
        msgLeftRightContainer.className = 'msg-left-right-container'

        const msgLeftSide = document.createElement('div')
        msgLeftSide.className = 'msg-left-side'
        msgLeftRightContainer.appendChild(msgLeftSide)

        const msgRightSide = document.createElement('div')
        msgRightSide.className = 'msg-right-side'
        msgLeftRightContainer.appendChild(msgRightSide)

        msgDataDiv.appendChild(msgLeftRightContainer)

        const msgDateShortContainer = document.createElement('div')

        msgDateShortContainer.textContent = `${msgDate.toLocaleTimeString(this.locale, this.dateHourShort)}`
        msgDateShortContainer.className = 'msg-date-short'

        msgLeftSide.appendChild(msgDateShortContainer)

        const msgTextContainer = document.createElement('div')
        msgTextContainer.className = 'msg-text-container'

        // now create a <div> under name and date that displays the message
        const msgTextDiv = document.createElement('span')
        msgTextDiv.className = 'msg-text'

        // look for URLs in the message and make them clickable
        msgTextDiv.innerHTML = message.replace(/https?:\/\/[^\s/$.?#].[^\s]*/g, (url) => {
            if (url.endsWith('.gif') || url.endsWith('.jpg') || url.endsWith('.jpeg') || url.endsWith('.png') || url.endsWith('.webp')) {
                return `<a href='${url}' target='_blank'><img src='${url}'></a>`
            } else {
                return `<a href='${url}' target='_blank'>${url}</a>`
            }
        })

        // append both name/date <div> and msg <div> to msgDatDiv
        msgTextContainer.appendChild(msgTextDiv)

        msgRightSide.appendChild(msgTextContainer)

        // append both the profile pic and message data to the <li>
        li.appendChild(img)
        li.appendChild(msgDataDiv)

        // insert the messages ordered by message id
        const messages = chatMessageList.querySelectorAll('li.msg')

        let inserted = false
        for (let i = 0; i < messages.length; i++) {
            if (li.id < messages[i].id) {
                chatMessageList.insertBefore(li, messages[i])
                inserted = true
                break
            }
        }

        if (!inserted) {
            chatMessageList.appendChild(li)
        }

        // add attachments
        if (attachments !== undefined && attachments !== null && attachments.length > 0) {
            const attachmentContainer = document.createElement('div')
            for (let i = 0; i < attachments.length; i++) {
                const extension = attachments[i].Name.split('.').pop().toLowerCase()

                const hashHex = MainClass.base64toSha256(attachments[i].Hash)

                const path = `/content/attachments/${hashHex + '.' + extension}`

                switch (extension) {
                    case 'mp3':
                    case 'wav':
                    case 'ogg':
                    case 'flac':
                        attachmentContainer.className = 'message-attachment-audios'
                        attachmentContainer.innerHTML += `
                        <audio controls class='attachment-audio'>
                            <source src='${path}'>${attachments[i].Name}</source>
                        </audio>`
                        msgRightSide.appendChild(attachmentContainer)
                        break
                    case 'mp4':
                    case 'webm':
                    case 'mov':
                        attachmentContainer.className = 'message-attachment-videos'

                        attachmentContainer.innerHTML += `
                        <video controls class='attachment-video'>
                            <source src='${path}'>${attachments[i].Name}</source>
                        </video>`
                        msgRightSide.appendChild(attachmentContainer)
                        break
                    case 'jpg':
                    case 'jpeg':
                    case 'webp':
                    case 'png':
                    case 'gif':
                    case 'jfif':
                        attachmentContainer.className = 'message-attachment-pictures'

                        const img = document.createElement('img')
                        img.src = path
                        img.className = 'attachment-pic'
                        img.setAttribute('name', attachments[i].Name)
                        attachmentContainer.appendChild(img)


                        // attachmentContainer.innerHTML += `<img src='${path}' class='attachment-pic'>`
                        msgRightSide.appendChild(attachmentContainer)

                        ContextMenuClass.registerContextMenu(img, (pageX, pageY) => {
                            ContextMenuClass.pictureCtxMenu(path, attachments[i].Name, pageX, pageY)
                        })


                        MainClass.registerClick(img, () => {
                            const pictureViewerContainer = document.getElementById('picture-viewer-container')
                            const pictureViewer = document.getElementById('picture-viewer')

                            pictureViewerContainer.style.display = 'block'
                            pictureViewer.src = path
                            console.log('clicked on:', path)
                            console.log('pic name:', img.getAttribute('name'))

                            MainClass.currentPictureViewerPicName = img.getAttribute('name')
                            MainClass.currentPictureViewerPicPath = img.src

                            MainClass.registerClick(pictureViewerContainer, () => {
                                pictureViewerContainer.style.display = 'none'
                                pictureViewer.src = ''
                                MainClass.currentPictureViewerPicPath = ''
                                MainClass.currentPictureViewerPicName = ''
                            })
                        })
                        break
                    default:
                        console.warn('Unsupported attachment type:', extension)
                        attachmentContainer.className = 'attachment-unknown'
                        attachmentContainer.innerHTML += `<a href='${path}' class='url' target='_blank'>${attachments[i].Name}</a>`
                        msgRightSide.appendChild(attachmentContainer)
                        break
                }
            }
        }
        this.checkPreviousMessage(messageID)

        if (edited) {
            this.setEdited(messageID)
        }

        li.addEventListener('mouseenter', () => {
            msgDateShortContainer.style.display = 'block'
            li.style.backgroundColor = '#2E3035'
        })

        li.addEventListener('mouseleave', () => {
            msgDateShortContainer.style.display = 'none'
            li.style.backgroundColor = ''
        })

        msgDateShortContainer.style.display = 'none'
    }

    static checkPreviousMessage(messageID) {
        const li = document.getElementById(messageID)
        const userID = li.getAttribute('user-id')
        const img = li.querySelector('.msg-profile-pic')
        const msgDataDiv = li.querySelector('.msg-data')
        const msgNameAndDateDiv = li.querySelector('.msg-name-and-date')


        function normal() {
            const msgLeftSide = li.querySelector('.msg-left-side')
            img.style.display = ''
            msgNameAndDateDiv.style.display = ''
            msgLeftSide.style.display = 'none'
            msgDataDiv.style.marginLeft = '14px'
            li.style.paddingTop = '16px'
            li.style.paddingLeft = '16px'
        }

        function short() {
            const msgDateShort = li.querySelector('.msg-date-short')
            img.style.display = 'none'
            msgNameAndDateDiv.style.display = 'none'
            msgDateShort.style.display = 'block'
            // msgDataDiv.style.marginLeft = '6px'
            li.style.paddingTop = '0px'
            li.style.paddingLeft = '0'
            // msgDateShort.style.

        }

        const previousElement = li.previousElementSibling
        if (previousElement === null) {
            return
        }

        if (previousElement.className === 'date-between-msgs') {
            normal()
        }

        if (previousElement.className !== 'msg') {
            normal()
        }


        // this is to make sure if previous message was before 00:00, the one after 00:00 will show up as full
        const messageDate = MainClass.extractDateFromId(messageID)
        const prevMessageDate = MainClass.extractDateFromId(previousElement.id)

        const messageUnixMin = Math.floor(messageDate.getTime() / 1000 / 60)
        const prevMessageUnixMin = Math.floor(prevMessageDate.getTime() / 1000 / 60)

        const elapsedMinutes = messageUnixMin - prevMessageUnixMin

        let sameUser = false
        if (previousElement.getAttribute('user-id') === userID) {
            sameUser = true
        }

        let sameDay = false
        if (messageDate.toLocaleDateString() === prevMessageDate.toLocaleDateString()) {
            sameDay = true
        }

        let older = false
        if (elapsedMinutes > 5) {
            older = true
        }

        if (sameUser && !older && sameDay) {
            short()
        } else {
            normal()
        }
    }

    static deleteChatMessage(json) {
        const messageID = json.MessageID

        const nextMessage = document.getElementById(messageID).nextElementSibling
        let nextMessageID
        if (nextMessage !== null) {
            nextMessageID = nextMessage.id
        }

        console.log(`Deleting message ID [${messageID}]`)
        document.getElementById(messageID).remove()
        this.amountOfMessagesChanged()
        if (nextMessage !== null) {
            this.checkPreviousMessage(nextMessageID)
        }
    }

    static async chatMessageReceived(json) {
        console.log('msg: ', json.Msg)
        if (!this.#channelHistoryReceived) {
            console.warn(`Won't add received chat message as it was meant for the previous channel`)
            return
        }
        if (!MainClass.memberListLoaded) {
            console.log('Loading member list...')
            await MainClass.waitUntilBoolIsTrue(() => MainClass.memberListLoaded) // wait until members are loaded
        }

        // console.log(`New chat message ID [${json.MsgID}] received`)
        this.addChatMessage(json.MsgID, json.UserID, json.Msg, json.Att, json.Edited, false)

        // play notification sound if messages is from other user
        if (json.UserID !== MainClass.getOwnUserID()) {
            if (Notification.permission === 'granted') {
                NotificationClass.sendNotification(json.UserID, json.Msg)
            } else {
                NotificationClass.NotificationSound.play()
            }
        }

        const chatMessageList = document.getElementById('chat-message-list')
        if (chatMessageList === null) {
            console.error(`Can't scroll down to bottom on new message because chat message list isnt loaded`)
            return
        }
        if (MainClass.getScrollDistanceFromBottom(chatMessageList) < 200 || json.IDu === MainClass.getOwnUserID()) {
            chatMessageList.scrollTo({
                top: chatMessageList.scrollHeight,
                behavior: 'smooth'
            })
        } else {
            console.log('Too far from current chat messages, not scrolling down on new message')
        }

        this.amountOfMessagesChanged()
    }

    static async chatHistoryReceived(json) {
        console.log(`Requested chat history for channel ID [${json[0]}] arrived`)
        if (MainClass.getCurrentChannelID() !== json[0]) {
            console.warn(`The received chat history was meant to be for channel ID [${json[0]}], but you are on [${MainClass.getCurrentChannelID()}]`)
            return
        }
        const chatMessageList = document.getElementById('chat-message-list')
        if (chatMessageList === null) {
            console.error(`Can't insert chat history because chat message list isnt loaded`)
            return
        }

        // if (main.currentChannelID === this.lastReceivedChannelHistoryID) {
        //     console.warn('You already received the chat history for this channel')
        //     return
        // }
        if (!MainClass.memberListLoaded) {
            await MainClass.waitUntilBoolIsTrue(() => MainClass.memberListLoaded) // wait until members are loaded
        }

        const chatMessages = []
        if (json[1].length !== 0) {
            // runs if json contains chat history
            for (let u = 0; u < json[1].length; u++) {
                for (let m = 0; m < json[1][u].Msgs.length; m++) {
                    // add message in this format to a chatMessages list
                    const chatMessage = {
                        MessageID: json[1][u].Msgs[m][0], // message id
                        UserID: json[1][u].UserID, // user id
                        Message: json[1][u].Msgs[m][1], // message
                        Edited: json[1][u].Msgs[m][2],
                        Attachments: json[1][u].Msgs[m][3] // attachments
                    }
                    chatMessages.push(chatMessage)
                }
            }
            // sort the history here because message history is not received ordered
            chatMessages.sort((a, b) => a.MessageID - b.MessageID)
            for (let i = 0; i < chatMessages.length; i++) {
                this.addChatMessage(chatMessages[i].MessageID, chatMessages[i].UserID, chatMessages[i].Message, chatMessages[i].Attachments, chatMessages[i].Edited, false)
            }

            // only auto scroll down when entering channel, and not when
            // server sends rest of history while scrolling up manually
            if (!this.#channelHistoryReceived) {
                // this runs when entered a channel
                chatMessageList.scrollTo({
                    top: chatMessageList.scrollHeight,
                    behavior: 'instant'
                })
                // set this so it won't scroll down anymore as messages arrive while scrolling up
                // and won't request useless chat history requests when scrolling on top
                // if already reached the beginning
                MainClass.lastChannelID = MainClass.getCurrentChannelID()
            }
        } else {
            // run if server sent json that doesn't contain any more messages
            if (MainClass.getCurrentChannelID() === MainClass.lastChannelID) {
                // this can only run if already in channel
                console.warn(`Reached the beginning of the chat, don't request more`)
                // will become false upon entering another channel
                this.#reachedBeginningOfChannel = true
            } else {
                // and this only when entering a channel
                console.warn('Current channel has no chat history')
            }
        }
        this.setLoadingChatMessagesIndicator(false)
        this.amountOfMessagesChanged()
        this.#channelHistoryReceived = true
    }

    static someoneStartedTyping(typing, userID, channelID) {
        if (channelID === MainClass.getCurrentChannelID()) {
            if (typing) {
                this.addUserToTypingList(userID)
            } else {
                this.removeUserFromTypingList(userID)
            }
        }
    }

    static addUserToTypingList(userID) {
        if (userID === MainClass.getOwnUserID()) {
            return
        }
        if (!this.#peopleTyping.includes(userID)) {
            const timerID = setTimeout(() => {
                this.removeUserFromTypingList(userID);
            }, 20000)
            this.#peopleTyping.push({UserID: userID, Timer: timerID})
            this.setTypingText()
        }
    }


    static removeUserFromTypingList(userID) {
        const index = this.#peopleTyping.findIndex(pair => pair.UserID === userID)
        if (index !== -1) {
            clearTimeout(this.#peopleTyping[index].Timer)
            this.#peopleTyping.splice(index, 1)
            this.setTypingText()
        }
    }

    static setTypingText() {
        const someoneTyping = document.getElementById('someone-typing')
        const svgContainer = document.getElementById('svg-container')

        if (this.#peopleTyping.length !== 0) {
            svgContainer.style.display = 'flex'
        }
        if (this.#peopleTyping.length === 0) {
            someoneTyping.textContent = ''
            svgContainer.style.display = 'none'
        } else if (this.#peopleTyping.length === 1) {
            someoneTyping.innerHTML = `<b>${MemberListClass.getMemberName(this.#peopleTyping[0].UserID)}</b> is typing...`
        } else if (this.#peopleTyping.length > 1 && this.#peopleTyping.length <= 4) {
            let text = ''
            for (let i = 0; i < this.#peopleTyping.length; i++) {
                if (i !== this.#peopleTyping.length - 1) {
                    text += `<b>${MemberListClass.getMemberName(this.#peopleTyping[i].UserID)}</b> and `
                } else {
                    text += `<b>${MemberListClass.getMemberName(this.#peopleTyping[i].UserID)}</b> are typing...`
                }
            }
            someoneTyping.innerHTML = text
        } else {
            someoneTyping.innerHTML = `<b>${this.#peopleTyping.length} people</b>  are typing...`
        }
    }

    static amountOfMessagesChanged() {
        this.#amountOfMessagesLoaded = document.getElementById('chat-message-list').querySelectorAll('li').length
        console.log('Amount of messages loaded:', this.#amountOfMessagesLoaded)
        this.updateDaySeparatorsInChat()
    }

    static changeDisplayNameInChatMessageList(userID, newDisplayName) {
        const chatMessages = document.getElementById('chat-message-list').querySelectorAll('.msg')
        chatMessages.forEach((chatMessage) => {
            if (chatMessage.getAttribute('user-id') === userID) {
                chatMessage.querySelector('.msg-user-name').textContent = newDisplayName
            }
        })
    }

    static setChatMessageProfilePic(userID, pic) {
        const chatMessages = document.getElementById('chat-message-list').querySelectorAll('.msg')
        chatMessages.forEach((chatMessage) => {
            if (chatMessage.getAttribute('user-id') === userID) {
                chatMessage.querySelector('.msg-profile-pic').src = pic
            }
        })
    }

    static async checkIfNeedsHistory() {
        if (this.#channelHistoryReceived && this.#amountOfMessagesLoaded >= 50) {
            const chatMessage = document.getElementById('chat-message-list').querySelector('li.msg')
            if (chatMessage != null) {
                await WebsocketClass.requestChatHistory(MainClass.getCurrentChannelID(), chatMessage.id)
                this.setLoadingChatMessagesIndicator(true)
            }
        }
    }

    static setLoadingChatMessagesIndicator(loading) {
        const chatLoadingIndicator = document.getElementById('chat-loading-indicator')
        const chatMessageList = document.getElementById('chat-message-list')
        if (loading) {
            chatLoadingIndicator.style.display = 'flex'
            chatMessageList.style.overflowY = 'hidden'
        } else {
            chatLoadingIndicator.style.display = 'none'
            chatMessageList.style.overflowY = ''
        }
    }

    static setMessagePic(messageID, pic) {
        const msgPic = document.getElementById(messageID).querySelector('.msg-profile-pic')
        if (pic !== '') {
            msgPic.src = pic
        } else {
            msgPic.src = 'content/static/questionmark.svg'
        }
    }

    static disableChat() {
        console.log('Disabling chat')
        this.resetChatMessages()
        MainClass.setCurrentChannelID('0')
        document.getElementById('chat-container').style.display = 'none'
    }

    static enableChat() {
        console.log('Enabling chat')
        document.getElementById('chat-container').style.display = 'flex'
        FriendListClass.friendListContainer.style.display = 'none'
    }

    static editChatMessage(messageID, newMessage) {
        console.log(`Editing chat message ID [${messageID}]`)
        const msgElement = document.getElementById(messageID)
        if (msgElement === null) {
            console.log(`Message ID [${messageID}] was not found, possibly out of view, unable to edit`)
        } else {
            msgElement.querySelector('.msg-text').textContent = newMessage
            this.setEdited(messageID)
        }
    }

    static setEdited(messageID) {
        console.log(`Marking chat message ID [${messageID}] as edited`)
        const msg = document.getElementById(messageID)
        if (msg !== null && msg.querySelector('.msg-edited') === null) {
            const editedSpan = document.createElement('span')
            editedSpan.className = 'msg-edited'
            editedSpan.textContent = '(edited)'

            const msgTextContainer = msg.querySelector('.msg-text-container')
            msgTextContainer.appendChild(editedSpan)

        }
    }
}

class ChatInputClass {
    static #typing = false
    static sendingChatMsg = false

    static maxFiles = 5
    static files = []

    static #canDrag = false

    static create() {
        document.getElementById('third-column-main').innerHTML += `
                <div id="chat-input-container">
                        <div id="attachment-list"></div>
                        <div id="chat-input-form">
                            <label id="upload-percentage"></label>
                            <button class="chat-button" id="attachment-button">
                                <svg class="chat-button-icon" id="attachment-button-icon" width="24" height="24"
                                     xmlns="http://www.w3.org/2000/svg">
                                    <line x1="12" y1="5" x2="12" y2="19" stroke="#383a40" stroke-width="2"/>
                                    <line x1="5" y1="12" x2="19" y2="12" stroke="#383a40" stroke-width="2"/>
                                </svg>
                            </button>
                            <textarea autocomplete="off" id="chat-input" placeholder="Message ..."></textarea>
                            <!--                    <button class="chat-button" id="send-button" onclick="readChatInput()">-->
                            <!--                        <div class="chat-button-icon" id="send-button-icon">Send</div>-->
                            <!--                    </button>-->
                        </div>

                    </div>
                    <div id="someone-typing-container">
                        <div id="svg-container" style="display: none;">
                            <svg width="6" height="6" xmlns="http://www.w3.org/2000/svg">
                                <circle cx="3" cy="3" r="3" fill="white"/>
                            </svg>
                            <svg width="6" height="6" xmlns="http://www.w3.org/2000/svg">
                                <circle cx="3" cy="3" r="3" fill="white"/>
                            </svg>
                            <svg width="6" height="6" xmlns="http://www.w3.org/2000/svg">
                                <circle cx="3" cy="3" r="3" fill="white"/>
                            </svg>
                        </div>
                        <label id="someone-typing"></label>
                    </div>
                </div>`

        const chatInput = document.getElementById('chat-input')

        chatInput.addEventListener('keydown', this.chatEnterPressed.bind(this))

        chatInput.addEventListener('input', async () => {
            this.resizeChatInput()
            await this.checkIfTyping()
        })

        document.getElementById('attachment-button').addEventListener('click', () => {
            this.AttachmentInput.click()
        })

        // this is when user clicks on attachment button and uploads files from there
        this.AttachmentInput = document.getElementById('attachment-input')
        this.AttachmentInput.addEventListener('change', () => {
            for (let i = 0; i < this.AttachmentInput.files.length; i++) {
                this.addAttachment(this.AttachmentInput.files[i])
            }
        })

        this.fileDropZone = document.getElementById('file-drop-zone')
        this.fileDropMsg = document.getElementById('file-drop-msg')

        this.fileDropZone.addEventListener('dragenter', e => {
            e.preventDefault()
            this.#canDrag = true
            console.log('Started dragging a file into window')
        })

        this.fileDropZone.addEventListener('dragover', e => {
            e.preventDefault()
            // if (this.#canDrag) {
            this.fileDropZone.style.display = 'flex'
            this.fileDropMsg.textContent = 'Upload to:\n\n' + MainClass.getCurrentChannelID()
            // }
        })

        // this when user drags files into webpage
        this.fileDropZone.addEventListener('drop', e => {
            e.preventDefault()
            console.log('dropped file')

            for (let i = 0; i < e.dataTransfer.items.length; i++) {
                const file = e.dataTransfer.items[i]
                if (e.dataTransfer.items[i].kind === 'file') {
                    const file = e.dataTransfer.items[i].getAsFile();
                    this.addAttachment(file)
                }
            }
            this.hideFileDropUI()
            this.#canDrag = false
        })

        this.fileDropZone.addEventListener('dragleave', e => {
            e.preventDefault()
            console.log('Cancelled file dragging')
            this.hideFileDropUI()
            this.#canDrag = false
        })

        document.addEventListener('paste', e => {
            const items = e.clipboardData.items
            if (items) {
                for (let i = 0; i < items.length; i++) {
                    const item = items[i]

                    // Only handle files
                    if (item.kind === 'file') {
                        const file = item.getAsFile()
                        this.addAttachment(file)
                    }
                }
            }
        })

        this.resizeChatInput()
    }

    static reset() {

    }


    // dynamically resize the chat input textarea to fit the text content
    // runs whenever the chat input textarea content changes
    // or pressed enter
    static resizeChatInput() {
        const chatInput = document.getElementById('chat-input')
        chatInput.style.height = 'auto'
        chatInput.style.height = chatInput.scrollHeight + 'px'
    }

    static async checkIfTyping() {
        const chatInput = document.getElementById('chat-input')
        if (chatInput.value !== '' && !this.#typing) {
            this.#typing = true
            console.log('started typing')
            await WebsocketClass.startedTyping(true)
        }
        if (chatInput.value === '') {
            this.#typing = false
            console.log('stopped typing')
            await WebsocketClass.startedTyping(false)
        }
    }

    // send the text message on enter
    static async chatEnterPressed(event) {
        if (event.key === 'Enter' && !event.shiftKey) {
            event.preventDefault()
            if (this.sendingChatMsg) {
                console.warn(`Sending a message currently, can't send any new yet`)
                return
            }
            this.disableChatInput()
            let attachmentToken = null
            if (this.files.length !== 0) {
                console.log(`Chat message has [${this.files.length}] attachments, sending those first...`)

                // these hashes are of the attachments that already exist on server, no need to upload them
                const existingHashes = await this.checkAttachments()

                attachmentToken = await this.sendAttachment(existingHashes)
                console.log('http response to uploading attachment:', attachmentToken)
            }

            const chatInput = document.getElementById('chat-input')
            if (chatInput.value.trim() !== '' || attachmentToken !== null) {
                if (attachmentToken !== null) {
                    await WebsocketClass.sendChatMessage(chatInput.value.trim(), MainClass.getCurrentChannelID(), attachmentToken.AttToken)
                } else {
                    await WebsocketClass.sendChatMessage(chatInput.value.trim(), MainClass.getCurrentChannelID(), null)
                }
                console.log('Resetting chat input and attachment input values')
                chatInput.value = ''
                this.AttachmentInput.value = ''
                this.resizeChatInput()
                this.#typing = false
                this.enableChatInput()
            }
        }
    }

    static disableChatInput() {
        console.warn('Disabling chat input')
        this.sendingChatMsg = true
        // this.ChatInput.disabled = true
        // this.ChatInputForm.style.backgroundColor = ColorsClass.bitDarkerColor
    }

    static enableChatInput() {
        console.warn('Enable chat input')
        this.sendingChatMsg = false
        // this.ChatInput.disabled = false
        document.getElementById('chat-input').focus()
        // this.ChatInputForm.style.backgroundColor = ''
    }

    static async checkAttachments() {
        console.log('Checking if prepare attachments already exist on server')

        let hashes = []
        for (let i = 0; i < this.files.length; i++) {
            const hash = await MainClass.calculateSHA256(this.files[i])
            hashes.push(hash)
        }

        const xhr = new XMLHttpRequest()

        return new Promise((resolve, reject) => {
            xhr.onload = function () {
                if (xhr.status === 200) {
                    const existingHashes = JSON.parse(xhr.responseText)
                    if (existingHashes === null) {
                        console.log('All attachments need to be uploaded')
                        resolve(null)
                    } else {
                        console.log(`[${existingHashes.length}] attachments don't need to be uploaded`)
                        resolve(existingHashes)
                    }

                } else {
                    console.error('Failed asking the server if given attachment hashes exist')
                    reject(null)
                }
            }


            xhr.onerror = function () {
                console.error('Error asking the server if given attachment hashes exist')
                reject(null)
            }

            xhr.open('POST', '/check-attachment')
            xhr.setRequestHeader('Content-Type', 'application/json')
            xhr.send(JSON.stringify(hashes))
        })
    }

    static async sendAttachment(existingHashes) {
        console.log('Sending attachments to server')
        const formData = new FormData()

        // loops through added attachments
        for (let i = 0; i < this.files.length; i++) {
            if (i > this.maxFiles - 1) {
                console.warn('Too many attachments, ignoring those after 4th...')
                continue
            }

            console.log(`Preparing attachment index [${i}] called [${this.files[i].name}] for sending`)
            const hash = await MainClass.calculateSHA256(this.files[i])

            let exists = false
            if (existingHashes === null) {
                console.warn(`existingHashes is null, uploading attachment index [${i}]`)
                exists = false
            } else {
                for (let h = 0; h < existingHashes.length; h++) {
                    console.log(`Comparing [${hash}] with [${existingHashes[h]}]`)
                    if (MainClass.areArraysEqual(hash, existingHashes[h])) {
                        exists = true
                        break
                    }
                }
            }

            if (!exists) {
                console.log(`Attachment index [${i}] doesn't exist on server, uploading...`)
                formData.append('a', this.files[i])
            } else {
                console.log(`Attachment index [${i}] exists on server, sending hash only...`)
                const name = this.files[i].name
                const jsonString = JSON.stringify({Hash: hash, Name: name})
                formData.append('h', jsonString)
            }
        }

        const xhr = new XMLHttpRequest()

        return new Promise((resolve, reject) => {
            xhr.onload = () => {
                if (xhr.status === 200) {
                    const attachmentToken = JSON.parse(xhr.responseText)
                    console.log('Attachment was uploaded successfully')
                    this.resetAttachments()
                    this.calculateAttachments()
                    resolve(attachmentToken)
                } else {
                    console.error('Failed asking the server if given attachment hashes exist')
                    reject(null)
                }
            }

            xhr.onloadstart = function () {
                console.log('Starting upload...')
            }
            xhr.onloadend = function () {
                console.log('Finished upload')
            }

            xhr.upload.onprogress = async function (e) {
                console.log(e.loaded, e.total)
                if (e.lengthComputable) {
                    const indicator = document.getElementById('upload-percentage')
                    let percent = (e.loaded / e.total) * 100

                    percent = Math.round(percent)
                    indicator.textContent = percent.toString() + ' %'
                    if (percent >= 100) {
                        indicator.textContent = ''
                    }
                }
            }


            xhr.onerror = function () {
                console.error('Error asking the server if given attachment hashes exist')
                reject(null)
            }

            xhr.open('POST', '/upload-attachment')
            xhr.send(formData)
        })

    }

    static resetAttachments() {
        console.log('Resetting attachments')
        document.getElementById('attachment-list').innerHTML = ''
        this.files = []
    }

    static addAttachment(entry) {
        if (this.files.length >= this.maxFiles) {
            console.warn('Too many attachments, ignoring those after 4th...')
            return
        }
        this.files.push(entry)
        console.log(`Added attachment [${entry.name}], current attachment count: [${this.files.length}]`)

        const reader = new FileReader()
        reader.readAsDataURL(entry)

        // when the file is loaded into the browser
        reader.onload = (e) => {
            const attachmentList = document.getElementById('attachment-list')

            const attachmentContainer = document.createElement('div')
            attachmentList.appendChild(attachmentContainer)

            // when clicked on the attachment, it removes it
            attachmentContainer.addEventListener('click', () => {
                attachmentContainer.remove()
                this.removeAttachment(entry)
                if (attachmentList.length <= 0) {
                    this.AttachmentInput.value = ''
                }
                console.log(`Removed attachment [${entry.name}], current attachment count: [${this.files.length}]`)
                this.calculateAttachments()
            })

            const text = false

            const attachmentPreview = document.createElement('div')
            attachmentPreview.className = 'attachment-preview'
            if (text) {
                attachmentContainer.style.height = '224px'
            } else {
                attachmentContainer.style.height = '200px'
            }
            const imgElement = document.createElement('img')
            imgElement.src = e.target.result
            imgElement.style.display = 'block'
            attachmentPreview.appendChild(imgElement)
            attachmentContainer.appendChild(attachmentPreview)

            if (text) {
                const attachmentName = document.createElement('div')
                attachmentName.className = 'attachment-name'
                attachmentName.textContent = 'test.jpg'
                attachmentContainer.appendChild(attachmentName)
            }
            this.calculateAttachments()
        }
    }

    static removeAttachment(entry) {
        this.files.splice(this.files.indexOf(entry), 1)
    }

    static hideFileDropUI() {
        this.fileDropZone.style.display = 'none'
    }

    static calculateAttachments() {
        const attachmentList = document.getElementById('attachment-list')

        const count = attachmentList.children.length

        const ChatInputForm = document.getElementById('chat-input-form')

        if (count > 0 && attachmentList.style.display !== 'flex') {
            attachmentList.style.display = 'flex'
            ChatInputForm.style.borderTopLeftRadius = '0px'
            ChatInputForm.style.borderTopRightRadius = '0px'
            ChatInputForm.style.borderTopStyle = 'solid'
        } else if (count <= 0 && attachmentList.style.display === 'flex') {
            attachmentList.style.display = 'none'
            ChatInputForm.style.borderTopLeftRadius = '12px'
            ChatInputForm.style.borderTopRightRadius = '12px'
            ChatInputForm.style.borderTopStyle = 'none'
        }

        document.getElementById('chat-input').focus()
    }

    static setChatInputPlaceHolderText(channelName) {
        const chatInput = document.getElementById('chat-input')
        if (chatInput === null) {
            console.error(`Can't set channel name in chat input placeholder text because chat input isn't loaded`)
            return
        }
        chatInput.placeholder = `${Translation.get('message')} #${channelName}`
    }
}

class FriendListClass {
    static create() {
        document.getElementById('third-column-main').innerHTML = `
            <div id="friend-list-container">
                <div id="search-friend"></div>
                <label id="friend-count"></label>
                <ul id="friend-list"></ul>
            </div>`


        // document.getElementById('dm-friends-button').style.backgroundColor = ColorsClass.mainColor
        // this.friendListContainer.style.display = 'flex'
        this.addCurrentFriends()
        // ChatMessageListClass.disableChat()
        // ChannelListClass.selectNoChannel(true)
    }

    static disableFriendList() {
        this.friendListContainer.style.display = 'none'
        this.friendList.innerHTML = ''
    }

    static addCurrentFriends() {
        this.updateFriendCount()
        for (let f = 0; f < MainClass.myFriends.length; f++) {
            this.addFriend(MainClass.myFriends[f], "name", "username", '/content/static/default_profilepic.webp', true, 1, "test status")
        }
    }

    static addFriend(userID, displayName, username, picture, online, status, statusText) {
        const friendStr =
            `<li friend-id="${userID}">
                <div class="profile-pic-container" style="width: 32px; height: 32px">
                    <img class="profile-pic" src="${picture}">
                    <div class="user-status"></div>
                </div>   
                <div class="user-data">
                    <span class="user-name">${userID}</span>
                    <div class="user-status-text">${statusText}</div>
                </div>
            </li>`

        const friendList = document.getElementById('friend-list')

        friendList.insertAdjacentHTML('beforeend', friendStr)

        const friend = friendList.querySelector(`[friend-id="${userID}"]`)
        ContextMenuClass.registerContextMenu(friend, (pageX, pageY) => {
            ContextMenuClass.userCtxMenu(userID, pageX, pageY)
        })
    }

    static updateFriendCount() {
        document.getElementById('third-column-main').querySelector('label').textContent = `friends - ${MainClass.myFriends.length}`
    }
}