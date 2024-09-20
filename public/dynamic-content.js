
const grey1 = '#949BA4'
var lastRightClickMenu

// start -- right click menu
function createRightClickMenu(actions, event) {
    event.preventDefault()
    deleteRightClickMenu()

    // create the right click menu
    const rightClickMenu = document.createElement('div')
    rightClickMenu.id = 'right-click-menu'
    // rightClickMenu.id = 'wtf'
    document.body.appendChild(rightClickMenu)

    // create ul that holds the menu items
    let ul = document.createElement('ul')
    rightClickMenu.appendChild(ul)

    lastRightClickMenu = rightClickMenu

    // add a menu item for each action
    actions.forEach(function (action) {
        const li = document.createElement('li')
        li.textContent = action.text
        if (action.color === 'red') {
            li.className = 'cm-red'
        }
        //
        li.onclick = function () { // runs the function thats inside the action
            action.func()
        }

        ul.appendChild(li)
    })

    // creates the right click menu on cursor position
    rightClickMenu.style.display = 'block'
    rightClickMenu.style.left = event.pageX + 'px'
    rightClickMenu.style.top = event.pageY + 'px'
}

function deleteRightClickMenu() {
    if (lastRightClickMenu != null) {
        lastRightClickMenu.remove()
    }
}


// adds the new chat message into html
function addChatMessage(messageID, channelID, userID, username, message) {
    // extract the message date from messageID
    const msgDate = new Date(Number((BigInt(messageID) >> BigInt(20)))).toLocaleString()

    const chatNameColor = '#e7e7e7'
    const pic = 'profilepic.jpg'

    // create a <li> that holds the message
    const li = document.createElement('li')
    li.className = 'msg'
    li.id = messageID
    li.setAttribute('on-context-menu', 'message')

    li.setAttribute('user-id', userID)

    // create a <img> that shows profile pic on the left
    const img = document.createElement('img')
    // img.className = 'msg-profile-pic'
    img.className = 'msg-profile-pic'
    img.setAttribute('on-context-menu', 'msgUser')
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
    msgNameDiv.setAttribute('on-context-menu', 'msgUser')
    msgNameDiv.textContent = username
    msgDataDiv.style.color = chatNameColor

    // and next to it create a <div> that displays the date of msg on the right
    const msgDateDiv = document.createElement('div')
    msgDateDiv.className = 'msg-date'
    msgDateDiv.textContent = msgDate

    // append name and date to msgNameAndDateDiv
    msgNameAndDateDiv.appendChild(msgNameDiv)
    msgNameAndDateDiv.appendChild(msgDateDiv)

    // now create a <div> under name and date that displays the message
    const msgTextDiv = document.createElement('div')
    msgTextDiv.className = 'msg-text'
    msgTextDiv.setAttribute('on-context-menu', 'message')
    msgTextDiv.textContent = message

    // append both name/date <div> and msg <div> to msgDatDiv
    msgDataDiv.appendChild(msgNameAndDateDiv)
    msgDataDiv.appendChild(msgTextDiv)

    // append both the profile pic and message data to the <li>
    li.appendChild(img)
    li.appendChild(msgDataDiv)

    // and finally append the message to the message list
    document.getElementById('chat-message-list').appendChild(li)

    // const chatNameColor = '#e7e7e7'
    // const pic = 'profilepic.jpg'

    // const chatElement = 
    // `<li class="msg" id="${messageID}" user-id="${userID}">
    //     <img class="msg-profile-pic" src="${pic}" width="40" height="40">
    //     <div class="msg-data">
    //         <div class="msg-name-and-date">
    //             <div class="msg-user-name" style="color: ${chatNameColor}">${username}</div>
    //             <div class="msg-date">${msgDate}</div>
    //         </div>
    //         <div class="msg-text">${message}</div>
    //     </div>
    // </li>`

    // const message
    // const msgProfilePic = document.getElementById(messageID)
    // const msgUserName = document.getElementById(messageID).querySelector('.msg-user-name')
    

    // document.getElementById('chat-message-list').insertAdjacentHTML('beforeend', chatElement)
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
    memberList.appendChild(li)
}

function addServer(serverID, serverName, picture) {
    const button = document.createElement('button')
    // button.className = 'server'
    button.id = serverID
    button.setAttribute('server-name', serverName)
    // button.setAttribute('clickable')
    button.style.backgroundImage = `url(${picture})`
    document.getElementById('server-list').append(button)


    // const serverElement = `<button class="server" id="${serverID}" server-name="${serverName}"></button>`
    // document.getElementById('server-list').insertAdjacentHTML('beforeend', serverElement)
    // const button = document.getElementById(serverID)
    
    button.addEventListener('click', () => {
        selectServer(BigInt(button.id))
    })

    // button.addEventListener('contextmenu', function(event) {
    //     console.log('rightclicked on server', button.id)

    //     const actions = [
    //         { text: 'Rename server', color: '', func: () => renameServer(serverID) },
    //         { text: 'Delete server', color: 'red', func: () => deleteServer(serverID) }
    //     ]

    //     createRightClickMenu(actions, event)
    // })
}

function addChannel(channelID, channelName) {
    const button = document.createElement('button')
    button.className = 'channel'
    button.id = channelID

    const buttonName = document.createElement('div')
    buttonName.textContent = channelName

    button.appendChild(buttonName)

    document.getElementById('channels-list').appendChild(button)

    button.addEventListener('click', () => {
        selectChannel(button.id)
    })

    // button.addEventListener('contextmenu', function(event) {
    //     console.log('rightclicked on channel', button.id)

    //     const actions = [
    //         { text: 'Rename channel', color: '', func: () => renameChannel(channelID) },
    //         { text: 'Delete channel', color: 'red', func: () => deleteChannel(channelID) }
    //     ]
    //     createRightClickMenu(actions, event)
    // })
}

var channelsHidden = false
function toggleChannelsVisibility() {
    const list = document.getElementById('channels-list')
    const channels = Array.from(list.children)

    channels.forEach(channel => {
        if (!channelsHidden) {
            if (channel.id != getCurrentChannelID()) {
                channel.style.display = 'none'
            }
        } else {
            channel.style.display = ''
        }
    })
    if (!channelsHidden) {
        channelsHidden = true
    } else {
        channelsHidden = false
    }
}

function setSelectedChannelBackground(channelID, previousChannelID) {
    document.getElementById(channelID.toString()).style.backgroundColor = '#36393f'
    // document.getElementById(previousChannelID.toString()).style.background = 'transparent'
    document.getElementById(previousChannelID.toString()).removeAttribute('style')
}

function resetChannels() {
    document.getElementById('channels-list').innerHTML = ''
}

function resetMessages() {
    document.getElementById('chat-message-list').innerHTML = '' // empties chat
}

