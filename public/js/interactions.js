const memberList = document.getElementById('member-list')
const hideMemberListButton = document.getElementById('hide-member-list-button')

const form = document.getElementById('chat-input-form')
const input = document.getElementById('chat-input')

// hide member list
hideMemberListButton.addEventListener('click', function () {
    if (memberList.style.display === 'none') {
        memberList.style.display = 'flex'
    } else {
        memberList.style.display = 'none'
    }
})

// handles sending message like pressing enter
form.addEventListener('submit', (e) => {
    e.preventDefault()
    if (input.value) {
        if (input.value === 'quit') {
            console.log(wsClient)
            wsClient.close();
            console.log(wsClient)
        }
        const clientChatMsg = {
            ChanID: 1942,
            ChatMsg: input.value
        }
        sendPacket('clientChatMsg', JSON.stringify(clientChatMsg))
        input.value = ''
    }
})

// create the right click menu on right click
document.addEventListener('contextmenu', function (event) {
    event.preventDefault()
    deleteRightClickMenu()
    // console.log(event.target)
    createRightClickMenu(event)
})

// delete the right click menu when clicking elsewhere
document.addEventListener('click', function () {
    deleteRightClickMenu()
})