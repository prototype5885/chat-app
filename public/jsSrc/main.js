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
const ChatLoadingIndicator = document.getElementById("chat-loading-indicator")
const loading = document.getElementById("loading")

var ownUserID // this will be the first thing server will send
var receivedOwnUserID = false // don't continue loading until own user ID is received
var memberListLoaded = false // don't add chat history until server member list is received

var currentServerID
var currentChannelID
var lastChannelID
var reachedBeginningOfChannel = false

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

function fadeOutLoading() {
    setTimeout(() => {
        loading.style.display = "none"
    }, 250)

    loading.style.pointerEvents = "none"
    loading.style.opacity = "0%"
}

function fadeInLoading() {
    loading.style.display = "block"
    loading.style.opacity = "100%"
    loading.style.pointerEvents = "auto"
    loading.innerText = "Reconnecting..."
}

function refreshWebsocketContent() {
    document.querySelectorAll('.server').forEach(server => {
        server.remove();
    })

    requestServerList()
    selectServer("2000")
}

function main() {
    // this runs after webpage was loaded
    document.addEventListener("DOMContentLoaded", async function () {
        initNotification()
        initLocalStorage()
        initContextMenu()

        addServer("2000", 0, "Direct Messages", "hs.svg", "dm") // add the direct messages button

        // add placeholder servers depending on how many servers the client was in, will delete on websocket connection
        // purely visual
        const placeholderButtons = createPlaceHolderServers()
        serversSeparatorVisibility()
        console.log("Placeholder buttons:", placeholderButtons.length)

        // this will continue when websocket connected
        await connectToWebsocket()

        // waits until server sends user's own ID
        console.log("Waiting for server to send own user ID...")
        await waitUntilBoolIsTrue(() => receivedOwnUserID)


        // fadeOutLoading()

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