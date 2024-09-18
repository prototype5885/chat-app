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

// when clicking on add server button
const addServerButton = document.getElementById('add-server-button')
addServerButton.addEventListener('click', () => {
    requestAddServer('test server')
})

// read the text message for sending
function readChatInput() {
    if (inputArea.value) {
        sendChatMessage(inputArea.value, 2002)
        inputArea.value = ''
        resizeChatInput()
    }
}

// when clicking on any server button on the left
function listenServerButtonsClick(button) {
    button.addEventListener('click', () => {
        console.log('Clicked:', button.id)
    })
}

var hidden = false
function toggleChannelsVisibility() {
    const list = document.getElementById('channels-list')
    const channels = Array.from(list.children)

    channels.forEach(channel => {
        if (!hidden) {
            if (channel.id != 2) {
                channel.style.display = 'none'
            }
        } else {
            channel.style.display = ''
        }
    })
    if (!hidden) {
        hidden = true
    } else {
        hidden = false
    }
}

function listenChannelButtonsClick(button) {
    button.addEventListener('click', () => {
        console.log('clicked:', button.id)
    })
}