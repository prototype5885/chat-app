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
    console.log('Global context menu on:', event.target.getAttribute('on-context-menu'))

    // this inner function is like a macro because it's needed multiple times later
    function getUserContextMenuActions(userID) {
        return [
            { text: 'Add friend', func: () => addFriend(userID)},
            { text: 'Report user', color: 'red', func: () => reportUser(userID) },
            { text: 'Remove friend', color: 'red', func: () => removeFriend(userID) },
            { text: 'Copy user ID', func: () => copyUserID(userID) }
        ]
    }

    const value = event.target.getAttribute('on-context-menu')

    if (value == 'msgUser') {
        const userID = BigInt(event.target.closest('.msg').getAttribute('user-id'))
        actions = getUserContextMenuActions(userID)
        createRightClickMenu(actions, event)
    } else if (value == 'user') {
        const userID = BigInt(event.target.closest('.user').getAttribute('user-id'))
        actions = getUserContextMenuActions(userID)
        createRightClickMenu(actions, event)
    } else if (value == 'message') {
        const messageID = BigInt(event.target.closest('.msg').id)
        actions = [
            { text: 'Delete message', color: 'red', func: () => requestChatMessageDeletion(messageID)}
        ]
        createRightClickMenu(actions, event)
    } else {
        actions = [
            { text: 'Action 1' },
            { text: 'Action 2', color: 'red' },
            { text: 'Action 3' }
        ]
        createRightClickMenu(actions, event)
    }
})

// delete the right click menu when clicking elsewhere
document.addEventListener('click', function () {
    deleteRightClickMenu()

    console.log('Global clicked on', event.target)
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

function renameServer(serverID) {
    console.log('renaming server', serverID)
}

function deleteServer(serverID) {
    console.log('deleting server', serverID)
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

function renameChannel(channelID) {
    console.log('renaming channel', channelID)
}

function deleteChannel(channelID) {
    console.log('deleting channel', channelID)
}

function addFriend(userID) {
    console.log('Adding friend', userID)
}

function reportUser(userID) {
    console.log('Reporting user', userID)
}

function removeFriend(userID) {
    console.log('Removing friend', userID)
} 

function copyUserID(userID) {
    console.log('Copying user ID', userID)
    navigator.clipboard.writeText(userID)
}