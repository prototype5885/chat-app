const ChatLoadingIndicator = document.getElementById("chat-loading-indicator")

let amountOfMessagesLoaded = 0
let lastReceivedChannelHistoryID = ""

const locale = navigator.language
const dateHourShort = { timeStyle: "short" }
const dateOptionsDay = { year: "numeric", month: "long", day: "numeric" }
const dateOptionsLong = { year: "numeric", month: "long", day: "numeric", hour: "2-digit", minute: "2-digit" }


function updateDaySeparatorsInChat() {
    // insert a separator that separates message each day
    const messages = Array.from(ChatMessagesList.querySelectorAll("li.msg"))

    // remove previous day separators
    const daySeparators = Array.from(ChatMessagesList.querySelectorAll("li.date-between-msgs"))
    daySeparators.forEach(daySeparator => {
        daySeparator.remove()
    })

    let lastDate = ""
    for (let i = 0; i < messages.length; i++) {
        // extract date from message id
        const date = new Date(Number((BigInt(messages[i].id) >> BigInt(22)))).toLocaleDateString(locale, dateOptionsDay)

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

            ChatMessagesList.insertBefore(dateBetweenMsgs, messages[i])
        }
        lastDate = date
    }
}

function removeGhostMessages() {
    const ghostMessages = ChatMessagesList.querySelectorAll(".ghost-msg")
    if (ghostMessages.length === 0) {
        return
    }
    for (let i = 0; i < ghostMessages.length; i++) {
        ghostMessages[i].remove()
    }
}

// adds the new chat message into html
function addChatMessage(messageID, userID, message, attachments, ghost) {
    if (document.getElementById(messageID) !== null) {
        console.error("A message already exists in chat list, won't add it again")
        return
    }

    // extract the message date from messageID
    let msgDate
    if (ghost) {
        msgDate = new Date()
    } else {
        msgDate = new Date(Number((BigInt(messageID) >> BigInt(22))))
    }

    let msgDateStr = ""

    const today = new Date()

    const yesterday = new Date()
    yesterday.setDate(yesterday.getDate() - 1)



    if (msgDate.toLocaleDateString() === today.toLocaleDateString()) {
        msgDateStr = translation.today + msgDate.toLocaleTimeString(locale, dateHourShort)
    } else if (msgDate.toLocaleDateString() === yesterday.toLocaleDateString()) {
        msgDateStr = "Yesterday at " + msgDate.toLocaleTimeString(locale, dateHourShort)
    } else {
        msgDateStr = msgDate.toLocaleString(locale, dateOptionsLong)
    }

    const userInfo = getUserInfo(userID)

    // create a <li> that holds the message
    const li = document.createElement("li")
    if (ghost) {
        li.className = "msg ghost-msg"
    } else {
        li.className = "msg"
    }
    li.id = messageID
    li.setAttribute("user-id", userID)

    var owner = false
    if (userID === ownUserID) {
        owner = true
    }

    registerContextMenu(li, (pageX, pageY) => { messageCtxMenu(messageID, owner, pageX, pageY) })

    // create a <img> that shows profile pic on the left
    const img = document.createElement("img")
    img.className = "msg-profile-pic"

    if (userInfo.pic !== "") {
        img.src = userInfo.pic
    } else {
        img.src = "discord.webp"
    }

    img.width = 40
    img.height = 40

    registerContextMenu(img, (pageX, pageY) => { userCtxMenu(userID, pageX, pageY) })

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

    registerContextMenu(msgNameDiv, (pageX, pageY) => { userCtxMenu(userID, pageX, pageY) })

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
        return `<a href="${url}" class="url" target="_blank">${url}</a>`
    })

    // append both name/date <div> and msg <div> to msgDatDiv
    msgDataDiv.appendChild(msgTextDiv)

    // append both the profile pic and message data to the <li>
    li.appendChild(img)
    li.appendChild(msgDataDiv)

    // insert the messages ordered by message id
    const messages = ChatMessagesList.querySelectorAll("li.msg")

    let inserted = false
    for (let i = 0; i < messages.length; i++) {
        if (li.id < messages[i].id) {
            ChatMessagesList.insertBefore(li, messages[i])
            inserted = true
            break
        }
    }

    if (!inserted) {
        ChatMessagesList.appendChild(li)
    }

    // add attachments
    if (attachments !== undefined && attachments !== null && attachments.length > 0) {
        for (let i = 0; i < attachments.length; i++) {
            const path = `/content/attachments/${attachments[i]}`
            const extension = attachments[i].split(".").pop().toLowerCase()

            const attachmentContainer = document.createElement("div")

            switch (extension) {
                case "mp3":
                    attachmentContainer.className = "message-attachment-audios"
                    attachmentContainer.innerHTML += `
                        <audio controls class="attachment-audio">
                            <source src="${path}">
                        </audio>`
                    msgDataDiv.appendChild(attachmentContainer)
                    break
                case "mp4":
                case "webm":
                case "mov":
                    attachmentContainer.className = "message-attachment-videos"

                    attachmentContainer.innerHTML += `
                        <video controls class="attachment-video">
                            <source src="${path}">
                        </video>`
                    msgDataDiv.appendChild(attachmentContainer)
                    break
                case "jpg":
                case "jpeg":
                case "webp":
                case "png":
                case "gif":
                    attachmentContainer.className = "message-attachment-pictures"
                    attachmentContainer.innerHTML += `<img src="${path}" class="attachment-pic">`
                    msgDataDiv.appendChild(attachmentContainer)
                    break
                default:
                    console.warn("Unsupported attachment type:", extension)
                    attachmentContainer.className = "attachment-unknown"
                    attachmentContainer.innerHTML += `<a href="${path}" class="url" target="_blank">${attachments[i]}</a>`
                    msgDataDiv.appendChild(attachmentContainer)
                    break
            }
        }
    }
}

function deleteChatMessage() {
    const messageID = json
    console.log(`Deleting message ID [${messageID}]`)
    document.getElementById(messageID).remove()
    amountOfMessagesChanged()
}

async function chatMessageReceived(json) {
    if (!channelHistoryReceived) {
        console.warn("Won't add received chat message as it was meant for the previous channel")
        return
    }
    if (!memberListLoaded) {
        await waitUntilBoolIsTrue(() => memberListLoaded) // wait until members are loaded
    }

    // console.log(`New chat message ID [${json.MsgID}] received`)
    addChatMessage(json.MsgID, json.UserID, json.Msg, json.Att, false)


    if (getScrollDistanceFromBottom(ChatMessagesList) < 200 || json.IDu === ownUserID) {
        ChatMessagesList.scrollTo({
            top: ChatMessagesList.scrollHeight,
            behavior: "smooth"
        })
    } else {
        console.log("Too far from current chat messages, not scrolling down on new message")
    }

    // play notification sound if messages is from other user
    if (json.UserID !== ownUserID) {
        if (Notification.permission === "granted") {
            sendNotification(json.UserID, json.Msg)
        } else {
            NotificationSound.play()
        }
    }
    amountOfMessagesChanged()
}


async function chatHistoryReceived(json) {
    console.log(`Requested chat history for channel ID [${json[0]}] arrived`)
    if (currentChannelID !== json[0]) {
        console.warn(`The received chat history was meant to be for channel ID [${json[0]}], but you are on [${currentChannelID}]`)
        return
    }
    // if (currentChannelID === lastReceivedChannelHistoryID) {
    //     console.warn("You already received the chat history for this channel")
    //     return
    // }
    if (!memberListLoaded) {
        await waitUntilBoolIsTrue(() => memberListLoaded) // wait until members are loaded
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
            addChatMessage(chatMessages[i].MessageID, chatMessages[i].UserID, chatMessages[i].Message, chatMessages[i].Attachments, false)
        }

        // only auto scroll down when entering channel, and not when
        // server sends rest of history while scrolling up manually
        if (!channelHistoryReceived) {
            // this runs when entered a channel
            ChatMessagesList.scrollTo({
                top: ChatMessagesList.scrollHeight,
                behavior: "instant"
            })
            // set this so it won't scroll down anymore as messages arrive while scrolling up
            // and won't request useless chat history requests when scrolling on top
            // if already reached the beginning
            lastChannelID = currentChannelID
        }
    } else {
        // run if server sent json that doesn't contain any more messages
        if (currentChannelID === lastChannelID) {
            // this can only run if already in channel
            console.warn("Reached the beginning of the chat, don't request more")
            // will become false upon entering an other channel
            reachedBeginningOfChannel = true
        } else {
            // and this only when entering a channel
            console.warn("Current channel has no chat history")
        }
    }
    setLoadingChatMessagesIndicator(false)
    amountOfMessagesChanged()
    setChannelHistoryReceived(true)
    lastReceivedChannelHistoryID = json[0]
}

function amountOfMessagesChanged() {
    amountOfMessagesLoaded = ChatMessagesList.querySelectorAll("li").length
    console.log("Amount of messages loaded:", amountOfMessagesLoaded)
    updateDaySeparatorsInChat()
    // removeGhostMessages()
}

function changeDisplayNameInChatMessageList(userID, newDisplayName) {
    const chatMessages = ChatMessagesList.querySelectorAll(".msg")
    chatMessages.forEach((chatMessage) => {
        if (chatMessage.getAttribute("user-id") === userID) {
            chatMessage.querySelector(".msg-user-name").textContent = newDisplayName
        }
    })
}

function changeProfilePicInChatMessageList(userID, pic) {
    const chatMessages = ChatMessagesList.querySelectorAll(".msg")
    chatMessages.forEach((chatMessage) => {
        if (chatMessage.getAttribute("user-id") === userID) {
            chatMessage.querySelector(".msg-profile-pic").src = pic
        }
    })
}

function scrolledOnChat() {
    if (!reachedBeginningOfChannel && ChatMessagesList.scrollTop < 200) {
        checkIfNeedsHistory()
    }
}

function checkIfNeedsHistory() {
    if (channelHistoryReceived && amountOfMessagesLoaded >= 50) {
        const chatMessage = ChatMessagesList.querySelector("li.msg")
        if (chatMessage != null) {
            requestChatHistory(currentChannelID, chatMessage.id)
            setLoadingChatMessagesIndicator(true)
        }
    }
}

function resetChatMessages() {
    // empties chat
    ChatMessagesList.innerHTML = ""

    // this makes sure there will be a little gap between chat input box
    // and the chat messages when user is viewing the latest message at the bottom
    ChatMessagesList.appendChild(document.createElement("div"))
}

function setLoadingChatMessagesIndicator(loading) {
    if (loading) {
        ChatLoadingIndicator.style.display = "flex"
        ChatMessagesList.style.overflowY = "hidden"
    } else {
        ChatLoadingIndicator.style.display = "none"
        ChatMessagesList.style.overflowY = ""
    }
}