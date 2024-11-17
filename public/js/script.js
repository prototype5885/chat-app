// main.js

const AddChannelButton = document.getElementById("add-channel-button")
const ServerList = document.getElementById("server-list")
const serverSeparators = ServerList.querySelectorAll(".servers-separator")
const ChannelList = document.getElementById("channel-list")
const MemberList = document.getElementById("member-list")
const ChatMessagesList = document.getElementById("chat-message-list")
const UserPanelName = document.getElementById("user-panel-name")
const AddServerButton = document.getElementById("add-server-button")
const UserSettingsButton = document.getElementById("user-settings-button")
const ToggleMicrophoneButton = document.getElementById("toggle-microphone-button")
const ChatInput = document.getElementById("chat-input")
const ChatInputForm = document.getElementById("chat-input-form")
const NotificationSound = document.getElementById("notification-sound")
const ChannelNameTop = document.getElementById("channel-name-top")
const AboveFriendsChannels = document.getElementById("above-friends-channels")
const ServerNameButton = document.getElementById("server-name-button")
const ServerName = document.getElementById("server-name")
const AttachmentInput = document.getElementById("attachment-input")
const AttachmentContainer = document.getElementById("attachment-container")
const AttachmentList = document.getElementById("attachment-list")

var ownUserID // this will be the first thing server will send
var receivedOwnUserID = false // don't continue loading until own user ID is received
var memberListLoaded = false // don't add chat history until server member list is received

var currentServerID
var currentChannelID
var lastChannelID
var reachedBeginning = false

function waitUntilBoolIsTrue(checkFunction, interval = 10) {
    return new Promise((resolve) => {
        const intervalId = setInterval(() => {
            if (checkFunction()) {
                clearInterval(intervalId)
                resolve()
            }
        }, interval)
    })
}

function main() {
    // this runs after webpage was loaded
    document.addEventListener("DOMContentLoaded", async function () {
        initNotification()
        initLocalStorage()
        initContextMenu()

        addServer("2000", 0, "Direct Messages", "hs.svg", "dm") // add the direct messages button

        // add place holder servers depending on how many servers the client was in, will delete on websocket connection
        // purely visual
        const placeholderButtons = createPlaceHolderServers()
        serversSeparatorVisibility()
        console.log("Placeholder buttons:", placeholderButtons.length)

        // this will continue when websocket connected
        await connectToWebsocket()

        // waits until server sends user"s own ID
        console.log("Waiting for server to send own user ID...")
        await waitUntilBoolIsTrue(() => receivedOwnUserID)

        const loading = document.getElementById("loading")
        const fadeOut = 0.25 //seconds

        setTimeout(() => {
            loading.remove() // Remove the element from the DOM
        }, fadeOut * 1000)

        loading.style.transition = `background-color ${fadeOut}s ease`
        loading.style.backgroundColor = "#00000000"
        loading.style.pointerEvents = "none"

        // remove placeholder servers
        for (let i = 0; i < placeholderButtons.length; i++) {
            placeholderButtons[i].remove()
        }

        requestServerList()

        registerHoverListeners() // add event listeners for hovering

        console
        selectServer("2000")
    })
}

function getScrollDistanceFromBottom(e) {
    return e.scrollHeight - e.scrollTop - e.clientHeight
}

function getScrollDistanceFromTop(e) {

}

main()

// notification.js

function initNotification() {
    if (Notification.permission !== "granted") {
        console.warn("Notifications are not enabled, requesting permission...")
        Notification.requestPermission()
    } else {
        console.log("Notifications are enabled")
    }
}

function sendNotification(userID, message) {
    const userInfo = getUserInfo(userID)
    if (Notification.permission === "granted") {
        new Notification(userInfo.username, {
            body: message,
            icon: userInfo.pic // Optional icon
        })
    }
}

// localStorage.js

var localStorageSupported = false

function initLocalStorage() {
    if (typeof (Storage) === "undefined") {
        console.log("Browser doesn't support storage")
    } else {
        console.log("Browser supports storage")
        localStorageSupported = true
    }
}

function getLastChannels() {
    return localStorage.getItem("lastChannels")
}

function setLastChannels(value) {
    localStorage.setItem("lastChannels", value)
}

function updateLastChannels() {
    if (!localStorageSupported) {
        console.warn("Local storage isn't enabled in browser, can't update lastChannels value")
        return
    }

    const json = getLastChannels()

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
    setLastChannels(JSON.stringify(lastChannels))
}

// delete servers from lastChannels that no longer exist
function lookForDeletedServersInLastChannels() {
    if (!localStorageSupported) {
        console.warn("Local storage isn't enabled in browser, can't look for deleted servers in lastChannels value")
        return
    }

    const json = getLastChannels()
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
            setLastChannels(JSON.stringify(newLastChannels))
        }
    } else {
        console.log("No lastChannels in localStorage exists")
    }
}

// delete a single server from lastChannels
function removeServerFromLastChannels(serverID) {
    if (!localStorageSupported) {
        console.warn(`Local storage isn't enabled in browser, can't delete server ID [${serverID}] from lastChannels value`)
        return
    }

    const json = getLastChannels()
    if (json != null) {
        let lastChannels = JSON.parse(json)
        if (serverID.toString() in lastChannels) {
            delete lastChannels[serverID.toString()]
            setLastChannels(JSON.stringify(lastChannels))
            console.log(`Removed server ID ${serverID} from lastChannels`)
        }
        else {
            console.log(`Server ID ${serverID} doesn"t exist in lastChannels`)
        }
    }
}

// selects the last selected channel after clicking on a server
function selectLastChannels(firstChannelID) {
    if (!localStorageSupported) {
        console.warn("Local storage isn't enabled in browser, can't select last used channel on server, selecting first channel instead")
        selectChannel(firstChannelID)
        return
    }

    const json = getLastChannels()
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

// lastServer 2.

// function getLastServer() {
//     return localStorage.getItem("lastServer")
// }

// function setLastServer(value) {
//     localStorage.setItem("lastServer", value)
// }

// serverCount 3.

function getServerCount() {
    if (!localStorageSupported) {
        console.warn(`Local storage isn't enabled in browser, can't get serverCount value, returning 0`)
        return 0
    } else {
        return localStorage.getItem("serverCount")
    }
}

function setServerCount(value) {
    if (!localStorageSupported) {
        console.warn(`Local storage isn't enabled in browser, can't set serverCount value`)
        return 0
    } else {
        localStorage.setItem("serverCount", value)
    }
}

// comp/httpRequests.js

async function sendPostRequest(url, struct) {
    const response = await fetch(window.location.origin + url, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(struct)
    })
    const result = await response.json()
    return result

    // console.log('Success:', result.Success)
    // console.log('Message:', result.Message)
}

// async function requestChannelList() {
//     const response = await fetch(`/channels/${currentChannelID}`);
//     const data = await response.text();
// }

// comp/contextMenu.js

var defaultRightClick = false

function initContextMenu() {
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
}

function registerContextMenu(element, callback) {
    element.addEventListener("contextmenu", (event) => {
        event.preventDefault()
        deleteCtxMenu()
        event.stopPropagation()
        callback(event.pageX, event.pageY)
    })
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

function deleteCtxMenu() {
    const rightClickmenu = document.getElementById("right-click-menu")
    if (rightClickmenu != null) {
        rightClickmenu.remove()
    }
}

function serverCtxMenu(serverID, owned, pageX, pageY) {
    const actions = []

    if (owned) { actions.push({ text: "Server Settings", func: () => addWindow("server-settings") }) }
    if (owned) { actions.push({ text: "Create Invite Link", func: () => requestInviteLink(serverID) }) }
    // if (owned) { actions.push({ text: "Delete Server", color: "red", func: () => requestDeleteServer(serverID) }) }
    if (!owned) { actions.push({ text: "Leave Server", color: "red", func: () => requestLeaveServer(serverID) }) }
    // if (!owned) { actions.push({ text: "Report Server", color: "red" }) }

    createContextMenu(actions, pageX, pageY)
}

function channelCtxMenu(channelID, pageX, pageY) {
    function renameChannel(channelID) {
        console.log("renaming channel", channelID)
    }

    function deleteChannel(channelID) {
        console.log("deleting channel", channelID)
    }

    const actions = [
        { text: "Rename channel", color: "", func: () => renameChannel(channelID) },
        { text: "Delete channel", color: "red", func: () => deleteChannel(channelID) }
    ]
    createContextMenu(actions, pageX, pageY)
}

function userCtxMenu(userID, pageX, pageY) {
    function addFriend(userID) {
        console.log("Adding friend", userID)
    }

    function reportUser(userID) {
        console.log("Reporting user", userID)
    }

    function removeFriend(userID) {
        console.log("Removing friend", userID)
    }

    function copyUserID(userID) {
        console.log("Copying user ID", userID)
        navigator.clipboard.writeText(userID)
    }

    const actions = [
        { text: "Add friend", func: () => addFriend(userID) },
        { text: "Report user", color: "red", func: () => reportUser(userID) },
        { text: "Remove friend", color: "red", func: () => removeFriend(userID) },
        { text: "Copy user ID", func: () => copyUserID(userID) }
    ]
    createContextMenu(actions, pageX, pageY)
}

function messageCtxMenu(messageID, owner, pageX, pageY) {
    function copyText() {
        const chatMsg = document.getElementById(messageID).querySelector(".msg-text").textContent
        console.log("Copied to clipboard:", chatMsg)
        navigator.clipboard.writeText(chatMsg)
    }

    const actions = []
    actions.push({ text: "Copy text", func: () => copyText() })
    if (owner) { actions.push({ text: "Delete message", color: "red", func: () => requestDeleteChatMessage(messageID) }) }
    if (!owner) { actions.push({ text: "Report message", color: "red" }) }
    createContextMenu(actions, pageX, pageY)
}

// comp/bubble.js

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

// comp/serverList.js

function createPlaceHolderServers() {
    const serverCount = getServerCount()
    const placeholderButtons = []
    if (serverCount !== 0) {
        for (i = 0; i < serverCount; i++) {
            const buttonParent = addServer("", 0, "phs", "", "placeholder-server")
            let button = buttonParent.querySelector("button")
            button.nextElementSibling.style.backgroundColor = "transparent"
            button.textContent = ""
            placeholderButtons.push(buttonParent)
        }
    }
    return placeholderButtons
}

function addServer(serverID, ownerID, serverName, picture, className) {
    // this li will hold the server and notification thing, which is the span
    const li = document.createElement("li")
    li.className = className
    ServerList.append(li)

    // create the server button itself
    const button = document.createElement("button")
    button.id = serverID
    button.setAttribute("name", serverName)

    li.append(button)

    // set picture of server
    if (picture !== "") {
        button.style.backgroundImage = `url(${picture})`
    } else {
        button.textContent = serverName[0].toUpperCase()
    }

    const span = document.createElement("span")
    span.className = "server-notification"
    li.append(span)

    // bubble on hover
    function onHoverIn() {
        if (serverID != currentServerID) {
            button.style.borderRadius = "35%"
            button.style.backgroundColor = "#5865F2"
            span.style.height = "24px"
        }
        createbubble(button, serverName, "right", 15)
    }

    function onHoverOut() {
        if (serverID != currentServerID) {
            button.style.borderRadius = "50%"
            button.style.backgroundColor = ""
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
    registerContextMenu(button, (pageX, pageY) => { serverCtxMenu(serverID, owned, pageX, pageY) })
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

    // this will reset the previously selected server's visuals
    const previousServerButton = document.getElementById(currentServerID)
    if (previousServerButton != null) {
        previousServerButton.nextElementSibling.style.height = "8px"
        previousServerButton.style.backgroundColor = ""
        previousServerButton.style.borderRadius = "50%"
    }

    currentServerID = serverID


    serverButton.nextElementSibling.style.height = "36px"

    const owned = serverButton.getAttribute("owned")

    // hide add channel button if server isn't own
    if (owned == "true") {
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
    resetChatMessages()
    resetMemberList()

    requestChannelList()
    requestMemberList()

    ServerName.textContent = serverButton.getAttribute("name")
}

function deleteServer(serverID) {
    console.log("Deleting server ID:", serverID)
    // check if class is correct
    document.getElementById(serverID).parentNode.remove()
    serversSeparatorVisibility()
}

function serversSeparatorVisibility() {
    const servers = ServerList.querySelectorAll(".server, .placeholder-server")
    setServerCount(servers.length)

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

function serverWhiteThingSize(thing, newSize) {
    thing.style.height = `${newSize}px`
}

// comp/memberList.js

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
    userNameDiv.style.color = grayTextColor

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

function getUserInfo(userID) {
    const member = document.getElementById(userID)
    if (member != null) {
        pic = member.querySelector('img.profile-pic').src
        username = member.querySelector('div.user-name').textContent
        return { username: username, pic: pic }
    } else {
        return { username: userID, pic: "" }
    }
}

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

function resetMemberList() {
    MemberList.innerHTML = ""
}

function changeDisplayNameInMemberList(userID, newDisplayName) {
    const user = document.getElementById(userID)
    user.querySelector(".user-name").textContent = newDisplayName
}

// comp/channelList.js

function addChannel(channelID, channelName) {
    const button = document.createElement("button")
    button.id = channelID

    const nameContainer = document.createElement("div")
    nameContainer.textContent = channelName

    button.appendChild(nameContainer)

    ChannelList.appendChild(button)

    registerClick(button, () => { selectChannel(channelID) })
    registerContextMenu(button, (pageX, pageY) => { channelCtxMenu(channelID, pageX, pageY) })
}

function selectChannel(channelID) {
    console.log("Selected channel ID:", channelID)
    reachedBeginning = false

    if (channelID == currentChannelID) {
        console.log("Channel selected is already the current one")
        return
    }

    const channelButton = document.getElementById(channelID)
    channelButton.style.backgroundColor = mainColor

    const previousChannel = document.getElementById(currentChannelID)
    if (previousChannel != null) {
        document.getElementById(currentChannelID).removeAttribute("style")
    }

    // sets the placeholder text in the area where you enter the chat message
    channelName = channelButton.querySelector("div").textContent
    ChatInput.placeholder = `Message ${channelName}`

    currentChannelID = channelID

    resetChatMessages()
    updateLastChannels()
    requestChatHistory(channelID, 0)
    ChannelNameTop.textContent = channelButton.querySelector("div").textContent
    // window.history.pushState(currentChannelID, currentChannelID, `/channel/${currentServerID}/${currentChannelID}`)
}

function toggleChannelsVisibility() {
    const channels = Array.from(ChannelList.children)

    channels.forEach(channel => {
        // check if channel is visible
        if (channel.style.display = "") {
            // hide if the channel isn't the current selected one
            if (channel.id != currentChannelID) {
                channel.style.display = "none"
            }
        } else {
            // make all channels visible
            channel.style.display = ""
        }
    })
}

function resetChannels() {
    ChannelList.innerHTML = ""
}

// comp/chatMessageList.js

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

// comp/window.js

var openWindows = [] // this stores every open windows as hashmap by type value
var lastSelected = new Map()

// this is called when something creates as new window
function addWindow(type) {
    openWindows.push(new Window(type))
}

class Window {
    constructor(type) {
        this.window
        this.topBar
        this.topBarLeft
        this.main
        this.type = type
        this.lastTop
        this.lastLeft
        this.lastWidth
        this.lastHeight
        this.maximized
        this.isDragging = false
        this.offsetX
        this.offsetY
        this.handleMouseDown = this.mouseDown.bind(this)
        this.handleMouseMove = this.mouseMove.bind(this)
        this.handleMouseUp = this.mouseUp.bind(this)
        this.handleSelectWindow = this.selectWindow.bind(this)
        this.createWindow()
        this.selectWindow()
    }

    deleteWindow() {
        // remove event listeners
        this.topBarLeft.removeEventListener("mousedown", this.handleMouseDown)
        this.window.removeEventListener("mousedown", this.handleSelectWindow)
        // remove html element before deleting from openWindows array
        this.window.remove()
        // remove from lastSelected

        // find and delete from array
        for (let i = 0; i < openWindows.length; i++) {
            if (openWindows[i] == this) {
                openWindows.splice(i, 1)
                lastSelected.delete(i)
            }
        }
    }

    maximizeWindow() {
        if (this.maximized) {
            this.window.style.top = this.lastTop
            this.window.style.left = this.lastLeft
            this.window.style.width = this.lastWidth
            this.window.style.height = this.lastHeight

            this.maximized = false

            this.makeActive()
        } else {
            this.lastTop = this.window.style.top
            this.lastLeft = this.window.style.left
            this.lastWidth = this.window.style.width
            this.lastHeight = this.window.style.height

            this.window.style.top = ""
            this.window.style.left = ""
            this.window.style.width = "100%"
            this.window.style.height = "100%"

            this.maximized = true

            this.topBar.style.backgroundColor = darkNonTransparent
            this.window.style.border = ""
        }
    }

    makeActive() {
        this.topBar.style.backgroundColor = darkTransparent
        this.window.style.border = "1px solid var(--dark-transparent)"
    }

    makeInactive() {
        this.topBar.style.backgroundColor = brighterTransparent
        this.window.style.border = "1px solid var(--brighter-transparent)"
    }

    // this runs when the top bar of window is held to move the window
    mouseDown(e) {
        document.addEventListener('mousemove', this.handleMouseMove)
        document.addEventListener('mouseup', this.handleMouseUp)
        e.preventDefault()
        this.isDragging = true
        this.offsetX = e.clientX - this.window.getBoundingClientRect().left
        this.offsetY = e.clientY - this.window.getBoundingClientRect().top
        this.topBarLeft.style.cursor = "grabbing"
    }

    mouseMove(e) {
        if (this.isDragging) {
            let newPosX = e.clientX - this.offsetX
            let newPosY = e.clientY - this.offsetY

            // clamn so it can leave the window
            newPosX = Math.max(0, Math.min(window.innerWidth - this.window.clientWidth, newPosX))
            newPosY = Math.max(0, Math.min(window.innerHeight - this.window.clientHeight, newPosY))

            this.window.style.left = `${newPosX}px`
            this.window.style.top = `${newPosY}px`
        }
    }

    mouseUp(e) {
        if (this.isDragging) {
            this.isDragging = false
            this.topBarLeft.style.cursor = ""
            // remove event listeners when stopped moving window
            document.removeEventListener("mousemove", this.handleMouseMove)
            document.removeEventListener("mouseup", this.handleMouseUp)
        }
    }

    // when window is clicked on, makes it selected window
    selectWindow() {
        // check if selected window is maximized, then don't select if it is
        if (this.maximized) {
            return
        }

        this.makeActive()

        // set order 0 for selected window
        for (let i = 0; i < openWindows.length; i++) {
            if (openWindows[i] == this) {
                lastSelected.set(i, 0)
                break
            }
        }

        // add + 1 for the order value of each other windows
        // also look for highest value among them
        let highestValue = 0
        for (let i = 0; i < openWindows.length; i++) {
            if (openWindows[i] != this) {
                const value = lastSelected.get(i) + 1
                lastSelected.set(i, value)
                if (value > highestValue) {
                    highestValue = value
                }
            }
        }

        // order the values here
        const orderedKeys = []
        for (let i = 0; i < highestValue + 1; i++) {
            for (const [key, value] of lastSelected.entries()) {
                if (value == i) {
                    orderedKeys.push(key)
                }
            }
        }
        // then trim the array
        // for example 0 1 6 8 would be 0 1 2 3
        for (let i = 0; i < orderedKeys.length; i++) {
            lastSelected.set(orderedKeys[i], i)
        }


        // set the z index for each window
        for (const [key, value] of lastSelected.entries()) {
            if (openWindows[key] != null) {
                openWindows[key].window.style.zIndex = (900 - value).toString()
                if (openWindows[key] != this) {
                    openWindows[key].makeInactive()
                }

            }
        }
    }

    createSettingsWindowArea() {
        const leftSide = document.createElement("div")
        leftSide.className = "settings-left"
        const rightSide = document.createElement("div")
        rightSide.className = "settings-right"

        this.main.appendChild(leftSide)
        this.main.appendChild(rightSide)
    }

    addElementsLeftSide(elements) {

    }

    createWindow() {
        // create main window div
        this.window = document.createElement("div")
        this.window.className = "window"
        this.window.setAttribute("type", this.type)

        const size = 50
        const topLeft = 50 / (100 / (100 - size))

        this.window.style.top = `${topLeft}%`
        this.window.style.left = `${topLeft}%`
        this.window.style.width = `${size}%`
        this.window.style.height = `${size}%`

        this.window.style.border = "1px solid var(--dark-transparent)"
        this.window.style.zIndex = "901"

        // this will be the top bar that holds title and exit buttons etc
        this.topBar = document.createElement("div")
        this.topBar.className = "window-top-bar"
        this.topBar.style.backgroundColor = darkTransparent
        this.window.appendChild(this.topBar)

        // the left part that holds title name
        this.topBarLeft = document.createElement("div")
        this.topBarLeft.className = "window-top-bar-left"
        this.topBar.appendChild(this.topBarLeft)

        // the right part that holds buttons
        const topBarRight = document.createElement("div")
        topBarRight.className = "window-top-bar-right"
        this.topBar.appendChild(topBarRight)

        // button that maximizes/returns to size
        const maximizeButton = document.createElement("button")
        maximizeButton.className = "window-maximize-button"
        topBarRight.appendChild(maximizeButton)

        // this is the main part under the top bar that holds content
        this.main = document.createElement("div")
        this.main.className = "window-main"
        this.window.appendChild(this.main)

        registerClick(maximizeButton, () => { this.maximizeWindow() })

        // button that closes the window
        const exitButton = document.createElement("button")
        exitButton.className = "window-exit-button"
        topBarRight.appendChild(exitButton)

        // register the exit button
        registerClick(exitButton, () => { this.deleteWindow() })

        // and finally add it to html
        document.body.appendChild(this.window)

        // add listeners for moving mouse and releasing mouse button
        this.topBarLeft.addEventListener('mousedown', this.handleMouseDown)

        // this listener makes it possible to select active window
        this.window.addEventListener("mousedown", this.handleSelectWindow)

        const leftSide = []

        switch (this.type) {
            case "user-settings":
                this.topBarLeft.textContent = "User settings"
                this.createSettingsWindowArea()
                this.addElementsLeftSide(["wtf", "XDDDD"])
                break
            case "server-settings":
                this.topBarLeft.textContent = "Server settings"
                this.createSettingsWindowArea()
                this.addElementsLeftSide(leftSide)
                break
        }

    }
}

// comp/chatInput.js

// dynamically resize the chat input textarea to fit the text content
// runs whenever the chat input textarea content changes
function resizeChatInput() {
    ChatInput.style.height = "auto"
    ChatInput.style.height = ChatInput.scrollHeight + "px"
}

// send the text message on enter
function sendChatEnter(event) {
    if (event.key === "Enter" && !event.shiftKey) {
        event.preventDefault()
        readChatInput()
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

function uploadAttachment() {
    AttachmentInput.click()


    // const response = await fetch(url, {
    //     method: 'POST',
    //     headers: {
    //         'Content-Type': 'application/json'
    //     },
    //     body: JSON.stringify(dataToSend)
    // })
    // const result = await response.json()
}

function attachmentAdded() {
    for (i = 0; i < AttachmentInput.files.length; i++) {
        const reader = new FileReader()
        reader.readAsDataURL(AttachmentInput.files[i]) // Read the file as a data URL

        reader.onload = function (e) {
            const attachmentContainer = document.createElement("div")
            AttachmentList.appendChild(attachmentContainer)

            attachmentContainer.addEventListener("click", function () {
                attachmentContainer.remove()
                calculateAttachments()
            })

            const text = false

            const attachmentPreview = document.createElement("div")
            attachmentPreview.className = "attachment-preview"
            if (text) {
                attachmentContainer.style.height = "224px"
            } else {
                attachmentContainer.style.height = "200px"
            }
            const imgElement = document.createElement("img")
            imgElement.src = e.target.result
            imgElement.style.display = 'block'
            attachmentPreview.appendChild(imgElement)
            attachmentContainer.appendChild(attachmentPreview)

            if (text) {
                // attachmentPreview.style.height = "224px"
                const attachmentName = document.createElement("div")
                attachmentName.className = "attachment-name"
                attachmentName.textContent = "test.jpg"
                attachmentContainer.appendChild(attachmentName)
            }
            calculateAttachments()
        }
    }

    // }
    // } else if (AttachmentInput.files.length == 0) {
    //     AttachmentPreviewContainer.style.display = "none"
    //     ChatInputForm.style.borderTopLeftRadius = "12px"
    //     ChatInputForm.style.borderTopRightRadius = "12px"
    //     ChatInputForm.style.borderTopStyle = "none"
    // }
}

function calculateAttachments() {
    const count = AttachmentList.children.length
    console.log("attachments:", count)

    if (count > 0 && AttachmentContainer.style.display != "block") {
        AttachmentContainer.style.display = "block"
        ChatInputForm.style.borderTopLeftRadius = "0px"
        ChatInputForm.style.borderTopRightRadius = "0px"
        ChatInputForm.style.borderTopStyle = "solid"
    } else if (count == 0 && AttachmentContainer.style.display == "block") {
        AttachmentContainer.style.display = "none"
        ChatInputForm.style.borderTopLeftRadius = "12px"
        ChatInputForm.style.borderTopRightRadius = "12px"
        ChatInputForm.style.borderTopStyle = "none"
    }
}

// dynamicContent.js

const mainColor = "#36393f"
const bitDarkerColor = "#2B2D31"
const darkColor = "#232428"
const darkerColor = "#1E1F22"
const grayTextColor = "#949BA4"
const darkTransparent = "#111214d1"
const darkNonTransparent = "#111214"
const brighterTransparent = "#656565d1"
const loadingColor = "#00000080"

const textColor = "#C5C7CB"

const blue = "#5865F2"
const green = "#00b700"

// function getChannelname(channelID) {
//     return document.getElementById(channelID).querySelector("div").textContent
// }

// function setChannelname(channelID, channelName) {
//     document.getElementById(channelID).querySelector("div").textContent = channelName
// }

function registerClick(element, callback) {
    element.addEventListener("click", (event) => {
        deleteCtxMenu()
        event.stopPropagation()
        callback()
    })
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

function registerHover(element, callbackIn, callbackOut) {
    element.addEventListener('mouseover', (event) => {
        // console.log('hovering over', element)
        callbackIn()
    })

    element.addEventListener('mouseout', (event) => {
        // console.log('hovering out', element)
        callbackOut()
    })
}

// websocket.js

var wsClient

async function connectToWebsocket() {
    // check if protocol is http or https
    if (location.protocol === "https:") {
        wsClient = new WebSocket("wss://" + window.location.host + "/wss")
    } else {
        wsClient = new WebSocket("ws://" + window.location.host + "/ws")
    }

    // make the websocket work with byte arrays
    wsClient.binaryType = "arraybuffer"

    var websocketConnected
    wsClient.onopen = async function (_event) {
        console.log("Connected to WebSocket successfully.")
        websocketConnected = true
    }

    // when server sends a message
    wsClient.onmessage = async function (event) {
        let receivedBytes = new Uint8Array(event.data)

        // convert the first 4 bytes into uint32 to get the endIndex,
        // which marks the end of the packet
        const reversedBytes = receivedBytes.slice(0, 4).reverse()
        const endIndex = new DataView(reversedBytes.buffer).getUint32(0)

        // 5th byte is a 1 byte number which states the type of the packet
        const packetType = receivedBytes[4]

        // get the json string from the 6th byte to the end
        const packetJson = String.fromCharCode.apply(null, receivedBytes.slice(5, endIndex))

        console.log("Received packet:", endIndex, packetType, packetJson)

        const json = JSON.parse(packetJson)
        switch (packetType) {
            case 0: // Server sent rejection message
                console.warn(json.Reason)
                break
            case 1: // Server sent a chat message
                chatMessageReceived(json)
                break
            case 2: // Server sent the requested chat history
                chatHistoryReceived(json)
                break
            case 3: // Server sent which message was deleted
                deleteChatMessage(json)
                break
            case 21: // Server responded to the add server request
                console.log("Add server request response arrived")
                addServer(json.ServerID, json.OwnerID, json.Name, json.Picture, "server")
                selectServer(json.ServerID)
                break
            case 22: // Server sent the requested server list
                console.log("Requested server list arrived")
                if (json != null) {
                    for (let i = 0; i < json.length; i++) {
                        console.log("Adding server ID", json[i].ServerID)
                        addServer(json[i].ServerID, json[i].OwnerID, json[i].Name, json[i].Picture, "server")
                    }
                } else {
                    console.log("Not being in any servers")
                }
                lookForDeletedServersInLastChannels()
                break
            case 23: // Server sent which server was deleted
                console.log(`Server ID [${json.ServerID}] has beend deleted`)
                const serverID = json.ServerID
                deleteServer(serverID)
                removeServerFromLastChannels(serverID)
                if (serverID == currentServerID) {
                    selectServer("2000")
                }
                break
            case 24: // Server sent the requested invite link to the chat server
                console.log("Requested invite link to the chat server arrived, adding to clipboard")
                const inviteID = json
                const inviteLink = `${window.location.protocol}//${window.location.host}/invite/${inviteID}`
                console.log(inviteLink)
                navigator.clipboard.writeText(inviteLink)
                break
            case 31: // Server responded to the add channel request
                console.log(`Adding new channel called [${json.Name}]`)
                addChannel(json.ChannelID, json.Name)
                break
            case 32: // Server sent the requested channel list
                console.log("Requested channel list arrived")
                if (json == null) {
                    console.warn("No channels on server ID", currentServerID)
                    break
                }
                for (let i = 0; i < json.length; i++) {
                    addChannel(json[i].ChannelID, json[i].Name)
                }
                selectLastChannels(json[0].ChannelID)
                break
            case 42: // Server sent the requested member list
                console.log("Requested member list arrived")
                if (json == null) {
                    console.warn("No members on server ID", currentServerID)
                    break
                }
                for (let i = 0; i < json.length; i++) {
                    addMember(json[i].UserID, json[i].Name, json[i].Picture, json[i].Status)
                }
                memberListLoaded = true
                break
            case 43: // Server sent user which user left a server
                if (json.UserID == ownUserID) {
                    console.log(`Left server ID [${json.ServerID}], deleting it from list`)
                    deleteServer(json.ServerID)
                    selectServer("2000")
                } else {
                    console.log(`User ID [${json.UserID}] left server ID [${json.ServerID}]`)
                    removeMember(json.UserID)
                }
                break
            case 51: // Server sent that a user changed display name
                if (userID == ownUserID) {
                    console.log("New display name:", json.newName)
                } else {
                    console.log(`User ID [${json.UserID}] changed their name to [${json.NewName}]`)
                }
                changeDisplayNameInChatMessageList(userID, newDisplayName)
                changeDisplayNameInMemberList(userID, newDisplayName)
                break

            case 241: // Server sent the client"s own user ID
                ownUserID = json
                console.log("Received own user ID:", ownUserID)
                UserPanelName.textContent = ownUserID
                receivedOwnUserID = true
                break
            default:
                console.log("Server sent unknown message type")
        }
    }
    await waitUntilBoolIsTrue(() => websocketConnected)
    return
}

class ReceivedChatMessage {
    constructor(messageID, userID, message) {
        this.messageID = messageID;
        this.userID = userID;
        this.message = message;
    }

    static fromJSON(jsonString) {
        const data = JSON.parse(jsonString);
        return new ReceivedChatMessage(data.IDm, data.IDu, this.Msg);
    }
}

function preparePacket(type, bigIntIDs, struct) {
    if (wsClient.readyState === WebSocket.OPEN) {
        // convert the type value into a single byte value that will be the packet type
        const typeByte = new Uint8Array([1])
        typeByte[0] = type

        let json = JSON.stringify(struct)

        // workaround to turn uint64 value in json from string to normal number value
        // since javascript cant serialize BigInt
        for (i = 0; i < bigIntIDs.length; i++) {
            if (bigIntIDs[i] != 0) {
                json = json.replace(`"${bigIntIDs[i]}"`, bigIntIDs[i])
            }
        }

        // serialize the struct into json then convert to byte array
        let jsonBytes
        if (struct != null) {
            jsonBytes = new TextEncoder().encode(json)
        } else {
            jsonBytes = new Uint8Array([0])
        }

        // convert the end index uint32 value into 4 bytes
        const endIndex = jsonBytes.length + 5
        const buffer = new ArrayBuffer(4)
        new DataView(buffer).setUint32(0, endIndex, true)
        const endIndexBytes = new Uint8Array(buffer)

        // merge them into a single packet
        const packet = new Uint8Array(4 + 1 + jsonBytes.length)
        packet.set(endIndexBytes, 0) // first 4 bytes will be the length
        packet.set(typeByte, 4) // 5. byte will be the packet type
        packet.set(jsonBytes, 5) // rest will be the json byte array

        console.log("Prepared packet:", endIndex, packet[4], json)

        wsClient.send(packet)
    }
    else {
        console.log("Websocket is not open")
    }
}

function sendChatMessage(message, channelID) { // type is 1
    console.log("Sending a chat message")
    preparePacket(1, [channelID], {
        ChannelID: channelID,
        Message: message
    })
}
function requestChatHistory(channelID, lastMessageID) {
    console.log("Requesting chat history for channel ID", channelID)
    preparePacket(2, [channelID, lastMessageID], {
        ChannelID: channelID,
        FromMessageID: lastMessageID,
        Older: true // if true it will request older, if false it will request newer messages from the message id
    })
}
function requestDeleteChatMessage(messageID) {
    console.log("Requesting to delete chat message ID", messageID)
    preparePacket(3, [messageID], {
        MessageID: messageID
    })
}
function requestAddServer(serverName) {
    console.log("Requesting to add a new server")
    preparePacket(21, [0], {
        Name: serverName
    })
}

function requestRenameServer(serverID) {
    console.log("Requesting to rename server ID:", serverID)
}

function requestDeleteServer(serverID) {
    if (document.getElementById(serverID).getAttribute("owned") == "false") return
    console.log("Requesting to delete server ID:", serverID)
    preparePacket(23, [serverID], {
        ServerID: serverID
    })
}

function requestInviteLink(serverID) {
    if (document.getElementById(serverID).getAttribute("owned") == "false") return
    console.log("Requesting invite link creation for server ID:", serverID)
    preparePacket(24, [serverID], {
        ServerID: serverID,
        SingleUse: false,
        Expiration: 7
    })
}

function requestServerList() {
    console.log("Requesting server list")
    preparePacket(22, [0], null)
}

function requestAddChannel() {
    if (document.getElementById(currentServerID).getAttribute("owned") == "false") return
    console.log("Requesting to add new channel to server ID:", currentServerID)
    preparePacket(31, [currentServerID], {
        Name: "Channel",
        ServerID: currentServerID
    })
}

function requestChannelList() {
    console.log("Requesting channel list for current server ID", currentServerID)
    preparePacket(32, [currentServerID], {
        ServerID: currentServerID
    })
}

function requestMemberList() {
    console.log("Requesting member list for current server ID", currentServerID)
    preparePacket(42, [currentServerID], {
        ServerID: currentServerID
    })
}

function requestLeaveServer(serverID) {
    console.log("Requesting to leave a server ID", serverID)
    preparePacket(43, [serverID], {
        ServerID: serverID
    })
}

function requestChangeDisplayName(newName) {
    console.log("Requesting to change display name to:", newName)
    preparePacket(51, [], {
        NewName: newName
    })
}

