class ContextMenuClass {
    constructor() {
        this.defaultRightClick = false

        // delete context menu if left-clicked somewhere that's not
        // a context menu list element
        document.addEventListener("click", (event) => {
            ContextMenuClass.deleteCtxMenu()
        })

        // delete context menu if right-clicked somewhere that's not registered
        // with context menu listener
        document.addEventListener("contextmenu", (event) => {
            if (!this.defaultRightClick) {
                event.preventDefault()
            }
            ContextMenuClass.deleteCtxMenu()
        })
    }

    static registerLeftClickContextMenu(element, callback) {
        element.addEventListener("click", (event) => {
            event.preventDefault()
            ContextMenuClass.deleteCtxMenu()
            event.stopPropagation()
            callback(event.pageX, event.pageY)
        })
    }

    static registerContextMenu(element, callback) {
        element.addEventListener("contextmenu", (event) => {
            event.preventDefault()
            ContextMenuClass.deleteCtxMenu()
            event.stopPropagation()
            callback(event.pageX, event.pageY)
        })
    }

    static createContextMenu(actions, pageX, pageY) {
        if (actions.length === 0) {
            return
        }

        // create the right click menu
        const rightClickMenu = document.createElement("div")
        rightClickMenu.id = "ctx-menu"
        document.body.appendChild(rightClickMenu)

        // create ul that holds the menu items
        let ul = document.createElement("ul")
        rightClickMenu.appendChild(ul)

        // add a menu item for each action
        actions.forEach((action) => {
            const li = document.createElement("li")
            li.textContent = action.text
            if (action.color === "red") {
                li.className = "cm-red" // to make the text red from css
            }
            // this will assign the for each element
            li.onclick = () => {
                action.func()
            }

            ul.appendChild(li)
        })

        // creates the right click menu on cursor position
        rightClickMenu.style.display = "block"
        rightClickMenu.style.left = `${pageX}px`
        rightClickMenu.style.top = `${pageY}px`
    }

    static deleteCtxMenu() {
        console.log("Deleting ctx menu")
        const existingCtxMenus = document.querySelectorAll("#ctx-menu")
        for (let i = 0; i < existingCtxMenus.length; i++) {
            existingCtxMenus[i].remove()
        }
        // const rightClickMenu = document.getElementById("ctx-menu")
        // if (rightClickMenu != null) {
        //     rightClickMenu.remove()
        // }
    }

    pictureCtxMenu(path, name, pageX, pageY) {

        function openPicture() {
            window.open(path, '_blank');
        }

        function savePicture() {
            const link = document.createElement('a');
            link.href = path
            link.download = name

            // Trigger a click event on the <a> element to start the download
            link.click();
        }

        const actions = [
            {text: "Save", color: "", func: () => savePicture()},
            {text: "Open in new tab", color: "", func: () => openPicture()},
        ]

        ContextMenuClass.createContextMenu(actions, pageX, pageY)
    }

    static serverCtxMenu(serverID, owned, pageX, pageY) {
        console.log("serverctxmenu")
        const actions = []

        if (owned) {
            actions.push({
                text: "Server Settings",
                func: () => WindowManagerClass.addWindow(main, "server-settings", serverID)
            })
        }
        if (owned) {
            actions.push({text: "Create Invite Link", func: () => WebsocketClass.requestInviteLink(serverID)})
        }
        if (owned) {
            actions.push({
                text: "Delete Server",
                color: "red",
                func: () => WebsocketClass.requestDeleteServer(serverID)
            })
        }
        if (!owned) {
            actions.push({text: "Leave Server", color: "red", func: () => WebsocketClass.requestLeaveServer(serverID)})
        }
        // if (!owned) { actions.push({ text: "Report Server", color: "red" }) }

        ContextMenuClass.createContextMenu(actions, pageX, pageY)
    }

    static channelCtxMenu(channelID, owned, pageX, pageY) {
        const actions = []
        if (owned) {
            actions.push({
                text: "Channel Settings",
                func: () => WindowManagerClass.addWindow(main, "channel-settings", channelID)
            })
        }
        if (owned) {
            actions.push({text: "Delete channel", color: "red", func: () => deleteChannel(channelID)})
        }


        function channelSettings(channelID) {
            console.log("TODO renaming channel", channelID)
        }

        function deleteChannel(channelID) {
            WebsocketClass.requestRemoveChannel(channelID)
        }

        this.createContextMenu(actions, pageX, pageY)
    }

    static userCtxMenu(userID, pageX, pageY) {
        function reportUser(userID) {
            console.log("Reporting user", userID)
        }

        function copyUserID(userID) {
            console.log("Copying user ID", userID)
            navigator.clipboard.writeText(userID)
        }

        const actions = []
        if (!main.myFriends.includes(userID) && userID !== main.myUserID) {
            actions.push({text: "Add friend", func: () => WebsocketClass.requestAddFriend(userID)})
        }
        if (main.myFriends.includes(userID) && userID !== main.myUserID) {
            actions.push({text: "Remove friend", color: "red", func: () => WebsocketClass.requestUnfriend(userID)})
        }
        if (userID !== main.myUserID) {
            actions.push({text: "Block", color: "red", func: () => WebsocketClass.requestBlockUser(userID)})
        }
        if (userID !== main.myUserID) {
            actions.push({text: "Report user", color: "red", func: () => reportUser(userID)})
        }
        actions.push({text: "Copy user ID", func: () => copyUserID(userID)})

        ContextMenuClass.createContextMenu(actions, pageX, pageY)
    }

    static messageCtxMenu(messageID, owner, pageX, pageY) {
        function copyText() {
            const chatMsg = document.getElementById(messageID).querySelector(".msg-text").textContent
            console.log("Copied to clipboard:", chatMsg)
            navigator.clipboard.writeText(chatMsg)
        }

        const actions = []
        actions.push({text: "Copy text", func: () => copyText()})
        if (owner) {
            actions.push({
                text: "Delete message",
                color: "red",
                func: () => WebsocketClass.requestDeleteChatMessage(messageID)
            })
        }
        if (!owner) {
            actions.push({text: "Report message", color: "red"})
        }
        ContextMenuClass.createContextMenu(actions, pageX, pageY)
    }
}

