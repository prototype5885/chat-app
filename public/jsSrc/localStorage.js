var localStorageSupported = false

function initLocalStorage() {
    if (typeof (Storage) === "undefined") {
        console.log("Browser doesn't support storage")
    } else {
        console.log("Browser supports storage")
        localStorageSupported = true
    }
}

function getLastChannels() {
    return localStorage.getItem("lastChannels")
}

function setLastChannels(value) {
    localStorage.setItem("lastChannels", value)
}

function updateLastChannelsStorage() {
    if (!localStorageSupported) {
        console.warn("Local storage isn't enabled in browser, can't update lastChannels value")
        return
    }

    const json = getLastChannels()

    let lastChannels = {}

    // first parse existing list, in case it exists in browser
    if (json != null) {
        lastChannels = JSON.parse(json)

        // var serverIDstr = currentServerID
        // var channelIDstr = currentChannelID

        // if (serverIDstr in lastChannels && lastChannels[serverIDstr] === channelIDstr) {
        // if currentServerID and currentChannelID matches witht hose in lastChannels localStorage, don't do anything
        // }
    }
    // if channel was changed, overwrite with new one
    lastChannels[currentServerID] = currentChannelID
    setLastChannels(JSON.stringify(lastChannels))
}

// delete servers from lastChannels that no longer exist
function lookForDeletedServersInLastChannels() {
    if (!localStorageSupported) {
        console.warn("Local storage isn't enabled in browser, can't look for deleted servers in lastChannels value")
        return
    }

    const json = getLastChannels()
    if (json != null) {
        let lastChannels = JSON.parse(json)

        const li = ServerList.querySelectorAll(".server")

        const newLastChannels = {}
        li.forEach((li) => {
            const button = li.querySelector("button")
            const id = button.getAttribute("id")
            newLastChannels[id.toString()] = lastChannels[id.toString()]
        })

        if (JSON.stringify(lastChannels) === JSON.stringify(newLastChannels)) {
            console.log("All lastChannels servers in localStorage match")
        } else {
            // most likely one or more servers were deleted while user was offline
            console.warn("lastChannels servers in localStorage don't match with active servers")
            setLastChannels(JSON.stringify(newLastChannels))
        }
    } else {
        console.log("No lastChannels in localStorage exists")
    }
}

// delete a single server from lastChannels
function removeServerFromLastChannels(serverID) {
    if (!localStorageSupported) {
        console.warn(`Local storage isn't enabled in browser, can't delete server ID [${serverID}] from lastChannels value`)
        return
    }

    const json = getLastChannels()
    if (json != null) {
        let lastChannels = JSON.parse(json)
        if (serverID.toString() in lastChannels) {
            delete lastChannels[serverID.toString()]
            setLastChannels(JSON.stringify(lastChannels))
            console.log(`Removed server ID ${serverID} from lastChannels`)
        }
        else {
            console.log(`Server ID ${serverID} doesn"t exist in lastChannels`)
        }
    }
}

// selects the last selected channel after clicking on a server
function selectLastChannels(firstChannelID) {
    if (!localStorageSupported) {
        console.warn("Local storage isn't enabled in browser, can't select last used channel on server, selecting first channel instead")
        selectChannel(firstChannelID)
        return
    }

    const json = getLastChannels()
    if (json != null) {
        let lastChannels = JSON.parse(json)
        const lastChannel = lastChannels[currentServerID.toString()]
        if (lastChannel != null) {
            selectChannel(lastChannel)
        } else {
            console.log("Current server does not have any last channel set in localStorage, selecting first channel...")
            selectChannel(firstChannelID)
        }
    } else {
        console.log("No lastChannels in localStorage exists, selecting first channel...")
        selectChannel(firstChannelID)
    }
}

function getServerCount() {
    if (!localStorageSupported) {
        console.warn(`Local storage isn't enabled in browser, can't get serverCount value, returning 0`)
        return 0
    } else {
        return localStorage.getItem("serverCount")
    }
}

function setServerCount(value) {
    if (!localStorageSupported) {
        console.warn(`Local storage isn't enabled in browser, can't set serverCount value`)
        return 0
    } else {
        localStorage.setItem("serverCount", value)
    }
}