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