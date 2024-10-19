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
const AttachmentPreviewContainer = document.getElementById("attachment-preview-container")
const AttachmentPreview = document.getElementById("attachment-preview")

var ownUserID // this will be the first thing server will send
var receivedOwnUserID = false // don't continue loading until own user ID is received
var memberListLoaded = false // don't add chat history until server member list is received

var currentServerID
var currentChannelID

if (Notification.permission !== "granted") {
    console.warn("Notifications are not enabled, requesting permission...")
    Notification.requestPermission()
} else {
    console.log("Notifications are enabled")
}

// this runs after webpage was loaded
document.addEventListener("DOMContentLoaded", async function () {
    initLocalStorage()

    addServer("2000", 0, "Direct Messages", "hs.svg", "dm") // add the direct messages button

    // add place holder servers depending on how many servers the client was in, will delete on websocket connection
    // purely visual
    const placeholderButtons = createPlaceHolderServers()
    serversSeparatorVisibility()
    console.log("Placeholder buttons:", placeholderButtons.length)

    // this will continue when websocket connected
    console.log("1")
    await connectToWebsocket()
    console.log("2")

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