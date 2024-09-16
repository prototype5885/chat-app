
const grey1 = '#949BA4'

// start -- right click menu
function createRightClickMenu(event) {
    let type
    let actions // this will hold the list of actions in the right click menu
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
            // return
    }

    // create the right click menu
    const rightClickMenu = document.createElement('div')
    rightClickMenu.className = 'right-click-menu'
    document.body.appendChild(rightClickMenu)

    // create ul that holds the menu items
    let ul = document.createElement('ul')
    rightClickMenu.appendChild(ul)

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
        const messageID = BigInt(event.target.closest('.msg').id) // find parent with msg class
        switch (action.class) {
            case 'delete-message':
                requestChatMessageDeletion(messageID)
                break
        }
    }
}

// adds the new chat message into html
function addChatMessage(messageID, channelID, userID, username, message) {
    // extract the message date from messageID
    const msgDate = new Date(Number((BigInt(messageID) >> BigInt(20)))).toLocaleString()

    const chatNameColor = '#e7e7e7'
    const pic = 'profilepic.jpg'

    const chatElement = 
    `<li class="msg" id="${messageID}" user-id="${userID}">
        <img class="msg-profile-pic" src="${pic}" width="40" height="40">
        <div class="msg-data">
            <div class="msg-name-and-date">
                <div class="msg-user-name" style="color: ${chatNameColor}">${username}</div>
                <div class="msg-date">${msgDate}</div>
            </div>
            <div class="msg-text">${message}</div>
        </div>
    </li>`

    document.getElementById('chat-message-list').insertAdjacentHTML('beforeend', chatElement)
}

function deleteChatMessage(messageID) {
    console.log('Deleting message id ' + messageID)
    document.getElementById(messageID).remove()
}

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

function addServer(serverID, serverName) {
    const button = document.createElement('button')
    button.className = 'server'
    button.id = serverID
    button.setAttribute('server-name', serverName)

    document.getElementById('server-list').append(button)

    listenServerButtonsClick(button)
}