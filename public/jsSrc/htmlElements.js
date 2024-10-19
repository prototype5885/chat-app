var AddChannelButton
var ServerList
var serverSeparators
var ChannelList
var MemberList
var ChatMessagesList
var UserPanelName
var AddServerButton
var UserSettingsButton
var ToggleMicrophoneButton
var ChatInput
var NotificationSound
var ChannelNameTop
var AboveFriendsChannels
var ServerNameButton
var ServerName

function initHtmlElements() {
    AddChannelButton = document.getElementById("add-channel-button")
    ServerList = document.getElementById("server-list")
    serverSeparators = ServerList.querySelectorAll(".servers-separator")
    ChannelList = document.getElementById("channel-list")
    MemberList = document.getElementById("member-list")
    ChatMessagesList = document.getElementById("chat-message-list")
    UserPanelName = document.getElementById("user-panel-name")
    AddServerButton = document.getElementById("add-server-button")
    UserSettingsButton = document.getElementById("user-settings-button")
    ToggleMicrophoneButton = document.getElementById("toggle-microphone-button")
    ChatInput = document.getElementById("chat-input")
    NotificationSound = document.getElementById("notification-sound")
    ChannelNameTop = document.getElementById("channel-name-top")
    AboveFriendsChannels = document.getElementById("above-friends-channels")
    ServerNameButton = document.getElementById("server-name-button")
    ServerName = document.getElementById("server-name")
}