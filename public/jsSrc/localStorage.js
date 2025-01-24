class LocalStorageClass {
    // checkIfLocalStorageSupported() {
    //     if (typeof (Storage) === "undefined") {
    //         console.log("Browser doesn't support storage")
    //         return false
    //     } else {
    //         console.log("Browser supports storage")
    //         return true
    //     }
    // }

    static getLastChannels() {
        return localStorage.getItem("lastChannels")
    }

    static setLastChannels(value) {
        console.log("Last channel has been set to: " + value)
        localStorage.setItem("lastChannels", value)
    }

    static getLastServer() {
        return localStorage.getItem("lastServer")
    }

    static setLastServer(value) {
        console.log("Last server has been set to: " + value)
        localStorage.setItem("lastServer", value)
    }


    static updateLastChannelsStorage() {
        // if (!this.localStorageSupported) {
        //     console.warn("Local storage isn't enabled in browser, can't update lastChannels value")
        //     return
        // }

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
        lastChannels[MainClass.getCurrentServerID()] = MainClass.getCurrentChannelID()
        this.setLastChannels(JSON.stringify(lastChannels))
    }

    // delete servers from lastChannels that no longer exist
    static lookForDeletedServersInLastChannels() {
        // if (!this.localStorageSupported) {
        //     console.warn("Local storage isn't enabled in browser, can't look for deleted servers in lastChannels value")
        //     return
        // }

        const json = this.getLastChannels()
        if (json != null) {
            let lastChannels = JSON.parse(json)
            const li = ServerListClass.ServerList.querySelectorAll(".server, .dm")

            const newLastChannels = {}
            li.forEach((li) => {
                const serverButton = li.querySelector("button")
                newLastChannels[serverButton.id] = lastChannels[serverButton.id]
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
    static removeServerFromLastChannels(serverID) {
        // if (!this.localStorageSupported) {
        //     console.warn(`Local storage isn't enabled in browser, can't delete server ID [${serverID}] from lastChannels value`)
        //     return
        // }

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
    static selectLastChannel() {
        // if (!this.localStorageSupported) {
        //     console.warn("Local storage isn't enabled in browser, can't select last used channel on server, selecting first channel instead")
        //     return null
        // }

        const json = this.getLastChannels()
        if (json != null) {
            let lastChannels = JSON.parse(json)
            const lastChannel = lastChannels[MainClass.getCurrentServerID().toString()]
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

    static getServerCount() {
        // if (!this.localStorageSupported) {
        //     console.warn(`Local storage isn't enabled in browser, can't get serverCount value, returning 0`)
        //     return 0
        // } else {
        return localStorage.getItem("serverCount")
        // }
    }

    static setServerCount(value) {
        // if (!this.localStorageSupported) {
        //     console.warn(`Local storage isn't enabled in browser, can't set serverCount value`)
        //     return 0
        // } else {
        localStorage.setItem("serverCount", value)
        // }
    }
}