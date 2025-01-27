class ColorsClass {
    static mainColor = '#36393f'
    static hoverColor = 'rgba(128, 128, 128, 0.075)'
    static bitDarkerColor = '#2B2D31'
    static darkColor = '#232428'
    static darkerColor = '#1E1F22'
    static grayTextColor = '#949BA4'
    static darkTransparent = '#111214d1'
    static darkNonTransparent = '#111214'
    static brighterTransparent = '#656565d1'
    static loadingColor = '#000000b5'
    static selectedColor = 'rgba(255, 255, 255, 0.05)'

    static textColor = '#C5C7CB'

    static blue = '#5865F2'
    static green = '#00b700'
}


class MainClass {
    static receivedInitialUserData = false // don't continue loading until own user data is received
    static receivedImageHostAddress = false // don't continue loading until host address of image server arrived
    static memberListLoaded = false // don't add chat history until server member list is received

    static imageHost = ''

    static myUserID = ''
    static myDisplayName = ''
    static myProfilePic = ''
    static myPronouns = ''
    static myStatusText = ''
    static myFriends = []
    static myBlocks = []
    static #currentServerID = '1'
    static #currentChannelID = '0'

    static currentPictureViewerPicPath = ''
    static currentPictureViewerPicName = ''

    static defaultProfilePic = '/content/static/default_profilepic.webp'

    static init() {
        if (this.isElectron() || this.isPWA()) {
            document.getElementById('server-name-container').style.borderTopLeftRadius = '16px'
        }

        // this runs after webpage was loaded
        document.addEventListener('DOMContentLoaded', async () => {
            Translation.setLanguage()

            TouchControlsClass.init()
            ContextMenuClass.init()
            NotificationClass.init()
            // ChatInputClass.init()
            UserPanelClass.init()

            // ChatMessageListClass.init()
            // ChannelListClass.createChannelList()
            ServerListClass.init()

            ServerBannerClass.init()

            AttachmentInputClass.init()

            // add the direct messages button
            // ServerListClass.addServer('1', 0, Translation.get('dm'), 'content/static/mail.svg', '1')

            await WebsocketClass.connectToWebsocket()

            // setInterval(this.checkForUpdates, 3000)

            const pictureViewer = document.getElementById('picture-viewer')
            ContextMenuClass.registerContextMenu(pictureViewer, (pageX, pageY) => {
                ContextMenuClass.pictureCtxMenu(this.currentPictureViewerPicPath, this.currentPictureViewerPicName, pageX, pageY)
            })

        })
    }

    static setCurrentChannelID(id) {
        console.log(`Changing current channel ID to: ${id}`)
        this.#currentChannelID = id
    }

    static getCurrentChannelID() {
        return this.#currentChannelID
    }

    static setCurrentServerID(id) {
        console.log(`Changing current server ID to: ${id}`)
        this.#currentServerID = id
    }

    static getCurrentServerID() {
        return this.#currentServerID
    }

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

    static isElectron() {
        // Renderer process
        const text = 'Running in electron'
        if (typeof window !== 'undefined' && typeof window.process === 'object' && window.process.type === 'renderer') {
            console.log(text)
            return true
        }

        // Main process
        if (typeof process !== 'undefined' && typeof process.versions === 'object' && !!process.versions.electron) {
            console.log(text)
            return true
        }

        // Detect the user agent when the `nodeIntegration` option is set to true
        if (typeof navigator === 'object' && typeof navigator.userAgent === 'string' && navigator.userAgent.indexOf('Electron') >= 0) {
            console.log(text)
            return true
        }

        return false
    }

    static isPWA() {
        return ['fullscreen', 'standalone', 'minimal-ui'].some(
            (displayMode) => window.matchMedia('(display-mode: ' + displayMode + ')').matches
        )
    }

    static getAvatarFullPath(pic) {
        if (pic === '' || pic === undefined || pic == null) {
            return this.defaultProfilePic
        } else {
            return this.imageHost + '/content/avatars/' + pic
        }
    }

    static checkDisplayName(displayName) {
        if (displayName === '' || displayName === undefined || displayName === null) {
            return ''
        } else {
            return displayName
        }
    }


    static setOwnUserID(userID) {
        this.myUserID = userID
        console.log(`Own user ID has been set to [${this.myUserID}]`)
    }

    static getOwnUserID() {
        return this.myUserID
    }

    static setOwnDisplayName(displayName) {
        displayName = this.checkDisplayName(displayName)
        this.myDisplayName = displayName

        if (displayName === '') {
            UserPanelClass.setUserPanelName(this.myUserID)
        } else {
            UserPanelClass.setUserPanelName(displayName)
        }

        console.log(`Own display name has been set to [${this.myDisplayName}]`)
    }

    static setOwnPronouns(pronouns) {
        this.myPronouns = pronouns
        console.log(`Own pronouns have been set to [${this.myPronouns}]`)
    }

    static setOwnStatusText(statusText) {
        this.myStatusText = statusText
        UserPanelClass.setUserPanelStatusText(statusText)
        console.log(`Own status text has been set to [${this.myStatusText}]`)
    }

    static setOwnProfilePic(pic) {
        pic = this.getAvatarFullPath(pic)

        this.myProfilePic = pic
        UserPanelClass.setUserPanelPic(pic)
        console.log(`Own profile pic has been set to [${this.myProfilePic}]`)
    }

    static setMyFriends(friends) {
        this.myFriends = friends
        console.log(`You have [${this.myFriends.length}] friends, they are: [${this.myFriends}]`)
    }

    static removeFriend(userID) {
        for (let i = 0; i < this.myFriends.length; i++) {
            if (this.myFriends[i] === userID) {
                this.myFriends.splice(i, 1)
                return
            }
        }
        console.error(`Local error: could not remove user ID [${userID}] from ownFriends array`)
    }

    static setBlockedUsers(blocks) {
        this.myBlocks = blocks
        console.log(`You have blocked [${this.myBlocks.length}] users, they are: [${this.myBlocks}]`)
    }


    static setButtonActive(button, active) {
        if (active) {
            button.classList.remove('noHover')
            button.disabled = false
        } else {
            button.classList.add('noHover')
            button.disabled = true
        }
    }


    static getScrollDistanceFromBottom(e) {
        return e.scrollHeight - e.scrollTop - e.clientHeight
    }

    static getScrollDistanceFromTop(e) {

    }

    static registerClick(element, callback) {
        element.addEventListener('click', (event) => {
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

MainClass.init()