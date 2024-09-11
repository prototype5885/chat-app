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

// handles sending message like pressing enter
// form.addEventListener('submit', (event) => {
//     event.preventDefault()
//     if (input.value) {
//         prepareMessage(input.value, 2002)
//         input.value = ''
//     }
// })


// runs whenever the chat input textarea content changes
const inputArea = document.getElementById('chat-input')
inputArea.addEventListener('input', () => {
    resizeChatInput()
})

// dynamically resize the chat input textarea to fit the text content
function resizeChatInput() {
    inputArea.style.height = 'auto'
    inputArea.style.height = inputArea.scrollHeight + 'px'
}

// send the text message on enter
inputArea.addEventListener('keydown', function (event) {
    // wont send if its shift enter so can make new lines
    if (event.key === 'Enter' && !event.shiftKey) {
        event.preventDefault()
        readChatInput()
    }
})

// send the text message on send button click
const sendButton = document.getElementById('send-button')
sendButton.addEventListener('click', () => {
    readChatInput()
});

// read the text message for sending
function readChatInput() {
    if (inputArea.value) {
        sendChatMessage(inputArea.value, 2002)
        inputArea.value = ''
        resizeChatInput()
    }
}

// create the right click menu on right click, delete existing one beforehand
document.addEventListener('contextmenu', function (event) {
    // event.preventDefault()
    deleteRightClickMenu()
    createRightClickMenu(event)
})

// delete the right click menu when clicking elsewhere
document.addEventListener('click', function () {
    deleteRightClickMenu()
})

// when clicking on add channel button
const addServerButton = document.getElementById('add-server-button')
addServerButton.addEventListener('click', () => {
    requestAddServer('test server')
});