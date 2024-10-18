function createPlaceHolderServers() {
    const serverCount = getServerCount()
    const placeholderButtons = []
    if (serverCount !== 0) {
        for (i = 0; i < serverCount; i++) {
            const buttonParent = addServer("", 0, "phs", "", "placeholder-server")
            let button = buttonParent.querySelector("button")
            button.nextElementSibling.style.backgroundColor = "transparent"
            button.textContent = ""
            placeholderButtons.push(buttonParent)
        }
    }
    return placeholderButtons
}

function addServer(serverID, ownerID, serverName, picture, className) {
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
        button.style.backgroundImage = `url(${picture})`
    } else {
        button.textContent = serverName[0].toUpperCase()
    }

    const span = document.createElement("span")
    span.className = "server-notification"
    li.append(span)

    // bubble on hover
    function onHoverIn() {
        if (serverID != currentServerID) {
            button.style.borderRadius = "35%"
            button.style.backgroundColor = "#5865F2"
            span.style.height = "24px"
        }
        createbubble(button, serverName, "right", 15)
    }

    function onHoverOut() {
        if (serverID != currentServerID) {
            button.style.borderRadius = "50%"
            button.style.backgroundColor = ""
            span.style.height = "8px"
        }
        deletebubble()
    }

    var owned = false
    if (ownerID == ownUserID) {
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
    console.log("Selected server ID", serverID, "Requesting list of channels...")

    memberListLoaded = false

    const serverButton = document.getElementById(serverID)
    if (serverButton == null) {
        console.log("Previous server set in")
    }

    if (serverID == currentServerID) {
        console.log("Selected server is already the current one")
        return
    }

    // this will reset the previously selected server's visuals
    const previousServerButton = document.getElementById(currentServerID)
    if (previousServerButton != null) {
        previousServerButton.nextElementSibling.style.height = "8px"
        previousServerButton.style.backgroundColor = ""
        previousServerButton.style.borderRadius = "50%"
    }

    currentServerID = serverID


    serverButton.nextElementSibling.style.height = "36px"

    const owned = serverButton.getAttribute("owned")

    // hide add channel button if server isn't own
    if (owned == "true") {
        AddChannelButton.style.display = "block"
    } else {
        AddChannelButton.style.display = "none"
    }

    if (serverID == "2000") {
        hideMemberList()
    } else {
        showMemberList()
    }

    resetChannels()
    resetChatMessages()
    resetMemberList()

    requestChannelList()
    requestMemberList()

    ServerName.textContent = serverButton.getAttribute("name")
}

function deleteServer(serverID) {
    console.log("Deleting server ID:", serverID)
    // check if class is correct
    document.getElementById(serverID).parentNode.remove()
    serversSeparatorVisibility()
}