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
const Channels = document.getElementById("channels")
const FriendsChat = document.getElementById("friends-chat")
const FriendsChatList = document.getElementById("friends-chat-list")

let ownUserID = ""
let ownDisplayName = ""
let ownProfilePic = ""
let ownPronouns = ""
let ownStatusText = ""
let ownFriends = []
let ownBlocks = []

let receivedInitialUserData = false // don't continue loading until own user data is received
let receivedImageHostAddress = false // don't continue loading until host address of image server arrived
let memberListLoaded = false // don't add chat history until server member list is received

let currentServerID = 2000
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
    if (pic === "" || pic === undefined || pic == null) {
        return defaultProfilePic
    } else {
        return imageHost + "/content/avatars/" + pic
    }
}

function checkDisplayName(displayName) {
    if (displayName === "" || displayName === undefined || displayName === null) {
        return ""
    } else {
        return displayName
    }
}

function setMemberDisplayName(userID, displayName) {
    displayName = checkDisplayName(displayName)

    if (displayName === "") {
        displayName = userID
    }

    changeDisplayNameInMemberList(userID, displayName)
    changeDisplayNameInChatMessageList(userID, displayName)
}

function setMemberProfilePic(userID, pic) {
    pic = getAvatarFullPath(pic)
    changeProfilePicInMemberList(userID, pic)
    changeProfilePicInChatMessageList(userID, pic)
    console.log(`User ID [${userID}] changed profile pic to [${pic}]`)
}

function setOwnUserID(userID) {
    ownUserID = userID
    console.log(`Own user ID has been set to [${ownUserID}]`)
}

function setOwnDisplayName(displayName) {
    displayName = checkDisplayName(displayName)
    ownDisplayName = displayName

    if (displayName === "") {
        setUserPanelName(ownUserID)
    } else {
        setUserPanelName(displayName)
    }

    console.log(`Own display name has been set to [${ownDisplayName}]`)
}

function setOwnPronouns(pronouns) {
    ownPronouns = pronouns
    console.log(`Own pronouns have been set to [${ownPronouns}]`)
}

function setOwnStatusText(statusText) {
    ownStatusText = statusText
    console.log(`Own status text has been set to [${ownStatusText}]`)
}

function setOwnProfilePic(pic) {
    pic = getAvatarFullPath(pic)

    ownProfilePic = pic
    setUserPanelPic(pic)
    console.log(`Own profile pic has been set to [${ownProfilePic}]`)
}

function setOwnFriends(friends) {
    ownFriends = friends
    console.log(`You have [${ownFriends.length}] friends, they are: [${ownFriends}]`)
}

function removeFriend(userID) {
    for (i = 0; i < ownFriends.length; i++) {
        if (ownFriends[i] === userID) {
            ownFriends.splice(i, 1)
            return
        }
    }
    console.error(`Local error: could not remove user ID [${userID}] from ownFriends array`)
}

function setBlockedUsers(blocks) {
    ownBlocks = blocks
    console.log(`You have blocked [${ownBlocks.length}] users, they are: [${ownBlocks}]`)
}

function setCurrentChannel(channelID) {
    currentChannelID = channelID
    updateLastChannelsStorage()
}

function setButtonActive(button, active) {
    if (active) {
        button.classList.remove("noHover")
        button.disabled = false
    } else {
        button.classList.add("noHover")
        button.disabled = true
    }
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
        await waitUntilBoolIsTrue(() => receivedInitialUserData)

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