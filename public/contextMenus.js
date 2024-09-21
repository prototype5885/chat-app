function registerRightClick(element, callback) {
    element.addEventListener('contextmenu', (event) => {
        event.preventDefault()
        deleteRightClickMenu()
        event.stopPropagation()
        callback(event.pageX, event.pageY)
    })
}

function serverCtxMenu(serverID, pageX, pageY) {
    function renameServer(serverID) {
        console.log('renaming server', serverID)
    }
    
    function deleteServer(serverID) {
        console.log('deleting server', serverID)
    }

    const actions = [
        { text: 'Rename server', func: () => renameServer(serverID) },
        { text: 'Delete server', color: 'red', func: () => deleteServer(serverID) }
    ]
    createContextMenu(actions, pageX, pageY)
}

function channelCtxMenu(channelID, pageX, pageY) {
    function renameChannel(channelID) {
        console.log('renaming channel', channelID)
    }
    
    function deleteChannel(channelID) {
        console.log('deleting channel', channelID)
    }

    const actions = [
        { text: 'Rename channel', color: '', func: () => renameChannel(channelID) },
        { text: 'Delete channel', color: 'red', func: () => deleteChannel(channelID) }
    ]
    createContextMenu(actions, pageX, pageY)
}

function userCtxMenu(userID, pageX, pageY) {
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

    const actions = [
        { text: 'Add friend', func: () => addFriend(userID)},
        { text: 'Report user', color: 'red', func: () => reportUser(userID) },
        { text: 'Remove friend', color: 'red', func: () => removeFriend(userID) },
        { text: 'Copy user ID', func: () => copyUserID(userID) }
    ]
    createContextMenu(actions, pageX, pageY)
}

function messageCtxMenu(messageID, pageX, pageY) {
    const actions = [
        { text: 'Delete message', color: 'red', func: () => requestChatMessageDeletion(messageID)}
    ]
    createContextMenu(actions, pageX, pageY)
}