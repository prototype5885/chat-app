const AddChannelButton = document.getElementById("add-channel-button")
const ServerList = document.getElementById("server-list")
const serverSeparators = ServerList.querySelectorAll(".servers-separator")
const ChannelList = document.getElementById("channel-list")
const MemberList = document.getElementById("member-list")
const ChatMessagesList = document.getElementById("chat-message-list")
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

let ownUserID // this will be the first thing server will send
let ownDisplayName // and this too
let ownProfilePic
let ownPronouns
let ownStatusText;
let receivedOwnUserData = false // don't continue loading until own user ID is received
let receivedImageHostAddress = false // don't continue loading until host address of image server arrived
let memberListLoaded = false // don't add chat history until server member list is received

let currentServerID
let currentChannelID
let lastChannelID
let reachedBeginningOfChannel = false

// let imageHost = "http://localhost:8000/"
let imageHost = ""
const defaultProfilePic = "/content/static/default_profilepic.webp"

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

function getAvatarFullPath(pic) {
    return imageHost + "/content/avatars/" + pic
}

function setProfilePic(userID, pic) {
    if (pic === "") {
        pic = defaultProfilePic
    } else {
        pic = getAvatarFullPath(pic)
    }

    if (userID === ownUserID) {
        ownProfilePic = pic
        setUserPanelPic(pic)
    }

    changeProfilePicInMemberList(userID, pic)
}

function setDisplayName(userID, name) {
    if (name === "") {
        name = userID
    }

    if (userID === ownUserID) {
        ownDisplayName = name
        setUserPanelName(name)
    }
    changeDisplayNameInMemberList(userID, name)
    changeDisplayNameInChatMessageList(userID, name)
}

function fixJson(jsonString) {
    const valueNames = ["ChannelID", "UserID", "MessageID", "ServerID", "IDm", "IDu"]
    for (let i = 0; i < valueNames.length; i++) {
        jsonString = jsonString.replace(new RegExp(`"${valueNames[i]}":(\\d+)`, 'g'), (match, p1) => `"${valueNames[i]}":"${p1}"`)
    }
    return JSON.parse(jsonString)
}

function main() {
    // this runs after webpage was loaded
    document.addEventListener("DOMContentLoaded", async function () {
        initNotification()
        initLocalStorage()
        initContextMenu()

        addServer("2000", 0, "Direct Messages", "content/static/hs.svg", "dm") // add the direct messages button

        // add placeholder servers depending on how many servers the client was in, will delete on websocket connection
        // purely visual
        const placeholderButtons = createPlaceHolderServers()
        serversSeparatorVisibility()
        console.log("Placeholder buttons:", placeholderButtons.length)

        // this will continue when websocket connected
        await connectToWebsocket()

        // waits until server sends user's own ID and display name
        console.log("Waiting for server to send own user ID and display name...")
        await waitUntilBoolIsTrue(() => receivedOwnUserData)

        // request http address of image hosting server
        requestImageHostAddress()

        // wait until the address is received
        console.log("Waiting for server to send image host address..")
        await waitUntilBoolIsTrue(() => receivedImageHostAddress)

        // remove placeholder servers
        for (let i = 0; i < placeholderButtons.length; i++) {
            placeholderButtons[i].remove()
        }

        registerHoverListeners() // add event listeners for hovering

        refreshWebsocketContent()
    })
}

function getScrollDistanceFromBottom(e) {
    return e.scrollHeight - e.scrollTop - e.clientHeight
}

function getScrollDistanceFromTop(e) {

}

main()