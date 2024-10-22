// adds the new chat message into html
function addChatMessage(messageID, userID, message, after) {
    // extract the message date from messageID
    const msgDate = new Date(Number((BigInt(messageID) >> BigInt(22)))).toLocaleString()

    const userInfo = getUserInfo(userID)

    // create a <li> that holds the message
    const li = document.createElement("li")
    li.className = "msg"
    li.id = messageID
    li.setAttribute("user-id", userID)

    var owner = false
    if (userID == ownUserID) {
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
    msgDateDiv.textContent = msgDate

    // append name and date to msgNameAndDateDiv
    msgNameAndDateDiv.appendChild(msgNameDiv)
    msgNameAndDateDiv.appendChild(msgDateDiv)

    // now create a <div> under name and date that displays the message
    const msgTextDiv = document.createElement("div")
    msgTextDiv.className = "msg-text"

    // look for URLs in the message and make them clickable
    msgTextDiv.innerHTML = message.replace(/https?:\/\/[^\s/$.?#].[^\s]*/g, (url) => {
        return `<a href="${url}" class="url" target="_blank">${url}</a>`
    })

    // append both name/date <div> and msg <div> to msgDatDiv
    msgDataDiv.appendChild(msgNameAndDateDiv)
    msgDataDiv.appendChild(msgTextDiv)

    // append both the profile pic and message data to the <li>
    li.appendChild(img)
    li.appendChild(msgDataDiv)

    // and finally append the message to the message list
    if (after) {
        ChatMessagesList.insertAdjacentElement("beforeend", li)
    } else {
        ChatMessagesList.insertAdjacentElement("afterbegin", li)
    }
}

function deleteChatMessage() {
    const messageID = json
    console.log(`Deleting message ID [${messageID}]`)
    document.getElementById(messageID).remove()
    amountOfmessagesChanged()
}

async function chatMessageReceived(json) {
    if (!memberListLoaded) {
        await waitUntilBoolIsTrue(() => memberListLoaded) // wait until members are loaded
    }

    console.log(`New chat message ID [${json.IDm}] received`)
    addChatMessage(json.IDm, json.IDu, json.Msg, true)

    if (getScrollDistanceFromBottom(ChatMessagesList) < 200 || json.IDu == ownUserID) {
        ChatMessagesList.scrollTo({
            top: ChatMessagesList.scrollHeight,
            behavior: "smooth"
        })
    } else {
        console.log("Too far from current chat messages, not scrolling down on new message")
    }

    if (json.IDu !== ownUserID) {
        if (Notification.permission === "granted") {
            sendNotification(json.IDu, json.Msg)
        } else {
            NotificationSound.play()
        }
    }
    amountOfmessagesChanged()
}

async function chatHistoryReceived(json) {
    console.log(`Requested chat history for current channel arrived`)
    if (!memberListLoaded) {
        await waitUntilBoolIsTrue(() => memberListLoaded) // wait until members are loaded
    }

    if (json !== null) {
        // runs if json contains chat history
        // loop through the json and add each messages one by one
        for (let i = 0; i < json.length; i++) {
            // false here means these messages will be inserted before existing ones
            addChatMessage(json[i].IDm, json[i].IDu, json[i].Msg, false)
        }
        // only auto scroll down when entering channel, and not when
        // server sends rest of history while scrolling up manually
        if (currentChannelID != lastChannelID) {
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
        if (currentChannelID == lastChannelID) {
            // this can only run if already in channel
            console.warn("Reached the beginning of the chat, don't request more")
            // will become false upon entering an other channel
            reachedBeginning = true
        } else {
            // and this only when entering a channel
            console.warn("Current channel has no chat history")
        }
    }
    amountOfmessagesChanged()
}

function amountOfmessagesChanged() {
    const count = ChatMessagesList.querySelectorAll("li").length
    console.log(count)
}

function changeDisplayNameInChatMessageList(userID, newDisplayName) {
    const chatMessages = ChatMessagesList.querySelectorAll(".msg")
    chatMessages.forEach((chatMessage) => {
        if (chatMessage.getAttribute("user-id") == userID) {
            chatMessage.querySelector(".msg-user-name").textContent = newDisplayName
        }
    })
}

var alreadyReached = false
function scrolledOnChat(event) {
    if (!alreadyReached && !reachedBeginning && ChatMessagesList.scrollTop < 200) {
        const chatmessage = ChatMessagesList.querySelector("li")
        if (chatmessage != null) {
            requestChatHistory(currentChannelID, chatmessage.id)
            alreadyReached = true
        }
    } else if (alreadyReached == true && ChatMessagesList.scrollTop > 200) {
        alreadyReached = false
    }
}

function resetChatMessages() {
    // empties chat
    ChatMessagesList.innerHTML = ""

    // this makes sure there will be a little gap between chat input box
    // and the chat messages when user is viewing the latest message
    ChatMessagesList.appendChild(document.createElement("div"))
}