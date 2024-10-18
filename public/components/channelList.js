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
    requestChatHistory(channelID)
    ChannelNameTop.textContent = channelButton.querySelector("div").textContent
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