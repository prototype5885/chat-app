function createPlaceHolderServers() {
    const serverCount = getServerCount()
    // const placeholderButtons = []
    if (serverCount !== 0) {
        for (i = 0; i < serverCount; i++) {
            const buttonParent = addServer("", 0, "", "", "placeholder-server")
            let button = buttonParent.querySelector("button")
            button.nextElementSibling.style.backgroundColor = "transparent"
            button.textContent = ""
            // placeholderButtons.push(buttonParent)
        }
    }
    // return placeholderButtons
}

function removePlaceholderServers() {
    // remove placeholder servers
    const placeholderButtons = ServerList.querySelectorAll(".placeholder-server")
    for (let i = 0; i < placeholderButtons.length; i++) {
        placeholderButtons[i].remove()
    }
}

function addServer(serverID, userID, serverName, picture, className) {
    if (serverName === "") {
        serverName = serverID
    }

    // this li will hold the server and notification thing, which is the span
    const li = document.createElement("li")
    li.className = className
    ServerList.append(li)

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

    // bubble on hover
    function onHoverIn() {
        if (serverID !== currentServerID) {
            button.style.borderRadius = "35%"
            button.style.backgroundColor = "#5865F2"
            span.style.height = "24px"
        }
        createbubble(button, serverName, "right", 15)
    }

    function onHoverOut() {
        if (serverID !== currentServerID) {
            button.style.borderRadius = "50%"
            button.style.backgroundColor = ""
            span.style.height = "8px"
        }
        deletebubble()
    }

    var owned = false
    if (userID === ownUserID) {
        owned = true
    }

    button.setAttribute("owned", owned.toString())

    registerClick(button, () => { selectServer(serverID) })
    registerContextMenu(button, (pageX, pageY) => { serverCtxMenu(serverID, owned, pageX, pageY) })
    registerHover(button, () => { onHoverIn() }, () => { onHoverOut() })

    // this check needs to be made else adding placeholder servers will break serverCount value,
    // as it would reset the serverCount value while adding placeholders, as fix serverSeparatorVisibility
    // is ran manually only after creating each placeholder servers on startup
    if (className === "server") {
        serversSeparatorVisibility()
    }

    return li
}

function selectServer(serverID) {
    let dm = false
    if (serverID === "2000") {
        dm = true
    }

    if (dm) {
        console.log("Selected direct messages")
        FriendsChat.removeAttribute("style")
        Channels.style.display = "none"
    } else {
        console.log("Selected server ID", serverID, "Requesting list of channels...")
        Channels.removeAttribute("style")
        FriendsChat.style.display = "none"

    }

    const serverButton = document.getElementById(serverID)
    if (serverButton == null) {
        console.log("Previous server set in")
    }

    if (serverID === currentServerID) {
        console.log("Selected server is already the current one")
        return
    }

    memberListLoaded = false

    // this will reset the previously selected server's visuals
    const previousServerButton = document.getElementById(currentServerID)
    if (previousServerButton != null) {
        previousServerButton.nextElementSibling.style.height = "8px"
        previousServerButton.style.backgroundColor = ""
        previousServerButton.style.borderRadius = "50%"
    }

    currentServerID = serverID

    serverButton.nextElementSibling.style.height = "36px"

    if (!dm) {
        const owned = serverButton.getAttribute("owned")

        // hide add channel button if server isn't own
        if (owned === "true") {
            AddChannelButton.style.display = "block"
        } else {
            AddChannelButton.style.display = "none"
        }
    }


    if (dm) {
        hideMemberList()
    } else {
        showMemberList()
    }

    resetChannels()
    resetChatMessages()
    resetMemberList()

    if (!dm) {
        requestChannelList()
        requestMemberList()
    }

    ServerName.textContent = serverButton.getAttribute("name")

    selectDirectMessages()
}



function deleteServer(serverID) {
    console.log("Deleting server ID:", serverID)
    // check if class is correct
    document.getElementById(serverID).parentNode.remove()
    serversSeparatorVisibility()
}

function serversSeparatorVisibility() {
    const servers = ServerList.querySelectorAll(".server, .placeholder-server")
    setServerCount(servers.length)

    if (servers.length != 0) {
        serverSeparators.forEach((separator) => {
            separator.style.display = "block"
        })
    } else {
        serverSeparators.forEach((separator) => {
            separator.style.display = "none"
        })
    }
}

function serverWhiteThingSize(thing, newSize) {
    thing.style.height = `${newSize}px`
}