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
            const date = new Date(Number((BigInt(messages[i].id) >> BigInt(22)))).toLocaleDateString(Translation.lang, this.dateOptionsDay)

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
    static addChatMessage(messageID, userID, message, attachments, edited, replyID, ghost) {
        if (document.getElementById(messageID) !== null) {
            console.error(`This message already exists in chat list with same ID, won't add it again: ${messageID}`)
            return
        }

        if (!ghost) {
            this.removeGhostMessages()
        }

        if (ChatInputClass.sentAChatMessage && userID === MainClass.getOwnUserID()) {
            ChatInputClass.sentAChatMessage = false
            ChatInputClass.enableChatInput()
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
            msgDateStr = Translation.get('today') + ' ' + msgDate.toLocaleTimeString(Translation.lang, this.dateHourShort)
        } else if (msgDate.toLocaleDateString() === yesterday.toLocaleDateString()) {
            msgDateStr = Translation.get('yesterday') + ' ' + msgDate.toLocaleTimeString(Translation.lang, this.dateHourShort)
        } else {
            msgDateStr = msgDate.toLocaleString(Translation.lang, this.dateOptionsLong)
        }

        // create a <li> that holds the message
        const li = document.createElement('li')
        if (ghost) {
            li.className = 'msg ghost-msg'
        } else {
            li.className = 'msg'
        }
        li.id = messageID
        li.setAttribute('user-id', userID)


        if (ghost) {
            li.style.opacity = 0.25
        }

        let owner = false
        if (userID === MainClass.getOwnUserID()) {
            owner = true
        }

        if (!ghost) {
            ContextMenuClass.registerContextMenu(li, (pageX, pageY) => {
                ContextMenuClass.messageCtxMenu(messageID, owner, pageX, pageY)
            })
        }

        if (replyID !== 0) {
            li.setAttribute('reply-id', replyID)
            const msgTop = document.createElement('div')
            msgTop.className = 'msg-top'
            li.appendChild(msgTop)
            msgTop.style.display = 'flex'

            msgTop.innerHTML = ` <svg width="52" height="16">
                                    <line x1="20" y1="16" x2="20" y2="8" stroke-width="2"/>
                                    <line x1="20" y1="8" x2="48" y2="8" stroke-width="2"/>
                                </svg>`

            const replyPic = document.createElement('img')
            replyPic.className = 'reply-msg-pic'
            const replyName = document.createElement('div')
            replyName.className = 'reply-msg-name'
            const replyMessage = document.createElement('div')
            replyMessage.className = 'reply-msg-message'
            msgTop.appendChild(replyPic)
            msgTop.appendChild(replyName)
            msgTop.appendChild(replyMessage)

            const msg = document.getElementById(replyID)
            if (msg === null) {
                console.log(`The message to which message ID [${messageID}] replied to no longer exists`)
            } else {
                const msgUserID = msg.getAttribute('user-id')
                const userInfo = MemberListClass.getUserInfo(userID)
                replyPic.src = userInfo.pic
                replyName.textContent = userInfo.displayName
                replyMessage.textContent = msg.querySelector('.msg-text').textContent

                if (msgUserID === MainClass.getOwnUserID() && userID !== MainClass.getOwnUserID()) {
                    li.style.backgroundColor = ColorsClass.replyColor
                    li.style.borderLeftColor = ColorsClass.replyColorBorder
                    li.style.borderLeftStyle = 'solid'
                    li.style.borderLeftWidth = '2px'
                }

                msgTop.addEventListener('click', () => {
                    msg.scrollIntoView({behavior: 'smooth', block: 'center'})
                    const originalColor = msg.style.backgroundColor
                    msg.style.backgroundColor = ColorsClass.highlightColor
                    setTimeout(() => {
                        msg.style.backgroundColor = originalColor
                    }, 2000)
                })
            }
        }

        const msgBottom = document.createElement('div')
        msgBottom.className = 'msg-bottom'

        li.appendChild(msgBottom)


        // create a <img> that shows profile pic on the left
        const img = document.createElement('img')
        img.className = 'msg-profile-pic'

        const memberInfo = MemberListClass.getUserInfo(userID)

        if (memberInfo.pic !== '') {
            img.src = memberInfo.pic
        } else {
            img.src = '/content/static/discord.webp'
        }

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

        if (!ghost) {
            ContextMenuClass.registerContextMenu(img, (pageX, pageY) => {
                ContextMenuClass.userCtxMenu(userID, pageX, pageY)
            })
        }


        // create a nested <div> that will contain sender name, message and date
        const msgDataDiv = document.createElement('div')
        msgDataDiv.className = 'msg-data'

        // inside that create a sub nested <div> that contains sender name and date
        const msgNameAndDateDiv = document.createElement('div')
        msgNameAndDateDiv.className = 'msg-name-and-date'

        // and inside that create a <div> that displays the sender's name on the left
        const msgNameDiv = document.createElement('div')
        msgNameDiv.className = 'msg-display-name'
        msgNameDiv.textContent = memberInfo.displayName

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

        msgDateShortContainer.textContent = `${msgDate.toLocaleTimeString(Translation.lang, this.dateHourShort)}`
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

        msgTextDiv.innerHTML = message.replace(/<@(\d+)>/g, (match, id) => {
            const userElement = document.getElementById(id)
            if (userElement) {
                const userName = userElement.querySelector('.display-name').textContent
                return `<span class="mention">@${userName}</span>`
            } else {
                return `<span class="mention">@${id}</span>`
            }
        })


        // append both name/date <div> and msg <div> to msgDatDiv
        msgTextContainer.appendChild(msgTextDiv)

        msgRightSide.appendChild(msgTextContainer)

        // append both the profile pic and message data to the <li>
        msgBottom.appendChild(img)
        msgBottom.appendChild(msgDataDiv)

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
                    case 'svg':
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
            li.style.paddingTop = '4px'
            li.style.paddingLeft = '16px'
            // msgDataDiv.style.height = '40px'
        }

        function short() {
            img.style.display = 'none'
            msgNameAndDateDiv.style.display = 'none'
            li.style.marginTop = '0px'
            li.style.paddingLeft = '0'
        }

        if (document.getElementById(messageID).hasAttribute('reply-id')) {
            normal()
            return
        }

        const previousElement = li.previousElementSibling
        if (previousElement === null) {
            return
        }

        if (previousElement.className === 'date-between-msgs') {
            normal()
            return
        }

        if (previousElement.className !== 'msg') {
            normal()
            return
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
        this.addChatMessage(json.MsgID, json.UserID, json.Msg, json.Att, json.Edited, json.RepID, false)

        // play notification sound if messages is from other user
        // if (json.UserID !== MainClass.getOwnUserID()) {

        if (json.Msg.includes(`<@${MainClass.myUserID}>`)) {
            NotificationClass.sendNotification(json.UserID, 'Mentioned you')
        }

        if (json.RepID !== 0) {
            const msg = document.getElementById(json.RepID)
            if (msg !== null) {
                const userID = msg.getAttribute('user-id')
                if (userID === MainClass.myUserID) {
                    NotificationClass.sendNotification(json.UserID, 'Replied to you')
                }
            }
        }


        // NotificationClass.sendNotification(json.UserID, json.Msg)
        // }

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
                        Attachments: json[1][u].Msgs[m][3], // attachments
                        ReplyID: json[1][u].Msgs[m][4], // replied to other message
                    }
                    chatMessages.push(chatMessage)
                }
            }
            // sort the history here because message history is not received ordered
            chatMessages.sort((a, b) => a.MessageID - b.MessageID)
            for (let i = 0; i < chatMessages.length; i++) {
                this.addChatMessage(chatMessages[i].MessageID, chatMessages[i].UserID, chatMessages[i].Message, chatMessages[i].Attachments, chatMessages[i].Edited, chatMessages[i].ReplyID, false)
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
                chatMessage.querySelector('.msg-display-name').textContent = newDisplayName
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
    static #replyMsgID = 0
    static maxFiles = 5
    static files = []
    static xhr
    static sentAChatMessage = false
    static #pressedButton = false

    static create() {
        document.getElementById('third-column-main').innerHTML += `
                <div id="chat-input-container">
                        <div id="mentionable-users-container">
                            <label>members</label>
                            <ul id="mentionable-user-list"></ul>
                        </div>
                        <div id="reply-container">
                            <span>${Translation.get('replyingTo')}</span>
                            <button onclick="ChatInputClass.closeReplyContainer()">
                                <svg width="32" height="32">
                                    <circle cx="16" cy="16" r="8" />
                                    <line x1="12" y1="20" x2="20" y2="12" stroke="#383a40" stroke-width="2"/>
                                    <line x1="12" y1="12" x2="20" y2="20" stroke="#383a40" stroke-width="2"/>
                                </svg>
                            </button>
                        </div>
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
            this.#pressedButton = true
            this.resizeChatInput()
            await this.checkIfTyping()
            MentionUserClass.lookForMentions(chatInput)
        })


        chatInput.addEventListener('selectionchange', async () => {
            // this pressed button is needed so input won't trigger selectionchange,
            // which would result lookForMentions to run twice
            if (!this.#pressedButton) {
                MentionUserClass.lookForMentions(chatInput)
            } else {
                this.#pressedButton = false
            }

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

        this.resizeChatInput()
        this.resetChatInput()
    }

    static openReplyContainer(msgID, userID) {
        this.#replyMsgID = msgID

        const replyContainer = document.getElementById('reply-container')
        replyContainer.style.display = 'flex'

        const userInfo = MemberListClass.getUserInfo(userID)

        replyContainer.querySelector('span').textContent = Translation.get('replyingTo') + ` ${userInfo.displayName}`

        this.setChatInputBorders()
    }

    static closeReplyContainer() {
        this.#replyMsgID = 0

        document.getElementById('reply-container').style.display = 'none'
        this.setChatInputBorders()
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
            if (document.getElementById('mentionable-user-list').innerHTML !== '') {
                return
            }

            if (this.sendingChatMsg) {
                console.warn(`Sending a message currently, can't send any new yet`)
                return
            }

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
                this.disableChatInput()

                const fakeSnowflake = ((BigInt(Date.now()) << BigInt(22)) | BigInt(0) << BigInt(12) | BigInt(0)).toString()

                //add ghost
                ChatMessageListClass.addChatMessage(fakeSnowflake, MainClass.getOwnUserID(), chatInput.value.trim(), null, false, 0, true)
                ChatMessageListClass.amountOfMessagesChanged()

                if (attachmentToken !== null) {
                    await WebsocketClass.sendChatMessage(chatInput.value.trim(), MainClass.getCurrentChannelID(), attachmentToken.AttToken, this.#replyMsgID)
                } else {
                    await WebsocketClass.sendChatMessage(chatInput.value.trim(), MainClass.getCurrentChannelID(), null, this.#replyMsgID)
                }
                console.log('Resetting chat input and attachment input values')
                chatInput.value = ''
                this.AttachmentInput.value = ''
                this.resizeChatInput()
                this.closeReplyContainer()
                this.#typing = false
                this.sentAChatMessage = true
            }
        }
    }

    static disableChatInput() {
        console.warn('Disabling chat input')
        this.sendingChatMsg = true
        document.getElementById('chat-input').style.opacity = 0.25
    }

    static enableChatInput() {
        console.warn('Enable chat input')
        this.sendingChatMsg = false
        document.getElementById('chat-input').focus()
        document.getElementById('chat-input').style.opacity = 1
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

        this.xhr = new XMLHttpRequest()

        return new Promise((resolve, reject) => {
            this.xhr.onload = () => {
                if (this.xhr.status === 200) {
                    const attachmentToken = JSON.parse(this.xhr.responseText)
                    console.log('Attachment was uploaded successfully')
                    this.resetChatInput()
                    resolve(attachmentToken)
                } else {
                    console.error('Failed asking the server if given attachment hashes exist')
                    reject(null)
                }
            }

            this.xhr.onloadstart = function () {
                console.log('Starting upload...')
            }
            this.xhr.onloadend = function () {
                console.log('Finished upload')
            }

            this.xhr.upload.onprogress = async function (e) {
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


            this.xhr.onerror = function () {
                console.error('Error asking the server if given attachment hashes exist')
                reject(null)
            }

            this.xhr.open('POST', '/upload-attachment')
            this.xhr.send(formData)
        })

    }

    static resetChatInput() {
        if (this.xhr !== undefined) {
            console.log('Aborting upload of attachments')
            this.xhr.abort()
            this.xhr = undefined
        }

        this.resetAttachments()
        this.setChatInputBorders()
    }

    static resetAttachments() {
        console.log('Resetting attachments')

        const attachmentList = document.getElementById('attachment-list')
        const replyContainer = document.getElementById('reply-container')
        if (replyContainer.style.display === 'flex') {
            attachmentList.style.borderTopLeftRadius = '0px'
            attachmentList.style.borderTopRightRadius = '0px'
        }

        attachmentList.innerHTML = ''
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
                this.setChatInputBorders()
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
            this.setChatInputBorders()
        }
    }

    static removeAttachment(entry) {
        this.files.splice(this.files.indexOf(entry), 1)
    }

    static setChatInputPlaceHolderText(channelName) {
        const chatInput = document.getElementById('chat-input')
        if (chatInput === null) {
            console.error(`Can't set channel name in chat input placeholder text because chat input isn't loaded`)
            return
        }
        chatInput.placeholder = `${Translation.get('message')} #${channelName}`
    }

    static setChatInputBorders() {
        const attachmentList = document.getElementById('attachment-list')
        const chatInputForm = document.getElementById('chat-input-form')
        const replyContainer = document.getElementById('reply-container')

        const attachmentCount = attachmentList.children.length

        if (attachmentCount > 0) {
            attachmentList.style.display = 'flex'
        } else if (attachmentCount <= 0) {
            attachmentList.style.display = 'none'
        }

        if (attachmentList.style.display === 'flex') {
            chatInputForm.style.borderTopStyle = 'solid'
            chatInputForm.style.borderTopLeftRadius = '0px'
            chatInputForm.style.borderTopRightRadius = '0px'

            if (replyContainer.style.display === 'flex') {
                attachmentList.style.borderTopLeftRadius = '0px'
                attachmentList.style.borderTopRightRadius = '0px'
            } else {
                attachmentList.style.borderTopLeftRadius = '12px'
                attachmentList.style.borderTopRightRadius = '12px'
            }
        } else {
            if (replyContainer.style.display === 'flex') {
                chatInputForm.style.borderTopLeftRadius = '0px'
                chatInputForm.style.borderTopRightRadius = '0px'
            } else {
                chatInputForm.style.borderTopLeftRadius = '12px'
                chatInputForm.style.borderTopRightRadius = '12px'
            }
            chatInputForm.style.borderTopStyle = 'none'
        }

        document.getElementById('chat-input').focus()
    }
}

class MentionUserClass {
    static word = ''
    static currentIndex = 0


    static init() {
        document.addEventListener('keydown', (event) => {
            const mentionableUserList = document.getElementById('mentionable-user-list')
            if (mentionableUserList === null || mentionableUserList.children.length === 0) {
                return
            }

            if (event.key === 'ArrowDown') {
                event.preventDefault()
                this.currentIndex = (this.currentIndex + 1) % mentionableUserList.children.length
                this.updateActiveItem()
            } else if (event.key === 'ArrowUp') {
                event.preventDefault()
                this.currentIndex = (this.currentIndex - 1 + mentionableUserList.children.length) % mentionableUserList.children.length
                this.updateActiveItem()
            } else if (event.key === 'Enter') {
                event.preventDefault()
                this.addMentionUser(mentionableUserList.children[this.currentIndex].getAttribute('user-id'))
                this.removeMentionableWindow()
            }
        })
    }

    static lookForMentions(chatInput) {
        if (chatInput.value === '') {
            this.removeMentionableWindow()
            return
        }

        let wordStartsAt = chatInput.selectionStart
        while (wordStartsAt > 0 && !/\s/.test(chatInput.value[wordStartsAt - 1])) {
            wordStartsAt--
        }

        let wordEndsAt = chatInput.selectionStart;
        while (wordEndsAt < chatInput.value.length && !/\s/.test(chatInput.value[wordEndsAt])) {
            wordEndsAt++
        }

        this.word = chatInput.value.substring(wordStartsAt, wordEndsAt)

        if (this.word.charAt(0) === '@') {
            const container = document.getElementById('mentionable-users-container')
            const mentionableUserList = document.getElementById('mentionable-user-list')

            mentionableUserList.innerHTML = ''

            this.currentIndex = 0
            container.style.display = 'flex'
            const mentionMember = this.word.substring(1)
            console.log('searching for user:', mentionMember)


            const members = document.getElementById('member-list').querySelectorAll('.member')

            function addMember(memberInfo, id) {
                const button = document.createElement('button')
                button.className = 'mentionable-user'
                button.setAttribute('user-id', id)

                const img = document.createElement('img')
                img.src = memberInfo.pic
                button.appendChild(img)

                const span = document.createElement('span')
                span.textContent = memberInfo.displayName
                button.appendChild(span)

                mentionableUserList.appendChild(button)

                button.addEventListener('click', () => {
                    MentionUserClass.addMentionUser(id)
                })

                // list.innerHTML += `
                // <button>
                //     <img src="${memberInfo.pic}" alt="">
                //     <span>${memberInfo.displayName}</span>
                // </button>`
            }

            if (mentionMember === '') {
                // const currentMsgIndex = chatMessageList.length - 1
                // for (let m = 0; m < 16; m++) {
                //
                // }

                const max = 10
                for (let i = 0; i < members.length; i++) {
                    if (i > max) {
                        break
                    }
                    if (members[i].getAttribute('online') === 'true') {
                        const memberInfo = MemberListClass.getUserInfo(members[i].id)
                        addMember(memberInfo, members[i].id)
                    }
                }
            } else {
                for (let i = 0; i < members.length; i++) {
                    const memberInfo = MemberListClass.getUserInfo(members[i].id)
                    if (memberInfo.displayName.toLowerCase().includes(mentionMember)) {
                        addMember(memberInfo, members[i].id)
                    }
                }

            }
            this.updateActiveItem()
        }
    }


    static updateActiveItem() {
        console.log(MentionUserClass.currentIndex)
        document.getElementById('mentionable-user-list').querySelectorAll('button').forEach((item, index) => {
            if (index === MentionUserClass.currentIndex) {
                item.style.backgroundColor = ColorsClass.selectedColor
            } else {
                item.style.backgroundColor = ''
            }
        })
    }

    static addMentionUser(id) {
        const chatInput = document.getElementById('chat-input')
        console.log(`replacing word [${this.word}] with [<@${id}>]`)
        chatInput.value = chatInput.value.replace(this.word, `<@${id}>`)
        document.getElementById('chat-input').focus()
        MentionUserClass.removeMentionableWindow()
    }

    static removeMentionableWindow() {
        this.word = ''
        this.currentIndex = 0
        document.getElementById('mentionable-users-container').style.display = 'none'
        document.getElementById('mentionable-user-list').innerHTML = ''
    }
}

class AttachmentInputClass {
    static fileDropZone = document.getElementById('file-drop-zone')
    static fileDropMsg = document.getElementById('file-drop-msg')

    static init() {
        document.addEventListener('dragover', e => {
            e.preventDefault()
            this.fileDropZone.style.display = 'flex'
            this.fileDropMsg.textContent = 'Upload to:\n\n' + MainClass.getCurrentChannelID()
        })

        this.fileDropZone.addEventListener('dragenter', e => {
            e.preventDefault()
            console.log('Started dragging a file into window')
        })

        // this when user drags files into webpage
        this.fileDropZone.addEventListener('drop', e => {
            e.preventDefault()
            console.log('dropped file')

            for (let i = 0; i < e.dataTransfer.items.length; i++) {
                if (e.dataTransfer.items[i].kind === 'file') {
                    const file = e.dataTransfer.items[i].getAsFile();
                    ChatInputClass.addAttachment(file)
                }
            }
            this.hideFileDropUI()
        })

        this.fileDropZone.addEventListener('dragleave', e => {
            e.preventDefault()
            console.log('Cancelled file dragging')
            this.hideFileDropUI()
        })

        document.addEventListener('paste', e => {
            const items = e.clipboardData.items
            if (items) {
                for (let i = 0; i < items.length; i++) {
                    const item = items[i]

                    if (item.kind === 'file') {
                        const file = item.getAsFile()
                        ChatInputClass.addAttachment(file)
                    }
                }
            }
        })
    }

    static hideFileDropUI() {
        this.fileDropZone.style.display = 'none'
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

    // static disableFriendList() {
    //     this.friendListContainer.style.display = 'none'
    //     this.friendList.innerHTML = ''
    // }

    static addCurrentFriends() {
        this.updateFriendCount()
        for (let f = 0; f < MainClass.myFriends.length; f++) {
            this.addFriend(MainClass.myFriends[f], "name", "username", '/static/default_profilepic.webp', true, 1, "test status")
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
                    <span class="display-name">${userID}</span>
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