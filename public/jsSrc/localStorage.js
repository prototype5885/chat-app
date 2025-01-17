class LocalStorageClass {
    constructor(main) {
        this.main = main

        this.localStorageSupported = false

        if (typeof (Storage) === "undefined") {
            console.log("Browser doesn't support storage")
        } else {
            console.log("Browser supports storage")
            this.localStorageSupported = true
        }
    }

    getLastChannels() {
        return localStorage.getItem("lastChannels")
    }

    setLastChannels(value) {
        console.log("Last channel has been set to: " + value)
        localStorage.setItem("lastChannels", value)
    }

    getLastServer() {
        return localStorage.getItem("lastServer")
    }

    setLastServer(value) {
        console.log("Last server has been set to: " + value)
        localStorage.setItem("lastServer", value)
    }


    updateLastChannelsStorage() {
        if (!this.localStorageSupported) {
            console.warn("Local storage isn't enabled in browser, can't update lastChannels value")
            return
        }

        const json = this.getLastChannels()

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
        lastChannels[this.main.currentServerID] = this.main.currentChannelID
        this.setLastChannels(JSON.stringify(lastChannels))
    }

    // delete servers from lastChannels that no longer exist
    lookForDeletedServersInLastChannels(serverList) {
        if (!this.localStorageSupported) {
            console.warn("Local storage isn't enabled in browser, can't look for deleted servers in lastChannels value")
            return
        }

        const json = this.getLastChannels()
        if (json != null) {
            let lastChannels = JSON.parse(json)
            const li = serverList.querySelectorAll(".server")

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
                this.setLastChannels(JSON.stringify(newLastChannels))
            }
        } else {
            console.log("No lastChannels in localStorage exists")
        }
    }

    // delete a single server from lastChannels
    removeServerFromLastChannels(serverID) {
        if (!this.localStorageSupported) {
            console.warn(`Local storage isn't enabled in browser, can't delete server ID [${serverID}] from lastChannels value`)
            return
        }

        const json = this.getLastChannels()
        if (json != null) {
            let lastChannels = JSON.parse(json)
            if (serverID.toString() in lastChannels) {
                delete lastChannels[serverID.toString()]
                this.setLastChannels(JSON.stringify(lastChannels))
                console.log(`Removed server ID ${serverID} from lastChannels`)
            } else {
                console.log(`Server ID ${serverID} doesn"t exist in lastChannels`)
            }
        }
    }

    // selects the last selected channel after clicking on a server
    selectLastChannels() {
        if (!this.localStorageSupported) {
            console.warn("Local storage isn't enabled in browser, can't select last used channel on server, selecting first channel instead")
            return null
        }

        const json = this.getLastChannels()
        if (json != null) {
            let lastChannels = JSON.parse(json)
            const lastChannel = lastChannels[this.main.currentServerID.toString()]
            if (lastChannel != null) {
                return lastChannel
            } else {
                console.log("Current server does not have any last channel set in localStorage, selecting first channel...")
                return null
            }
        } else {
            console.log("No lastChannels in localStorage exists, selecting first channel...")
            return null
        }
    }

    getServerCount() {
        if (!this.localStorageSupported) {
            console.warn(`Local storage isn't enabled in browser, can't get serverCount value, returning 0`)
            return 0
        } else {
            return localStorage.getItem("serverCount")
        }
    }

    setServerCount(value) {
        if (!this.localStorageSupported) {
            console.warn(`Local storage isn't enabled in browser, can't set serverCount value`)
            return 0
        } else {
            localStorage.setItem("serverCount", value)
        }
    }
}