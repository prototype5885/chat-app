class ServerListClass {
    constructor(main, chatMessageList, memberList, channelList, localStorage, contextMenu) {
        this.main = main
        this.chatMessageList = chatMessageList
        this.channelList = channelList
        this.memberList = memberList
        this.localStorage = localStorage
        this.contextMenu = contextMenu

        this.ServerList = document.getElementById("server-list")
        this.serverSeparators = this.ServerList.querySelectorAll(".servers-separator")
        this.FriendsChat = document.getElementById("friends-chat")
        this.ServerName = document.getElementById("server-name")
        this.AddServerButton = document.getElementById("add-server-button")

        // add bubble when hovering over add server button
        MainClass.registerHover(this.AddServerButton, () => {
            BubbleClass.createBubble(this.AddServerButton, "Add Server", "right", 15)
        }, () => {
            BubbleClass.deleteBubble()
        })

        // hide notification marker as this doesn't use it,
        // but it's needed for formatting reasons
        this.AddServerButton.nextElementSibling.style.backgroundColor = "transparent"

        this.AddServerButton.addEventListener("click", () => {
            WebsocketClass.requestAddServer("server")
        })

    }

    createPlaceHolderServers() {
        console.log("Adding placeholder servers")
        this.removePlaceholderServers()
        const serverCount = this.localStorage.getServerCount()
        if (serverCount !== 0) {
            for (let i = 0; i < serverCount; i++) {
                this.addPlaceholderServer()
            }
        }
        this.calculateServerAmount()
    }

    removePlaceholderServers() {
        console.log("Removing placeholder servers")
        // remove placeholder servers
        const placeholderButtons = this.ServerList.querySelectorAll(".placeholder-server")
        for (let i = 0; i < placeholderButtons.length; i++) {
            placeholderButtons[i].remove()
        }
    }

    addPlaceholderServer() {
        const buttonParent = this.addServer("", 0, "", "", "placeholder-server")
        let button = buttonParent.querySelector("button")
        button.nextElementSibling.style.backgroundColor = "transparent"
        button.textContent = ""
    }

    addServer(serverID, owned, serverName, picture, className) {
        if (serverName === "") {
            serverName = serverID
        }

        // this li will hold the server and notification thing, which is the span
        const li = document.createElement("li")
        li.className = className
        this.ServerList.append(li)

        // create the server button itself
        const button = document.createElement("button")
        button.id = serverID
        button.setAttribute("name", serverName)
        li.append(button)

        // set picture of server
        if (picture !== "") {
            if (serverID !== "2000") {
                picture = "content/avatars/" + picture
            }
            button.style.backgroundImage = `url(${picture})`
        } else {
            if (serverName !== "") {
                button.textContent = serverName[0].toUpperCase()
            }
        }

        const span = document.createElement("span")
        span.className = "server-notification"
        li.append(span)

        button.setAttribute("owned", owned)

        MainClass.registerClick(button, () => {
            this.selectServer(serverID)
        })
        this.contextMenu.registerContextMenu(button, (pageX, pageY) => {
            this.contextMenu.serverCtxMenu(serverID, owned, pageX, pageY)
        })
        MainClass.registerHover(button, () => {
            if (serverID !== main.currentServerID) {
                button.style.borderRadius = "35%"
                button.style.backgroundColor = "#5865F2"
                span.style.height = "24px"
            }
            BubbleClass.createBubble(button, ServerListClass.getServerName(serverID), "right", 15)
        }, () => {
            if (serverID !== main.currentServerID) {
                button.style.borderRadius = "50%"
                button.style.backgroundColor = ""
                span.style.height = "8px"
            }
            BubbleClass.deleteBubble()
        })

        // this check needs to be made else adding placeholder servers will break serverCount value,
        // as it would reset the serverCount value while adding placeholders, as fix serverSeparatorVisibility
        // is run manually only after creating each placeholder servers on startup
        if (className === "server") {
            this.calculateServerAmount()
        }

        return li
    }

    removeServers() {
        document.querySelectorAll('.server').forEach(server => {
            server.remove()
        })
    }

    selectServer(serverID) {
        let dm = false
        if (serverID === "2000") {
            dm = true
        }

        if (dm) {
            console.log("Selected direct messages")
            this.FriendsChat.removeAttribute("style")
            this.channelList.Channels.style.display = "none"
            document.getElementById("chat-input-form-container").style.display = "none"
            this.main.currentChannelID = "0"
        } else {
            console.log("Selected server ID", serverID, ", requesting list of channels...")
            this.channelList.Channels.removeAttribute("style")
            this.FriendsChat.style.display = "none"
            document.getElementById("chat-input-form-container").style.display = ""
        }

        const serverButton = document.getElementById(serverID)
        if (serverButton == null) {
            console.log("Previous server set in")
        }

        if (serverID === main.currentServerID) {
            console.log("Selected server is already the current one")
            return
        }

        main.memberListLoaded = false

        // this will reset the previously selected server's visuals
        const previousServerButton = document.getElementById(main.currentServerID)
        if (previousServerButton != null) {
            previousServerButton.nextElementSibling.style.height = "8px"
            previousServerButton.style.backgroundColor = ""
            previousServerButton.style.borderRadius = "50%"
        }

        main.currentServerID = serverID

        serverButton.nextElementSibling.style.height = "36px"

        if (!dm) {
            const owned = serverButton.getAttribute("owned")

            // hide add channel button if server isn't own
            if (owned === "true") {
                this.channelList.AddChannelButton.style.display = "block"
            } else {
                this.channelList.AddChannelButton.style.display = "none"
            }
        }


        if (dm) {
            this.memberList.hideMemberList()
        } else {
            this.memberList.showMemberList()
        }

        this.channelList.resetChannels()
        this.chatMessageList.resetChatMessages()
        this.memberList.resetMemberList()

        if (!dm) {
            WebsocketClass.requestChannelList()
            WebsocketClass.requestMemberList()
        }

        this.ServerName.textContent = serverButton.getAttribute("name")

        this.localStorage.setLastServer(serverID)
    }


    deleteServer(serverID) {
        console.log("Deleting server ID:", serverID)
        // check if class is correct
        document.getElementById(serverID).parentNode.remove()
        this.calculateServerAmount()
    }

    setServerPicture(serverID, picture) {
        picture = "content/avatars/" + picture
        document.getElementById(serverID).style.backgroundImage = `url("${picture}")`
    }

    setServerName(serverID, name) {
        console.log(`Changing name of server ID [${serverID}] to [${name}]`)
        document.getElementById(serverID).setAttribute("name", name)
    }

    static getServerName(serverID) {
        return document.getElementById(serverID).getAttribute("name")
    }

    calculateServerAmount() {
        const allServers = this.ServerList.querySelectorAll(".server, .placeholder-server")
        const servers = this.ServerList.querySelectorAll(".server")
        this.localStorage.setServerCount(servers.length)

        if (allServers.length !== 0) {
            this.serverSeparators.forEach((separator) => {
                separator.style.display = "block"
            })
        } else {
            this.serverSeparators.forEach((separator) => {
                separator.style.display = "none"
            })
        }
    }


    serverWhiteThingSize(thing, newSize) {
        thing.style.height = `${newSize}px`
    }
}