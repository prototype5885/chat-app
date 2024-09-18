// hide member list
const memberList = document.getElementById('member-list')
const hideMemberListButton = document.getElementById('hide-member-list-button')
hideMemberListButton.addEventListener('click', function () {
    if (memberList.style.display === 'none') {
        memberList.style.display = 'flex'
    } else {
        memberList.style.display = 'none'
    }
})

// runs whenever the chat input textarea content changes
const inputArea = document.getElementById('chat-input')
inputArea.addEventListener('input', () => {
    resizeChatInput()
})

// send the text message on enter
inputArea.addEventListener('keydown', function (event) {
    // wont send if its shift enter so can make new lines
    if (event.key === 'Enter' && !event.shiftKey) {
        event.preventDefault()
        readChatInput()
    }
})

// dynamically resize the chat input textarea to fit the text content
function resizeChatInput() {
    inputArea.style.height = 'auto'
    inputArea.style.height = inputArea.scrollHeight + 'px'
}

// create the right click menu on right click, delete existing one beforehand
document.addEventListener('contextmenu', function (event) {
    event.preventDefault()
    deleteRightClickMenu()
    createRightClickMenu(event)
})

// delete the right click menu when clicking elsewhere
document.addEventListener('click', function () {
    deleteRightClickMenu()
})

// read the text message for sending
function readChatInput() {
    if (inputArea.value) {
        sendChatMessage(inputArea.value, getCurrentChannelID())
        inputArea.value = ''
        resizeChatInput()
    }
}

// when clicked on a server from server list
function selectServer(serverID) {
    console.log('Clicked on server:', serverID)
    resetChannels()
    resetMessages()
    setCurrentServerID(serverID)
    requestChannelList()
}

// when clicked on a channel from channel list
function selectChannel(channelID) {
    console.log('Clicked on channel:', channelID)
    const previousChannelID = getCurrentChannelID()

    if (channelID == previousChannelID) {
        console.log('Channel clicked on is already the current one')
        return
    }

    resetMessages()
    setCurrentChannelID(channelID)
    requestChatHistory(channelID)

    setSelectedChannelBackground(channelID, previousChannelID)
}