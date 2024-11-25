const mainColor = "#36393f"
const hoverColor = "#313338"
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