const grey1 = "#949BA4"
const discordGray = "#36393f"
const discordBlue = "#5865F2"
const discrodGreen = "#00b700"

var currentServerID
var currentChannelID

var defaultRightClick = false

// runs whenever the chat input textarea content changes
ChatInput.addEventListener("input", () => {
    resizeChatInput()
})

// send the text message on enter
ChatInput.addEventListener("keydown", function (event) {
    // wont send if its shift enter so can make new lines
    if (event.key === "Enter" && !event.shiftKey) {
        event.preventDefault()
        readChatInput()
    }
})

// dynamically resize the chat input textarea to fit the text content
function resizeChatInput() {
    ChatInput.style.height = "auto"
    ChatInput.style.height = ChatInput.scrollHeight + "px"
}

function getUserInfo(userID) {
    const member = document.getElementById(userID)
    pic = member.querySelector('img.profile-pic').src
    username = member.querySelector('div.user-name').textContent
    return { username: username, pic: pic }
}

// delete context menu if left clicked somewhere thats not
// a context menu list element
document.addEventListener("click", function (event) {
    deleteCtxMenu()
})

// delete context menu if right clicked somewhere thats not registered
// with context menu listener
document.addEventListener("contextmenu", function (event) {
    if (!defaultRightClick) {
        event.preventDefault()
    }
    deleteCtxMenu()
})

function toggleMemberListView() {
    if (MemberList.style.display === "none") {
        showMemberList()
    } else {
        hideMemberList()
    }
}

function hideMemberList() {
    MemberList.style.display = "none"
}

function showMemberList() {
    MemberList.style.display = "block"
}

function serversSeparatorVisibility() {
    const servers = ServerList.querySelectorAll(".server, .placeholder-server")
    localStorage.setItem("serverCount", servers.length)



    if (servers.length != 0) {
        serverSeparators.forEach((separator) => {
            separator.style.display = "block"
        })
    } else {
        serverSeparators.forEach((separator) => {
            separator.style.display = "none"
        })
    }
}

// read the text message for sending
function readChatInput() {
    if (ChatInput.value) {
        sendChatMessage(ChatInput.value, currentChannelID)
        ChatInput.value = ""
        resizeChatInput()
    }
}

function registerClick(element, callback) {
    element.addEventListener("click", (event) => {
        deleteCtxMenu()
        event.stopPropagation()
        callback()
    })
}

function updateLastChannels() {
    const json = localStorage.getItem("lastChannels")

    let lastChannels = {}

    // first parse existing list, in case it exists in browser
    if (json != null) {
        lastChannels = JSON.parse(json)

        var serverIDstr = currentServerID.toString()
        var channelIDstr = currentChannelID.toString()

        if (serverIDstr in lastChannels && lastChannels[serverIDstr] === channelIDstr) {
            // if currentServerID and currentChannelID matches witht hose in lastChannels localStorage, don"t do anything
        }
    }
    // if channel was changed, overwrite with new one
    lastChannels[serverIDstr] = channelIDstr
    localStorage.setItem("lastChannels", JSON.stringify(lastChannels))
}

// selects the last selected channel after clicking on a server
function selectLastChannels(firstChannelID) {
    const json = localStorage.getItem("lastChannels")
    if (json != null) {
        let lastChannels = JSON.parse(json)
        const lastChannel = lastChannels[currentServerID.toString()]
        if (lastChannel != null) {
            selectChannel(lastChannel)
        } else {
            console.log("Current server does not have any last channel set in localStorage, selecting first channel...")
            selectChannel(firstChannelID)
        }
    } else {
        console.log("No lastChannels in localStorage exists, selecting first channel...")
        selectChannel(firstChannelID)
    }
}

// delete servers from lastChannels that no longer exist
function lookForDeletedServersInLastChannels() {
    const json = localStorage.getItem("lastChannels")
    if (json != null) {
        let lastChannels = JSON.parse(json)

        const li = ServerList.querySelectorAll(".server")

        const newLastChannels = {}
        li.forEach((li) => {
            const button = li.querySelector("button")
            const id = button.getAttribute("id")
            newLastChannels[id.toString()] = lastChannels[id.toString()]
        })

        if (JSON.stringify(lastChannels) === JSON.stringify(newLastChannels)) {
            console.log("All lastChannels servers in localStorage match")
        } else {
            // most likely one or more servers were deleted while user was offline
            console.warn("lastChannels servers in localStorage don't match with active servers")
            localStorage.setItem("lastChannels", JSON.stringify(newLastChannels))
        }
    } else {
        console.log("No lastChannels in localStorage exists")
    }
}

// delete a single server from lastChannels
function removeServerFromLastChannels(serverID) {
    const json = localStorage.getItem("lastChannels")
    if (json != null) {
        let lastChannels = JSON.parse(json)
        if (serverID.toString() in lastChannels) {
            delete lastChannels[serverID.toString()]
            localStorage.setItem("lastChannels", JSON.stringify(lastChannels))
            console.log(`Removed server ID ${serverID} from lastChannels`)
        }
        else {
            console.log(`Server ID ${serverID} doesn"t exist in lastChannels`)
        }
    }
}

// adds the new chat message into html
function addChatMessage(messageID, userID, message) {
    // extract the message date from messageID
    const msgDate = new Date(Number((BigInt(messageID) >> BigInt(22)))).toLocaleString()

    const userInfo = getUserInfo(userID)

    const chatNameColor = "#e7e7e7"
    // const pic = "default_profilepic.webp"
    // const username = userID.toString()

    // create a <li> that holds the message
    const li = document.createElement("li")
    li.className = "msg"
    li.id = messageID
    li.setAttribute("user-id", userID)

    var owner = false
    if (userID == ownUserID) {
        owner = true
    }

    registerRightClick(li, (pageX, pageY) => { messageCtxMenu(messageID, owner, pageX, pageY) })

    // create a <img> that shows profile pic on the left
    const img = document.createElement("img")
    // img.className = "msg-profile-pic"
    img.className = "msg-profile-pic"
    img.src = userInfo.pic
    img.alt = "pfpic"
    img.width = 40
    img.height = 40

    registerRightClick(img, (pageX, pageY) => { userCtxMenu(userID, pageX, pageY) })

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
    msgDataDiv.style.color = chatNameColor

    registerRightClick(msgNameDiv, (pageX, pageY) => { userCtxMenu(userID, pageX, pageY) })

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
    ChatMessagesList.appendChild(li)
}

function registerHoverListeners() {
    // add server button
    {
        registerHover(AddServerButton, () => { createbubble(AddServerButton, "Add Server", "right", 15) }, () => { deletebubble() })
        // hide notification marker as this doesn"t use it,
        // but its needed for formatting reasons
        AddServerButton.nextElementSibling.style.backgroundColor = "transparent"
    }
    // user settings button
    {
        registerHover(UserSettingsButton, () => { createbubble(UserSettingsButton, "User Settings", "up", 15) }, () => { deletebubble() })
    }
    // toggle microphone button
    {
        registerHover(ToggleMicrophoneButton, () => { createbubble(ToggleMicrophoneButton, "Toggle Microphone", "up", 15) }, () => { deletebubble() })
    }
    // add channel button
    {
        registerHover(AddChannelButton, () => { createbubble(AddChannelButton, "Create Channel", "up", 0) }, () => { deletebubble() })
    }
}

function registerClickListeners() {
    // add channel button
    {
        registerClick(AddChannelButton, () => { requestAddChannel() })
    }
}

function createPlaceHolderServers() {
    const serverCount = localStorage.getItem("serverCount")
    console.log('serverCount', serverCount)
    const placeholderButtons = []
    for (i = 0; i < serverCount; i++) {
        const buttonParent = addServer("", 0, "phs", "", "placeholder-server")
        let button = buttonParent.querySelector("button")
        button.nextElementSibling.style.backgroundColor = "transparent"
        button.textContent = ""
        placeholderButtons.push(buttonParent)
    }
    return placeholderButtons
}

function createContextMenu(actions, pageX, pageY) {
    if (actions.length == 0) {
        return
    }

    // create the right click menu
    const rightClickMenu = document.createElement("div")
    rightClickMenu.id = "right-click-menu"
    document.body.appendChild(rightClickMenu)

    // create ul that holds the menu items
    let ul = document.createElement("ul")
    rightClickMenu.appendChild(ul)

    // add a menu item for each action
    actions.forEach(function (action) {
        const li = document.createElement("li")
        li.textContent = action.text
        if (action.color === "red") {
            li.className = "cm-red" // to make the text red from css
        }
        // this will assing the function for each element
        li.onclick = function () {
            action.func()
        }

        ul.appendChild(li)
    })

    // creates the right click menu on cursor position
    rightClickMenu.style.display = "block"
    rightClickMenu.style.left = `${pageX}px`
    rightClickMenu.style.top = `${pageY}px`
}


function createbubble(element, text, direction, distance) {
    const content = document.createElement("div")
    content.textContent = text

    // create bubble div that will hold the content
    const bubble = document.createElement("div")
    bubble.id = "bubble"
    document.body.appendChild(bubble)

    // add the content into it
    bubble.appendChild(content)

    // center of the element that created the bubble
    // bubble will be created relative to this
    const rect = element.getBoundingClientRect()

    const center = {
        x: rect.left + rect.width / 2 + window.scrollX,
        y: rect.top + rect.height / 2 + window.scrollY
    }

    const height = bubble.getBoundingClientRect().height
    const width = bubble.getBoundingClientRect().width

    switch (direction) {
        case "right":
            // get how tall the bubble will be, so can
            // offset the Y position to make it appear
            // centered next to the element


            // set the bubble position
            bubble.style.left = `${(center.x + element.clientWidth / 2) + distance}px`
            bubble.style.top = `${center.y - height / 2}px`
            break
        case "up":

            bubble.style.left = `${center.x - width / 2}px`
            bubble.style.top = `${(center.y - element.clientHeight - (element.clientHeight / 2) - distance)}px`
            break
    }


}

function deletebubble() {
    const bubble = document.getElementById("bubble")
    if (bubble != null) {
        bubble.remove()
    } else {
        console.warn("A bubble was to be deleted but was nowhere to be found")
    }
}

function addMember(userID, displayName, picture, status) {
    // create a <li> that holds the user
    const li = document.createElement("li")
    li.className = "member"
    li.id = userID

    // create a <img> that shows profile pic on the left
    const img = document.createElement("img")
    img.className = "profile-pic"
    img.src = picture
    img.alt = "pfpic"
    img.width = 32
    img.height = 32

    // create a nested <div> that will contain username and status
    const userDataDiv = document.createElement("div")
    userDataDiv.className = "user-data"

    // create <div> that will hold the user"s message
    const userNameDiv = document.createElement("div")
    userNameDiv.className = "user-name"
    userNameDiv.textContent = displayName
    userNameDiv.style.color = grey1

    // now create a <div> under name that display statis
    const userStatusDiv = document.createElement("div")
    userStatusDiv.className = "user-status-text"
    userStatusDiv.textContent = status

    // append both name/date <div> and msg <div> to msgDatDiv
    userDataDiv.appendChild(userNameDiv)
    userDataDiv.appendChild(userStatusDiv)

    // append both the profile pic and message data to the <li>
    li.appendChild(img)
    li.appendChild(userDataDiv)

    // and finally append the message to the message list
    MemberList.appendChild(li)
}

function removeMember(userID) {
    const element = document.getElementById(userID)
    if (element.className === "member") {
        element.remove()
    } else {
        console.log(`Trying to remove member ID [${userID}] but the element is not member class: [${element.className}]`)
    }
}

function updateServerImage(button, picture, firstCharacter) {
    if (picture !== "") {
        button.style.backgroundImage = `url(${picture})`
    } else {
        button.textContent = firstCharacter.toUpperCase()
    }
}

function addServer(serverID, ownerID, serverName, picture, className) {
    // this li will hold the server and notification thing, which is the span
    const li = document.createElement("li")
    li.className = className
    ServerList.append(li)

    // create the server button itself
    const button = document.createElement("button")
    button.id = serverID

    li.append(button)

    // set picture of server
    updateServerImage(button, picture, serverName[0])

    const span = document.createElement("span")
    span.className = "server-notification"
    li.append(span)

    // bubble on hover
    function onHoverIn() {
        if (serverID != currentServerID) {
            button.style.borderRadius = "35%"
            span.style.height = "24px"
        }
        createbubble(button, serverID.toString(), "right", 15)
    }

    function onHoverOut() {
        if (serverID != currentServerID) {
            button.style.borderRadius = "50%"
            span.style.height = "8px"
        }
        deletebubble()
    }

    var owned = false
    if (ownerID == ownUserID) {
        owned = true
    }

    button.setAttribute("owned", owned.toString())

    registerClick(button, () => { selectServer(serverID) })
    registerRightClick(button, (pageX, pageY) => { serverCtxMenu(serverID, owned, pageX, pageY) })
    registerHover(button, () => { onHoverIn() }, () => { onHoverOut() })

    // this check needs to be made else adding placeholder servers will break serverCount value,
    // as it would reset the serverCount value while adding placeholders, as fix serverSeparatorVisibility
    // is ran manually only after creating each placeholder servers on startup
    if (className === "server") {
        serversSeparatorVisibility()
    }

    return li
}

function selectServer(serverID) {
    console.log("Selected server ID", serverID, "Requesting list of channels...")

    memberListLoaded = false

    const serverButton = document.getElementById(serverID)
    if (serverButton == null) {
        console.log("Previous server set in")
    }

    if (serverID == currentServerID) {
        console.log("Selected server is already the current one")
        return
    }

    // this will reset the previously selected server"s
    // notification"s white thing"s size
    const previousServerButton = document.getElementById(currentServerID)
    if (previousServerButton != null) {
        previousServerButton.nextElementSibling.style.height = "8px"
    }

    if (previousServerButton != null) {
        previousServerButton.style.borderRadius = "50%"
    }

    currentServerID = serverID


    serverButton.nextElementSibling.style.height = "36px"

    // hide add channel button if server isn't own
    if (serverButton.getAttribute("owned") == "true") {
        AddChannelButton.style.display = "block"
    } else {
        AddChannelButton.style.display = "none"
    }

    if (serverID == "2000") {
        hideMemberList()
    } else {
        showMemberList()
    }

    resetChannels()
    resetMessages()
    resetMemberList()

    requestChannelList()
    requestMemberList()
    localStorage.setItem("lastServer", serverID)
}

function deleteServer(serverID) {
    console.log("Deleting server ID:", serverID)
    // check if class is correct
    document.getElementById(serverID).parentNode.remove()
    serversSeparatorVisibility()
}

function addChannel(channelID, channelName) {
    const button = document.createElement("button")
    button.id = channelID

    const buttonName = document.createElement("div")
    buttonName.textContent = channelID.toString()

    button.appendChild(buttonName)

    ChannelList.appendChild(button)

    registerClick(button, () => { selectChannel(channelID) })
    registerRightClick(button, (pageX, pageY) => { channelCtxMenu(channelID, pageX, pageY) })
}

function selectChannel(channelID) {
    console.log("Selected channel ID:", channelID)

    if (channelID == currentChannelID) {
        console.log("Channel selected is already the current one")
        return
    }

    document.getElementById(channelID).style.backgroundColor = discordGray
    const previousChannel = document.getElementById(currentChannelID)
    if (previousChannel != null) {
        document.getElementById(currentChannelID).removeAttribute("style")
    }

    // sets the placeholder text in the area where you enter the chat message
    ChatInput.placeholder = `Message ${channelID}`

    currentChannelID = channelID

    resetMessages()
    updateLastChannels()
    requestChatHistory(channelID)
}

var channelsHidden = false
function toggleChannelsVisibility() {
    const channels = Array.from(ChannelList.children)

    channels.forEach(channel => {
        if (!channelsHidden) {
            if (channel.id != currentChannelID) {
                channel.style.display = "none"
            }
        } else {
            channel.style.display = ""
        }
    })
    if (!channelsHidden) {
        channelsHidden = true
    } else {
        channelsHidden = false
    }
}


function resetChannels() {
    ChannelList.innerHTML = ""
}

function resetMessages() {
    // empties chat
    ChatMessagesList.innerHTML = ""

    // this makes sure there will be a little gap between chat input box
    // and the chat messages when user is viewing the latest message
    const chatScrollGap = document.createElement("div")
    ChatMessagesList.appendChild(chatScrollGap)
}

function resetMemberList() {
    MemberList.innerHTML = ""
}

