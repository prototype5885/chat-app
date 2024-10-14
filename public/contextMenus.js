var defaultRightClick = false

// delete context menu if left clicked somewhere thats not
// a context menu list element
document.addEventListener("click", function (event) {
    deleteCtxMenu()
})

// delete context menu if right clicked somewhere thats not registered
// with context menu listener
document.addEventListener("contextmenu", function (event) {
    if (!defaultRightClick) {
        event.preventDefault()
    }
    deleteCtxMenu()
})

function registerContextMenu(element, callback) {
    element.addEventListener("contextmenu", (event) => {
        event.preventDefault()
        deleteCtxMenu()
        event.stopPropagation()
        callback(event.pageX, event.pageY)
    })
}

function createContextMenu(actions, pageX, pageY) {
    if (actions.length == 0) {
        return
    }

    // create the right click menu
    const rightClickMenu = document.createElement("div")
    rightClickMenu.id = "right-click-menu"
    document.body.appendChild(rightClickMenu)

    // create ul that holds the menu items
    let ul = document.createElement("ul")
    rightClickMenu.appendChild(ul)

    // add a menu item for each action
    actions.forEach(function (action) {
        const li = document.createElement("li")
        li.textContent = action.text
        if (action.color === "red") {
            li.className = "cm-red" // to make the text red from css
        }
        // this will assing the function for each element
        li.onclick = function () {
            action.func()
        }

        ul.appendChild(li)
    })

    // creates the right click menu on cursor position
    rightClickMenu.style.display = "block"
    rightClickMenu.style.left = `${pageX}px`
    rightClickMenu.style.top = `${pageY}px`
}

function deleteCtxMenu() {
    const rightClickmenu = document.getElementById("right-click-menu")
    if (rightClickmenu != null) {
        rightClickmenu.remove()
    }
}

function serverCtxMenu(serverID, owned, pageX, pageY) {
    const actions = []

    if (owned) { actions.push({ text: "Server Settings", func: () => addWindow("server-settings") }) }
    if (owned) { actions.push({ text: "Create Invite Link", func: () => requestInviteLink(serverID) }) }
    // if (owned) { actions.push({ text: "Delete Server", color: "red", func: () => requestDeleteServer(serverID) }) }
    if (!owned) { actions.push({ text: "Leave Server", color: "red", func: () => requestLeaveServer(serverID) }) }
    // if (!owned) { actions.push({ text: "Report Server", color: "red" }) }

    createContextMenu(actions, pageX, pageY)
}

function channelCtxMenu(channelID, pageX, pageY) {
    function renameChannel(channelID) {
        console.log("renaming channel", channelID)
    }

    function deleteChannel(channelID) {
        console.log("deleting channel", channelID)
    }

    const actions = [
        { text: "Rename channel", color: "", func: () => renameChannel(channelID) },
        { text: "Delete channel", color: "red", func: () => deleteChannel(channelID) }
    ]
    createContextMenu(actions, pageX, pageY)
}

function userCtxMenu(userID, pageX, pageY) {
    function addFriend(userID) {
        console.log("Adding friend", userID)
    }

    function reportUser(userID) {
        console.log("Reporting user", userID)
    }

    function removeFriend(userID) {
        console.log("Removing friend", userID)
    }

    function copyUserID(userID) {
        console.log("Copying user ID", userID)
        navigator.clipboard.writeText(userID)
    }

    const actions = [
        { text: "Add friend", func: () => addFriend(userID) },
        { text: "Report user", color: "red", func: () => reportUser(userID) },
        { text: "Remove friend", color: "red", func: () => removeFriend(userID) },
        { text: "Copy user ID", func: () => copyUserID(userID) }
    ]
    createContextMenu(actions, pageX, pageY)
}

function messageCtxMenu(messageID, owner, pageX, pageY) {
    function copyText() {
        const chatMsg = document.getElementById(messageID).querySelector(".msg-text").textContent
        console.log("Copied to clipboard:", chatMsg)
        navigator.clipboard.writeText(chatMsg)
    }

    const actions = []
    actions.push({ text: "Copy text", func: () => copyText() })
    if (owner) { actions.push({ text: "Delete message", color: "red", func: () => requestDeleteChatMessage(messageID) }) }
    if (!owner) { actions.push({ text: "Report message", color: "red" }) }
    createContextMenu(actions, pageX, pageY)
}