
const grey1 = '#949BA4'

function getCenterCoordinates(element) {
    const rect = element.getBoundingClientRect();

    const center = {
        x: rect.left + rect.width / 2 + window.scrollX,
        y: rect.top + rect.height / 2 + window.scrollY
    }
    return center
}
// function getElementDimension(element) {
//     const rect = element.getBoundingClientRect();

//     const dimension = {
//         x: rect.width,
//         y: rect.height
//     }
//     return dimension
// }

function createContextMenu(actions, pageX, pageY) {
    // create the right click menu
    const rightClickMenu = document.createElement('div')
    rightClickMenu.id = 'right-click-menu'
    document.body.appendChild(rightClickMenu)

    // create ul that holds the menu items
    let ul = document.createElement('ul')
    rightClickMenu.appendChild(ul)

    // add a menu item for each action
    actions.forEach(function (action) {
        const li = document.createElement('li')
        li.textContent = action.text
        if (action.color === 'red') {
            li.className = 'cm-red' // to make the text red from css
        }
        // this will assing the function for each element
        li.onclick = function () { 
            action.func()
        }

        ul.appendChild(li)
    })

    // creates the right click menu on cursor position
    rightClickMenu.style.display = 'block'
    rightClickMenu.style.left = `${pageX}px`
    rightClickMenu.style.top = `${pageY}px`
}

function deleteRightClickMenu() {
    const rightClickmenu = document.getElementById('right-click-menu')
    if (rightClickmenu != null) {
        rightClickmenu.remove()
    }
}

function createbubble(content, element) {
    // create bubble div that will hold the content
    const bubble = document.createElement('div')
    bubble.id = 'bubble'
    document.body.appendChild(bubble)

    // add the content into it
    bubble.appendChild(content)

    // center of the element that created the bubble
    // bubble will be created relative to this
    const center = getCenterCoordinates(element)

    // get how tall the bubble will be, so can
    // offset the Y position to make it appear
    // centered next to the element
    const height = bubble.getBoundingClientRect().height

    // set the bubble position
    bubble.style.left = `${center.x+40}px`
    bubble.style.top = `${center.y-height/2}px`
}

function deletebubble() {
    const bubble = document.getElementById('bubble')
    if (bubble != null) {
        bubble.remove()
    }
}

// adds the new chat message into html
function addChatMessage(messageID, userID, message) {
    // extract the message date from messageID
    const msgDate = new Date(Number((BigInt(messageID) >> BigInt(20)))).toLocaleString()

    const chatNameColor = '#e7e7e7'
    const pic = 'profilepic.jpg'
    const username = userID.toString()

    // create a <li> that holds the message
    const li = document.createElement('li')
    li.className = 'msg'
    li.id = messageID
    li.setAttribute('user-id', userID)

    registerRightClick(li, (pageX, pageY) => { messageCtxMenu(messageID, pageX, pageY) })

    // create a <img> that shows profile pic on the left
    const img = document.createElement('img')
    // img.className = 'msg-profile-pic'
    img.className = 'msg-profile-pic'
    img.src = pic
    img.alt = 'pfpic'
    img.width = 40
    img.height = 40

    registerRightClick(img, (pageX, pageY) => { userCtxMenu(userID, pageX, pageY) })

    // create a nested <div> that will contain sender name, message and date
    const msgDataDiv = document.createElement('div')
    msgDataDiv.className = 'msg-data'

    // inside that create a sub nested <div> that contains sender name and date
    const msgNameAndDateDiv = document.createElement('div')
    msgNameAndDateDiv.className = 'msg-name-and-date'

    // and inside that create a <div> that displays the sender's name on the left
    const msgNameDiv = document.createElement('div')
    msgNameDiv.className = 'msg-user-name'
    msgNameDiv.textContent = username
    msgDataDiv.style.color = chatNameColor

    registerRightClick(msgNameDiv, (pageX, pageY) => { userCtxMenu(userID, pageX, pageY) })

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
    const li = document.createElement('li')
    li.className = 'server'
    document.getElementById('server-list').append(li)

    const button = document.createElement('button')
    button.id = serverID
    button.setAttribute('server-name', serverName)
    button.style.backgroundImage = `url(${picture})`
    li.append(button)

    const span = document.createElement('span')
    span.className = 'server-notification'
    li.append(span)

    // bubble on hover
    const bubbleText = document.createElement('div')
    bubbleText.textContent = serverID.toString()

    
    registerClick(button, () => { selectServer(serverID) })
    registerRightClick(button, (pageX, pageY) => { serverCtxMenu(serverID, pageX, pageY) })
    registerHover(button, () => { createbubble(bubbleText, button) },  () => { deletebubble() })
}

function addChannel(channelID, channelName) {
    const button = document.createElement('button')
    button.id = channelID

    const buttonName = document.createElement('div')
    buttonName.textContent = channelID.toString()

    button.appendChild(buttonName)

    document.getElementById('channels-list').appendChild(button)

    registerClick(button, () => { selectChannel(channelID) })
    registerRightClick(button, (pageX, pageY) => { channelCtxMenu(channelID, pageX, pageY) })
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
    document.getElementById(previousChannelID.toString()).removeAttribute('style')
}

function resetChannels() {
    document.getElementById('channels-list').innerHTML = ''
}

function resetMessages() {
    const chatMessageList = document.getElementById('chat-message-list')

    // empties chat
    chatMessageList.innerHTML = '' 

    // this makes sure there will be a little gap between chat input box
    // and the chat messages when user is viewing the latest message
    const chatScrollGap = document.createElement('div')
    chatMessageList.appendChild(chatScrollGap)
}

