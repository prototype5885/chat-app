const chatNameColor = '#e7e7e7'
const grey1 = '#949BA4'

// start -- right click menu
function createRightClickMenu(event) {
    // create the right click menu
    const rightClickMenu = document.createElement('div')
    rightClickMenu.className = 'right-click-menu'
    document.body.appendChild(rightClickMenu)

    // create ul that holds the menu items
    let ul = document.createElement('ul')
    rightClickMenu.appendChild(ul)

    let type
    let actions
    switch (event.target.className) {
        // if right-clicking on user from a friend/member list
        case 'user':
        case 'user-name':
        case 'profile-pic':
            actions = addUserRightClickActions()
            type = 'user'
            break
        // if right-clicking on user in chat
        case 'msg-user-name':
        case 'msg-profile-pic':
            actions = addUserRightClickActions()
            type = 'msg-user'
            break
        // if right-clicking on message
        case 'msg':
        case 'msg-text':
        case 'msg-date':
            actions = [
                { text: 'Delete message', class: 'delete-message' }
            ]
            type = 'message'
            break
        default:
            actions = [
                { text: 'Action 1', class: 'none' },
                { text: 'Action 2', class: 'none' },
                { text: 'Action 3', class: 'none' }
            ]
            break
    }

    // add a menu item for each action
    actions.forEach(function (action) {
        const li = document.createElement('li')
        li.className = action.class
        // if (action.color != '') {
        //     li.style.color = action.color
        // }
        li.textContent = action.text
        li.onclick = function () {
            rightClickMenuItemPressed(action, event, type)
        }
        ul.appendChild(li)
    })

    // creates the right click menu on cursor position
    rightClickMenu.style.display = 'block'
    rightClickMenu.style.left = event.pageX + 'px'
    rightClickMenu.style.top = event.pageY + 'px'
}

function deleteRightClickMenu() {
    let children = Array.from(document.body.children)
    children.forEach(function (child) {
        if (child.classList.contains('right-click-menu')) {
            child.parentNode.removeChild(child)
        }
    })
}

function addUserRightClickActions() {
    return [
        { text: 'Add friend', class: 'add-friend' },
        { text: 'Report user', class: 'report-user' },
        { text: 'Remove friend', class: 'remove-friend' },
        { text: 'Copy user ID', class: 'copy-userid' }
    ]
}

function rightClickMenuItemPressed(action, event, type) { // when something in right click menu was pressed
    deleteRightClickMenu()

    if (type === 'user' || type === 'msg-user') {
        let parent
        if (type === 'user') {
            parent = event.target.closest('.user')
        }
        else {
            parent = event.target.closest('.msg')
        }
        const userID = parent.getAttribute('user-id')
        switch (action.class) {
            case 'add-friend':
                console.log('Adding friend ' + userID)
                break
            case 'report-user':
                console.log('Reporting user ' + userID)
                break
            case 'remove-friend':
                console.log('Removing friend ' + userID)
                break
            case 'copy-userid':
                console.log('Copied user ID is: ' + userID)
                navigator.clipboard.writeText(userID);
                break
        }
    }
    else if (type === 'message') {
        const messageID = event.target.closest('.msg').id // find parent with msg class
        switch (action.class) {
            case 'delete-message':
                // messages.removeChild(messageID) // delete it
                deleteChatMessage(messageID)
                break
        }
    }
}
// end -- right click menu
// start -- chat message
function addChatMessage(messageDataJson) {
    let messageData = {}
    messageData.MsgID = undefined
    messageData.ChanID = undefined
    messageData.UserID = undefined
    messageData.Name = undefined
    messageData.Msg = undefined
    messageData = JSON.parse(messageDataJson)

    // create a <li> that holds the message
    const li = document.createElement('li')
    li.className = 'msg'
    li.id = messageData.MsgID
    // li.setAttribute('msg_id', messageData.msgID)

    li.setAttribute('user-id', BigInt(messageData.UserID))

    // create a <img> that shows profile pic on the left
    const img = document.createElement('img')
    img.className = 'msg-profile-pic'
    img.src = 'profilepic.jpg'
    img.alt = 'pfpic'
    img.width = 40
    img.height = 40

    // create a nested <div> that will contain sender name, message and date
    const msgDataDiv = document.createElement('div')
    msgDataDiv.className = 'msg-data'

    // inside that create a sub nested <div> that contains sender name and date
    const msgNameAndDateDiv = document.createElement('div')
    msgNameAndDateDiv.className = 'msg-name-and-date'

    // and inside that create a <div> that displays the sender's name on the left
    const msgNameDiv = document.createElement('div')
    msgNameDiv.className = 'msg-user-name'
    msgNameDiv.textContent = messageData.Name
    msgDataDiv.style.color = chatNameColor

    // and next to it create a <div> that displays the date of msg on the right
    const msgDateDiv = document.createElement('div')
    msgDateDiv.className = 'msg-date'
    msgDateDiv.textContent = new Date().toLocaleString()

    // append name and date to msgNameAndDateDiv
    msgNameAndDateDiv.appendChild(msgNameDiv)
    msgNameAndDateDiv.appendChild(msgDateDiv)

    // now create a <div> under name and date that displays the message
    const msgTextDiv = document.createElement('div')
    msgTextDiv.className = 'msg-text'
    msgTextDiv.textContent = messageData.Msg

    // append both name/date <div> and msg <div> to msgDatDiv
    msgDataDiv.appendChild(msgNameAndDateDiv)
    msgDataDiv.appendChild(msgTextDiv)

    // append both the profile pic and message data to the <li>
    li.appendChild(img)
    li.appendChild(msgDataDiv)

    // and finally append the message to the message list
    messages.appendChild(li)
}

function deleteChatMessage(messageID) {
    document.getElementById(messageID).remove()
    console.log('Deleting message id ' + messageID)
}
// end -- chat message
// start -- add member to group chat member list or friend list
function addMember(id, where) {
    // create a <li> that holds the user
    const li = document.createElement('li')
    li.className = 'user'
    li.id = "5"
    li.setAttribute('user-id', "5")

    // create a <img> that shows profile pic on the left
    const img = document.createElement('img')
    img.className = 'profile-pic'
    img.src = 'profilepic.jpg'
    img.alt = 'pfpic'
    img.width = 32
    img.height = 32

    // create a nested <div> that will contain username and status
    const userDataDiv = document.createElement('div')
    userDataDiv.className = 'user-data'

    // create <div> that will hold the user's message
    const userNameDiv = document.createElement('div')
    userNameDiv.className = 'user-name'
    userNameDiv.textContent = 'APFSDS'
    userNameDiv.style.color = grey1

    // now create a <div> under name that display statis
    const userStatusDiv = document.createElement('div')
    userStatusDiv.className = 'user-status-text'
    userStatusDiv.textContent = 'status'

    // append both name/date <div> and msg <div> to msgDatDiv
    userDataDiv.appendChild(userNameDiv)
    userDataDiv.appendChild(userStatusDiv)

    // append both the profile pic and message data to the <li>
    li.appendChild(img)
    li.appendChild(userDataDiv)

    // and finally append the message to the message list
    // if (where == '')
    memberList.appendChild(li)
}
// end -- add member to group chat member list or friend list
