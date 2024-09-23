const grey1 = '#949BA4'
const discordGray = '#36393f'
const discordBlue = '#5865F2'
const discrodGreen = '#00b700'

var currentServerID
var currentChannelID


// hide member list when pressing the button
document.getElementById('hide-member-list-button').addEventListener('click', function () {
    const memberList = document.getElementById('member-list')
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

// delete context menu if left clicked somewhere thats not
// a context menu list element
document.addEventListener('click', function (event) {
    deleteRightClickMenu()
})

// delete context menu if right clicked somewhere thats not registered
// with context menu listener
document.addEventListener('contextmenu', function (event) {
    event.preventDefault()
    deleteRightClickMenu()
})

// read the text message for sending
function readChatInput() {
    if (inputArea.value) {
        sendChatMessage(inputArea.value, currentChannelID)
        inputArea.value = ''
        resizeChatInput()
    }
}

function registerClick(element, callback) {
    element.addEventListener('click', (event) => {
        deleteRightClickMenu()
        event.stopPropagation()
        callback()
    })
}

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
    bubble.style.left = `${center.x + 40}px`
    bubble.style.top = `${center.y - height / 2}px`
}

function deletebubble() {
    const bubble = document.getElementById('bubble')
    if (bubble != null) {
        bubble.remove()
    } else {
        console.warn("A bubble was to be deleted but was nowhere to be found")
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

function updateServerImage(button, picture, defaultColor, firstCharacter) {
    if (picture !== '') {
        button.style.backgroundImage = `url(${picture})`

    } else {
        button.style.backgroundColor = defaultColor
        button.textContent = firstCharacter.toUpperCase()
    }
}

function addServer(serverID, serverName, picture, className, defaultColor, hoverColor) {
    // this li will hold the server and notification thing, which is the span
    const li = document.createElement('li')
    li.className = className
    document.getElementById('server-list').append(li)

    // create the server button itself
    const button = document.createElement('button')
    button.id = serverID
    button.setAttribute('server-name', serverName)
    li.append(button)

    // set picture of server
    updateServerImage(button, picture, defaultColor, serverName[0])


    const span = document.createElement('span')
    span.className = 'server-notification'
    li.append(span)

    // bubble on hover
    const bubbleText = document.createElement('div')
    bubbleText.textContent = serverID.toString()



    // this will reset the previously selected server's
    // notification's white thing's size
    function resetPreviousNotificationSize(previousButton) {
        if (previousButton != null) {
            previousButton.nextElementSibling.style.height = '8px'
        }
    }

    function onClick() {
        console.log('Clicked on server:', serverID)

        const previousButton = document.getElementById(currentServerID)

        if (serverID == currentServerID) {
            console.log('Selected server is already the current one')
            return
        }
        resetPreviousNotificationSize(previousButton)

        button.style.backgroundColor = hoverColor
        if (previousButton != null) {
            previousButton.style.backgroundColor = defaultColor
            previousButton.style.borderRadius = '50%'
        }

        currentServerID = serverID

        resetChannels()
        resetMessages()
        requestChannelList()
    }

    function onHoverIn() {
        createbubble(bubbleText, button)
        if (serverID != currentServerID) {
            button.style.backgroundColor = hoverColor
            button.style.borderRadius = '35%'
        }
        span.style.height = '24px'
    }

    function onHoverOut() {
        if (serverID != currentServerID) {
            span.style.height = '8px'
            button.style.backgroundColor = defaultColor
            button.style.borderRadius = '50%'
        }

        deletebubble()
    }

    registerClick(button, () => { onClick() })
    registerRightClick(button, (pageX, pageY) => { serverCtxMenu(serverID, pageX, pageY) })
    registerHover(button, () => { onHoverIn() }, () => { onHoverOut() })
}

function deleteServer(serverID) {
    console.log('Deleting server ID:', serverID)
    document.getElementById(serverID).parentNode.remove()
}

function addChannel(channelID, channelName) {
    const button = document.createElement('button')
    button.id = channelID

    const buttonName = document.createElement('div')
    buttonName.textContent = channelID.toString()

    button.appendChild(buttonName)

    document.getElementById('channels-list').appendChild(button)

    // when clicked on the channel from channel list
    function onClick(channelID) {
        console.log('Clicked on channel:', channelID)

        if (channelID == currentChannelID) {
            console.log('Channel clicked on is already the current one')
            return
        }

        document.getElementById(channelID).style.backgroundColor = discordGray
        const previousChannel = document.getElementById(currentChannelID)
        if (previousChannel != null) {
            document.getElementById(currentChannelID).removeAttribute('style')
        }

        resetMessages()
        currentChannelID = channelID
        requestChatHistory(channelID)
    }

    registerClick(button, () => { onClick(channelID) })
    registerRightClick(button, (pageX, pageY) => { channelCtxMenu(channelID, pageX, pageY) })
}

var channelsHidden = false
function toggleChannelsVisibility() {
    const list = document.getElementById('channels-list')
    const channels = Array.from(list.children)

    channels.forEach(channel => {
        if (!channelsHidden) {
            if (channel.id != currentChannelID) {
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

