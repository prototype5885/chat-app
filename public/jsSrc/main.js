class ColorsClass {
    static mainColor = "#36393f"
    static hoverColor = "#313338"
    static bitDarkerColor = "#2B2D31"
    static darkColor = "#232428"
    static darkerColor = "#1E1F22"
    static grayTextColor = "#949BA4"
    static darkTransparent = "#111214d1"
    static darkNonTransparent = "#111214"
    static brighterTransparent = "#656565d1"
    static loadingColor = "#000000b5"

    static textColor = "#C5C7CB"

    static blue = "#5865F2"
    static green = "#00b700"
}


class MainClass {
    constructor() {
        this.myUserID = ""
        this.myDisplayName = ""
        this.myProfilePic = ""
        this.myPronouns = ""
        this.myStatusText = ""
        this.myFriends = []
        this.myBlocks = []

        this.receivedInitialUserData = false // don't continue loading until own user data is received
        this.receivedImageHostAddress = false // don't continue loading until host address of image server arrived
        this.memberListLoaded = false // don't add chat history until server member list is received

        this.currentServerID = "2000"
        this.currentChannelID = "0"
        this.reachedBeginningOfChannel = false

        // this.imageHost = "http://localhost:8000/"
        this.imageHost = ""

        // this runs after webpage was loaded
        document.addEventListener("DOMContentLoaded", async () => {
            Translation.setLanguage(navigator.language)

            const contextMenu = new ContextMenuClass()
            const localStorage = new LocalStorageClass(this)

            const notification = new NotificationClass()

            const chatInput = new ChatInputClass(this)

            const userPanel = new UserPanelClass(this)

            const memberList = new MemberListClass()
            const chatMessageList = new ChatMessageListClass(this, notification, chatInput)
            const channelList = new ChannelListClass(this, chatMessageList, localStorage, chatInput)
            const serverList = new ServerListClass(this, chatMessageList, memberList, channelList, localStorage)


            const websocket = new WebsocketClass(this, serverList, chatMessageList, channelList, memberList, localStorage, userPanel)

            // add the direct messages button
            serverList.addServer("2000", 0, "Direct Messages", "content/static/mail.svg", "2000")

            await websocket.connectToWebsocket()

            // setInterval(this.checkForUpdates, 3000)
        })
    }

    // checkForUpdates() {
    //     console.log("Checking for update")
    // }

    static defaultProfilePic = "/content/static/default_profilepic.webp"


    static waitUntilBoolIsTrue(checkFunction, interval = 10) {
        return new Promise((resolve) => {
            const intervalId = setInterval(() => {
                if (checkFunction()) {
                    clearInterval(intervalId)
                    resolve()
                }
            }, interval)
        })
    }

    getAvatarFullPath(pic) {
        if (pic === "" || pic === undefined || pic == null) {
            return MainClass.defaultProfilePic
        } else {
            return this.imageHost + "/content/avatars/" + pic
        }
    }

    static checkDisplayName(displayName) {
        if (displayName === "" || displayName === undefined || displayName === null) {
            return ""
        } else {
            return displayName
        }
    }


    setOwnUserID(userID) {
        this.myUserID = userID
        console.log(`Own user ID has been set to [${this.myUserID}]`)
    }

    setOwnDisplayName(displayName) {
        displayName = MainClass.checkDisplayName(displayName)
        this.myDisplayName = displayName

        if (displayName === "") {
            UserPanelClass.setUserPanelName(this.myUserID)
        } else {
            UserPanelClass.setUserPanelName(displayName)
        }

        console.log(`Own display name has been set to [${this.myDisplayName}]`)
    }

    setOwnPronouns(pronouns) {
        this.myPronouns = pronouns
        console.log(`Own pronouns have been set to [${this.myPronouns}]`)
    }

    setOwnStatusText(statusText) {
        this.myStatusText = statusText
        UserPanelClass.setUserPanelStatusText(statusText)
        console.log(`Own status text has been set to [${this.myStatusText}]`)
    }

    setOwnProfilePic(pic) {
        pic = this.getAvatarFullPath(pic)

        this.myProfilePic = pic
        UserPanelClass.setUserPanelPic(pic)
        console.log(`Own profile pic has been set to [${this.myProfilePic}]`)
    }

    setMyFriends(friends) {
        this.myFriends = friends
        console.log(`You have [${this.myFriends.length}] friends, they are: [${this.myFriends}]`)
    }

    removeFriend(userID) {
        for (let i = 0; i < this.myFriends.length; i++) {
            if (this.myFriends[i] === userID) {
                this.myFriends.splice(i, 1)
                return
            }
        }
        console.error(`Local error: could not remove user ID [${userID}] from ownFriends array`)
    }

    setBlockedUsers(blocks) {
        this.myBlocks = blocks
        console.log(`You have blocked [${this.myBlocks.length}] users, they are: [${this.myBlocks}]`)
    }


    static setButtonActive(button, active) {
        if (active) {
            button.classList.remove("noHover")
            button.disabled = false
        } else {
            button.classList.add("noHover")
            button.disabled = true
        }
    }


    static getScrollDistanceFromBottom(e) {
        return e.scrollHeight - e.scrollTop - e.clientHeight
    }

    static getScrollDistanceFromTop(e) {

    }

    static registerClick(element, callback) {
        element.addEventListener("click", (event) => {
            ContextMenuClass.deleteCtxMenu()
            event.stopPropagation()
            callback()
        })
    }

    static registerHover(element, callbackIn, callbackOut) {
        element.addEventListener('mouseover', (event) => {
            // console.log('hovering over', element)
            callbackIn()
        })

        element.addEventListener('mouseout', (event) => {
            // console.log('hovering out', element)
            callbackOut()
        })
    }

    static async calculateSHA256(file) {
        const arrayBuffer = await file.arrayBuffer()
        const hashBuffer = await crypto.subtle.digest('SHA-256', arrayBuffer)
        const hashArray = Array.from(new Uint8Array(hashBuffer))
        const hashHex = hashArray.map(byte => byte.toString(16).padStart(2, '0')).join('')
        console.log(hashHex)
        return hashArray
    }

    static base64toSha256(base64) {
        const base64str = atob(base64)

        const byteArray = new Uint8Array(base64str.length);

        for (let i = 0; i < base64str.length; i++) {
            byteArray[i] = base64str.charCodeAt(i);
        }
        const hashArray = Array.from(new Uint8Array(byteArray))
        return hashArray.map(byte => byte.toString(16).padStart(2, '0')).join('')
    }

    static areArraysEqual(arr1, arr2) {
        if (arr1.length !== arr2.length) {
            return false;
        }
        return arr1.every((element, index) => element === arr2[index]);
    }

    static extractDateFromId(id) {
        return new Date(Number((BigInt(id) >> BigInt(22))))
    }
}

const main = new MainClass()