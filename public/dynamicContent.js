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

var currentServerID
var currentChannelID

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
    if (member != null) {
        pic = member.querySelector('img.profile-pic').src
        username = member.querySelector('div.user-name').textContent
        return { username: username, pic: pic }
    } else {
        return { username: userID, pic: "" }
    }
}

// function getChannelname(channelID) {
//     return document.getElementById(channelID).querySelector("div").textContent
// }

// function setChannelname(channelID, channelName) {
//     document.getElementById(channelID).querySelector("div").textContent = channelName
// }

function changeDisplayName(userID, newDisplayName) {
    const user = document.getElementById(userID)
    const username = user.querySelector(".user-name")

    if (userID == ownUserID) { console.log("Old name:", username.textContent) }
    username.textContent = newDisplayName
    if (userID == ownUserID) { console.log("New name:", username.textContent) }

    changeDisplayNameInChatMessages()
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
    // {
    //     registerClick(AddChannelButton, () => { requestAddChannel() })
    // }
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

function serverWhiteThingSize(thing, newSize) {
    thing.style.height = `${newSize}px`
}