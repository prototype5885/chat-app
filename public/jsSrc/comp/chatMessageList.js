class ChatMessageListClass {
    constructor(main, notification, chatInput) {
        this.main = main
        this.notification = notification
        this.chatInput = chatInput

        this.ChatLoadingIndicator = document.getElementById("chat-loading-indicator")
        this.ChatMessagesList = document.getElementById("chat-message-list")

        this.amountOfMessagesLoaded = 0
        this.lastReceivedChannelHistoryID = ""
        this.channelHistoryReceived = false

        this.locale = navigator.language
        this.dateHourShort = {timeStyle: "short"}
        this.dateOptionsDay = {year: "numeric", month: "long", day: "numeric"}
        this.dateOptionsLong = {year: "numeric", month: "long", day: "numeric", hour: "2-digit", minute: "2-digit"}

        this.ChatMessagesList.addEventListener("scroll", () => {
            if (!main.reachedBeginningOfChannel && this.ChatMessagesList.scrollTop < 200) {
                this.checkIfNeedsHistory()
            }
        })

        // const pictureViewer = document.getElementById("picture-viewer")
        // ContextMenuClass.registerContextMenu(pictureViewer, (pageX, pageY) => {
        //     ContextMenuClass.pictureCtxMenu(pictureViewer.src, pictureViewer.getAttribute("name"), pageX, pageY)
        // })
    }

    updateDaySeparatorsInChat() {
        // insert a separator that separates message each day
        const messages = Array.from(this.ChatMessagesList.querySelectorAll("li.msg"))

        // remove previous day separators
        const daySeparators = Array.from(this.ChatMessagesList.querySelectorAll("li.date-between-msgs"))
        daySeparators.forEach(daySeparator => {
            daySeparator.remove()
        })

        let lastDate = ""
        for (let i = 0; i < messages.length; i++) {
            // extract date from message id
            const date = new Date(Number((BigInt(messages[i].id) >> BigInt(22)))).toLocaleDateString(this.locale, this.dateOptionsDay)

            if (lastDate !== "" && lastDate !== date) {
                const dateBetweenMsgs = document.createElement("li")
                dateBetweenMsgs.className = "date-between-msgs"

                const leftLine = document.createElement("div")

                const dateText = document.createElement("span")
                dateText.textContent = date

                const rightLine = document.createElement("div")

                dateBetweenMsgs.appendChild(leftLine)
                dateBetweenMsgs.appendChild(dateText)
                dateBetweenMsgs.appendChild(rightLine)

                this.ChatMessagesList.insertBefore(dateBetweenMsgs, messages[i])
            }
            lastDate = date
        }
    }

    removeGhostMessages() {
        const ghostMessages = this.ChatMessagesList.querySelectorAll(".ghost-msg")
        if (ghostMessages.length === 0) {
            return
        }
        for (let i = 0; i < ghostMessages.length; i++) {
            ghostMessages[i].remove()
        }
    }

    // adds the new chat message into html
    addChatMessage(messageID, userID, message, attachments, ghost) {
        if (document.getElementById(messageID) !== null) {
            console.error("A message already exists in chat list, won't add it again")
            return
        }

        // extract the message date from messageID
        let msgDate
        if (ghost) {
            msgDate = new Date()
        } else {
            msgDate = MainClass.extractDateFromId(messageID)
        }

        let msgDateStr = ""

        const today = new Date()

        const yesterday = new Date()
        yesterday.setDate(yesterday.getDate() - 1)

        if (msgDate.toLocaleDateString() === today.toLocaleDateString()) {
            msgDateStr = Translation.translation.today + msgDate.toLocaleTimeString(this.locale, this.dateHourShort)
        } else if (msgDate.toLocaleDateString() === yesterday.toLocaleDateString()) {
            msgDateStr = "Yesterday at " + msgDate.toLocaleTimeString(this.locale, this.dateHourShort)
        } else {
            msgDateStr = msgDate.toLocaleString(this.locale, this.dateOptionsLong)
        }

        const userInfo = MemberListClass.getUserInfo(userID)

        // create a <li> that holds the message
        const li = document.createElement("li")
        if (ghost) {
            li.className = "msg ghost-msg"
        } else {
            li.className = "msg"
        }
        li.id = messageID
        li.setAttribute("user-id", userID)

        let owner = false
        if (userID === main.myUserID) {
            owner = true
        }

        ContextMenuClass.registerContextMenu(li, (pageX, pageY) => {
            ContextMenuClass.messageCtxMenu(messageID, owner, pageX, pageY)
        })

        // create a <img> that shows profile pic on the left
        const img = document.createElement("img")
        img.className = "msg-profile-pic"

        if (userInfo.pic !== "") {
            img.src = userInfo.pic
        } else {
            img.src = "/content/static/discord.webp"
        }

        img.width = 40
        img.height = 40


        MainClass.registerClick(img, () => {
            const pictureViewerContainer = document.getElementById("picture-viewer-container")
            const pictureViewer = document.getElementById("picture-viewer")

            pictureViewerContainer.style.display = "block"
            pictureViewer.src = img.src

            ContextMenuClass.registerContextMenu(pictureViewer, (pageX, pageY) => {
                ContextMenuClass.pictureCtxMenu(img.src, img.src, pageX, pageY)
            })

            MainClass.registerClick(pictureViewerContainer, () => {
                pictureViewerContainer.style.display = "none"
                pictureViewer.src = ""
            })
        })


        ContextMenuClass.registerContextMenu(img, (pageX, pageY) => {
            ContextMenuClass.userCtxMenu(userID, pageX, pageY)
        })

        // create a nested <div> that will contain sender name, message and date
        const msgDataDiv = document.createElement("div")
        msgDataDiv.className = "msg-data"

        // inside that create a sub nested <div> that contains sender name and date
        const msgNameAndDateDiv = document.createElement("div")
        msgNameAndDateDiv.className = "msg-name-and-date"

        // and inside that create a <div> that displays the sender"s name on the left
        const msgNameDiv = document.createElement("div")
        msgNameDiv.className = "msg-user-name"
        msgNameDiv.textContent = userInfo.username

        ContextMenuClass.registerContextMenu(msgNameDiv, (pageX, pageY) => {
            ContextMenuClass.userCtxMenu(userID, pageX, pageY)
        })

        // and next to it create a <div> that displays the date of msg on the right
        const msgDateDiv = document.createElement("div")
        msgDateDiv.className = "msg-date"
        msgDateDiv.textContent = msgDateStr

        // append name and date to msgNameAndDateDiv
        msgNameAndDateDiv.appendChild(msgNameDiv)
        msgNameAndDateDiv.appendChild(msgDateDiv)

        msgDataDiv.appendChild(msgNameAndDateDiv)

        // now create a <div> under name and date that displays the message
        const msgTextDiv = document.createElement("div")
        msgTextDiv.className = "msg-text"

        // look for URLs in the message and make them clickable
        msgTextDiv.innerHTML = message.replace(/https?:\/\/[^\s/$.?#].[^\s]*/g, (url) => {
            if (url.endsWith(".gif") || url.endsWith(".jpg") || url.endsWith(".jpeg") || url.endsWith(".png") || url.endsWith(".webp")) {
                return `<a href="${url}" class="url" target="_blank"><img src="${url}"></a>`
            } else {
                return `<a href="${url}" class="url" target="_blank">${url}</a>`
            }

        })

        // append both name/date <div> and msg <div> to msgDatDiv
        msgDataDiv.appendChild(msgTextDiv)

        // append both the profile pic and message data to the <li>
        li.appendChild(img)
        li.appendChild(msgDataDiv)

        // insert the messages ordered by message id
        const messages = this.ChatMessagesList.querySelectorAll("li.msg")

        let inserted = false
        for (let i = 0; i < messages.length; i++) {
            if (li.id < messages[i].id) {
                this.ChatMessagesList.insertBefore(li, messages[i])
                inserted = true
                break
            }
        }

        if (!inserted) {
            this.ChatMessagesList.appendChild(li)
        }

        // add attachments
        if (attachments !== undefined && attachments !== null && attachments.length > 0) {
            const attachmentContainer = document.createElement("div")
            for (let i = 0; i < attachments.length; i++) {
                const extension = attachments[i].Name.split(".").pop().toLowerCase()

                const hashHex = MainClass.base64toSha256(attachments[i].Hash)

                const path = `/content/attachments/${hashHex + "." + extension}`

                switch (extension) {
                    case "mp3":
                    case "wav":
                    case "ogg":
                    case "flac":
                        attachmentContainer.className = "message-attachment-audios"
                        attachmentContainer.innerHTML += `
                        <audio controls class="attachment-audio">
                            <source src="${path}">${attachments[i].Name}</source>
                        </audio>`
                        msgDataDiv.appendChild(attachmentContainer)
                        break
                    case "mp4":
                    case "webm":
                    case "mov":
                        attachmentContainer.className = "message-attachment-videos"

                        attachmentContainer.innerHTML += `
                        <video controls class="attachment-video">
                            <source src="${path}">${attachments[i].Name}</source>
                        </video>`
                        msgDataDiv.appendChild(attachmentContainer)
                        break
                    case "jpg":
                    case "jpeg":
                    case "webp":
                    case "png":
                    case "gif":
                    case "jfif":
                        attachmentContainer.className = "message-attachment-pictures"

                        const img = document.createElement("img")
                        img.src = path
                        img.className = "attachment-pic"
                        img.setAttribute("name", attachments[i].Name)
                        attachmentContainer.appendChild(img)


                        // attachmentContainer.innerHTML += `<img src="${path}" class="attachment-pic">`
                        msgDataDiv.appendChild(attachmentContainer)

                        ContextMenuClass.registerContextMenu(img, (pageX, pageY) => {
                            ContextMenuClass.pictureCtxMenu(path, attachments[i].Name, pageX, pageY)
                        })


                        MainClass.registerClick(img, () => {
                            const pictureViewerContainer = document.getElementById("picture-viewer-container")
                            const pictureViewer = document.getElementById("picture-viewer")

                            pictureViewerContainer.style.display = "block"
                            pictureViewer.src = path
                            console.log("clicked on:", path)

                            MainClass.registerClick(pictureViewerContainer, () => {
                                pictureViewerContainer.style.display = "none"
                                pictureViewer.src = ""
                            })
                        })
                        break
                    default:
                        console.warn("Unsupported attachment type:", extension)
                        attachmentContainer.className = "attachment-unknown"
                        attachmentContainer.innerHTML += `<a href="${path}" class="url" target="_blank">${attachments[i].Name}</a>`
                        msgDataDiv.appendChild(attachmentContainer)
                        break
                }
            }
        }
        this.checkPreviousMessage(messageID)
    }

    checkPreviousMessage(messageID) {
        const li = document.getElementById(messageID)
        const userID = li.getAttribute("user-id")
        const img = li.querySelector(".msg-profile-pic")
        const msgDataDiv = li.querySelector(".msg-data")
        const msgNameAndDateDiv = li.querySelector(".msg-name-and-date")

        function normal() {
            img.style.display = ""
            msgNameAndDateDiv.style.display = ""
            msgDataDiv.style.marginLeft = "14px"
            li.style.paddingTop = "16px"
        }

        function short() {
            img.style.display = "none"
            msgNameAndDateDiv.style.display = "none"
            msgDataDiv.style.marginLeft = "54px"
            li.style.paddingTop = "0px"
        }

        const previousElement = li.previousElementSibling
        if (previousElement === undefined) {
            return
        }

        if (previousElement.className === "date-between-msgs") {
            normal()
        }

        if (previousElement.className !== "msg") {
            normal()
        }


        // this is to make sure if previous message was before 00:00, the one after 00:00 will show up as full
        const messageDate = MainClass.extractDateFromId(messageID)
        const prevMessageDate = MainClass.extractDateFromId(previousElement.id)

        const messageUnixMin = Math.floor(messageDate.getTime() / 1000 / 60)
        const prevMessageUnixMin = Math.floor(prevMessageDate.getTime() / 1000 / 60)

        const elapsedMinutes = messageUnixMin - prevMessageUnixMin

        let sameUser = false
        if (previousElement.getAttribute("user-id") === userID) {
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

    deleteChatMessage(json) {
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

    async chatMessageReceived(json) {
        if (!this.channelHistoryReceived) {
            console.warn("Won't add received chat message as it was meant for the previous channel")
            return
        }
        if (!main.memberListLoaded) {
            await MainClass.waitUntilBoolIsTrue(() => main.memberListLoaded) // wait until members are loaded
        }

        // console.log(`New chat message ID [${json.MsgID}] received`)
        this.addChatMessage(json.MsgID, json.UserID, json.Msg, json.Att, false)


        if (MainClass.getScrollDistanceFromBottom(this.ChatMessagesList) < 200 || json.IDu === main.myUserID) {
            this.ChatMessagesList.scrollTo({
                top: this.ChatMessagesList.scrollHeight,
                behavior: "smooth"
            })
        } else {
            console.log("Too far from current chat messages, not scrolling down on new message")
        }

        // play notification sound if messages is from other user
        if (json.UserID !== main.myUserID) {
            if (Notification.permission === "granted") {
                this.notification.sendNotification(json.UserID, json.Msg)
            } else {
                this.notification.NotificationSound.play()
            }
        }
        this.amountOfMessagesChanged()
    }

    async chatHistoryReceived(json) {
        console.log(`Requested chat history for channel ID [${json[0]}] arrived`)
        if (main.currentChannelID !== json[0]) {
            console.warn(`The received chat history was meant to be for channel ID [${json[0]}], but you are on [${main.currentChannelID}]`)
            return
        }
        // if (main.currentChannelID === this.lastReceivedChannelHistoryID) {
        //     console.warn("You already received the chat history for this channel")
        //     return
        // }
        if (!main.memberListLoaded) {
            await MainClass.waitUntilBoolIsTrue(() => main.memberListLoaded) // wait until members are loaded
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
                        Attachments: json[1][u].Msgs[m][2] // attachments
                    }
                    chatMessages.push(chatMessage)
                }
            }
            // sort the history here because message history is not received ordered
            chatMessages.sort((a, b) => a.MessageID - b.MessageID)
            for (let i = 0; i < chatMessages.length; i++) {
                this.addChatMessage(chatMessages[i].MessageID, chatMessages[i].UserID, chatMessages[i].Message, chatMessages[i].Attachments, false)
            }

            // only auto scroll down when entering channel, and not when
            // server sends rest of history while scrolling up manually
            if (!this.channelHistoryReceived) {
                // this runs when entered a channel
                this.ChatMessagesList.scrollTo({
                    top: this.ChatMessagesList.scrollHeight,
                    behavior: "instant"
                })
                // set this so it won't scroll down anymore as messages arrive while scrolling up
                // and won't request useless chat history requests when scrolling on top
                // if already reached the beginning
                main.lastChannelID = main.currentChannelID
            }
        } else {
            // run if server sent json that doesn't contain any more messages
            if (main.currentChannelID === main.lastChannelID) {
                // this can only run if already in channel
                console.warn("Reached the beginning of the chat, don't request more")
                // will become false upon entering an other channel
                main.reachedBeginningOfChannel = true
            } else {
                // and this only when entering a channel
                console.warn("Current channel has no chat history")
            }
        }
        this.setLoadingChatMessagesIndicator(false)
        this.amountOfMessagesChanged()
        this.channelHistoryReceived = true
        this.lastReceivedChannelHistoryID = json[0]
    }

    amountOfMessagesChanged() {
        this.amountOfMessagesLoaded = this.ChatMessagesList.querySelectorAll("li").length
        console.log("Amount of messages loaded:", this.amountOfMessagesLoaded)
        this.updateDaySeparatorsInChat()
        // removeGhostMessages()
    }

    changeDisplayNameInChatMessageList(userID, newDisplayName) {
        const chatMessages = this.ChatMessagesList.querySelectorAll(".msg")
        chatMessages.forEach((chatMessage) => {
            if (chatMessage.getAttribute("user-id") === userID) {
                chatMessage.querySelector(".msg-user-name").textContent = newDisplayName
            }
        })
    }

    setChatMessageProfilePic(userID, pic) {
        const chatMessages = this.ChatMessagesList.querySelectorAll(".msg")
        chatMessages.forEach((chatMessage) => {
            if (chatMessage.getAttribute("user-id") === userID) {
                chatMessage.querySelector(".msg-profile-pic").src = pic
            }
        })
    }

    checkIfNeedsHistory() {
        if (this.channelHistoryReceived && this.amountOfMessagesLoaded >= 50) {
            const chatMessage = this.ChatMessagesList.querySelector("li.msg")
            if (chatMessage != null) {
                WebsocketClass.requestChatHistory(this.main.currentChannelID, chatMessage.id)
                this.setLoadingChatMessagesIndicator(true)
            }
        }
    }

    resetChatMessages() {
        // empties chat
        this.ChatMessagesList.innerHTML = ""

        // this makes sure there will be a little gap between chat input box
        // and the chat messages when user is viewing the latest message at the bottom
        this.ChatMessagesList.appendChild(document.createElement("div"))
    }

    setLoadingChatMessagesIndicator(loading) {
        if (loading) {
            this.ChatLoadingIndicator.style.display = "flex"
            this.ChatMessagesList.style.overflowY = "hidden"
        } else {
            this.ChatLoadingIndicator.style.display = "none"
            this.ChatMessagesList.style.overflowY = ""
        }
    }

    setMessagePic(messageID, pic) {
        const msgPic = document.getElementById(messageID).querySelector(".msg-profile-pic")
        if (pic !== "") {
            msgPic.src = pic
        } else {
            msgPic.src = "content/static/questionmark.svg"
        }
    }

    disableChat() {
        console.warn("Disabling chat")
        this.resetChatMessages()
        this.chatInput.disableChatInput()
        this.main.currentChannelID = "0"
    }

    enableChat() {
        this.chatInput.enableChatInput()
    }
}