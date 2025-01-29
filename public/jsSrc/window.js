class WindowManagerClass {

    static openWindows = [] // this stores every open windows as hashmap by type value
    static lastSelected = new Map()
    static currentUpdateUserDataLabel


    // this is called when something creates as new window
    static addWindow(type, id) {
        WindowManagerClass.openWindows.push(new Window(type, id))
    }

    static setCurrentUpdateUserDataResponseLabel(success) {
        if (success) {
            WindowManagerClass.currentUpdateUserDataLabel.style.color = 'green'
            WindowManagerClass.currentUpdateUserDataLabel.textContent = 'Success'
        } else {
            WindowManagerClass.currentUpdateUserDataLabel.style.color = 'red'
            WindowManagerClass.currentUpdateUserDataLabel.textContent = 'Failure'
        }
        WindowManagerClass.currentUpdateUserDataLabel = undefined
    }
}

class Window {
    constructor(type, id) {
        this.window
        this.id = id
        this.topBar
        this.topBarLeft
        this.windowMain
        this.type = type
        this.lastTop
        this.lastLeft
        this.lastWidth
        this.lastHeight
        this.maximized
        this.isDragging = false
        this.offsetX
        this.offsetY
        this.handleMouseDown = this.mouseDown.bind(this)
        this.handleMouseMove = this.mouseMove.bind(this)
        this.handleMouseUp = this.mouseUp.bind(this)
        this.handleSelectWindow = this.selectWindow.bind(this)
        this.createWindow()
        this.selectWindow()
    }

    deleteWindow() {
        // remove event listeners
        this.topBarLeft.removeEventListener('mousedown', this.handleMouseDown)
        this.window.removeEventListener('mousedown', this.handleSelectWindow)
        // remove html element before deleting from openWindows array
        this.window.remove()
        // remove from lastSelected

        // find and delete from array
        for (let i = 0; i < WindowManagerClass.openWindows.length; i++) {
            if (WindowManagerClass.openWindows[i] === this) {
                WindowManagerClass.openWindows.splice(i, 1)
                WindowManagerClass.lastSelected.delete(i)
            }
        }
    }

    maximizeWindow() {
        if (this.maximized) {
            this.window.style.top = this.lastTop
            this.window.style.left = this.lastLeft
            this.window.style.width = this.lastWidth
            this.window.style.height = this.lastHeight

            this.maximized = false

            this.makeActive()
        } else {
            this.lastTop = this.window.style.top
            this.lastLeft = this.window.style.left
            this.lastWidth = this.window.style.width
            this.lastHeight = this.window.style.height

            this.window.style.top = ''
            this.window.style.left = ''
            this.window.style.width = '100%'
            this.window.style.height = '100%'

            this.maximized = true

            this.topBar.style.backgroundColor = ColorsClass.darkNonTransparent
            this.window.style.border = ''
        }
    }

    makeActive() {
        this.topBar.style.backgroundColor = ColorsClass.darkTransparent
        this.window.style.border = '1px solid var(--dark-transparent)'
    }

    makeInactive() {
        this.topBar.style.backgroundColor = ColorsClass.brighterTransparent
        this.window.style.border = '1px solid var(--brighter-transparent)'
    }

    // this runs when the top bar of window is held to move the window
    mouseDown(e) {
        document.addEventListener('mousemove', this.handleMouseMove)
        document.addEventListener('mouseup', this.handleMouseUp)
        e.preventDefault()
        this.isDragging = true
        this.offsetX = e.clientX - this.window.getBoundingClientRect().left
        this.offsetY = e.clientY - this.window.getBoundingClientRect().top
        this.topBarLeft.style.cursor = 'grabbing'
    }

    mouseMove(e) {
        if (this.isDragging) {
            let newPosX = e.clientX - this.offsetX
            let newPosY = e.clientY - this.offsetY

            // clamp so it can't leave the window
            newPosX = Math.max(0, Math.min(window.innerWidth - this.window.clientWidth, newPosX))
            newPosY = Math.max(0, Math.min(window.innerHeight - this.window.clientHeight, newPosY))

            this.window.style.left = `${newPosX}px`
            this.window.style.top = `${newPosY}px`
        }
    }

    mouseUp(e) {
        if (this.isDragging) {
            this.isDragging = false
            this.topBarLeft.style.cursor = ''
            // remove event listeners when stopped moving window
            document.removeEventListener('mousemove', this.handleMouseMove)
            document.removeEventListener('mouseup', this.handleMouseUp)
        }
    }

    // when window is clicked on, makes it selected window
    selectWindow() {
        // check if selected window is maximized, then don't select if it is
        if (this.maximized) {
            return
        }

        this.makeActive()

        // set order 0 for selected window
        for (let i = 0; i < WindowManagerClass.openWindows.length; i++) {
            if (WindowManagerClass.openWindows[i] === this) {
                WindowManagerClass.lastSelected.set(i, 0)
                break
            }
        }

        // add + 1 for the order value of each other windows
        // also look for highest value among them
        let highestValue = 0
        for (let i = 0; i < WindowManagerClass.openWindows.length; i++) {
            if (WindowManagerClass.openWindows[i] !== this) {
                const value = WindowManagerClass.lastSelected.get(i) + 1
                WindowManagerClass.lastSelected.set(i, value)
                if (value > highestValue) {
                    highestValue = value
                }
            }
        }

        // order the values here
        const orderedKeys = []
        for (let i = 0; i < highestValue + 1; i++) {
            for (const [key, value] of WindowManagerClass.lastSelected.entries()) {
                if (value === i) {
                    orderedKeys.push(key)
                }
            }
        }
        // then trim the array
        // for example 0 1 6 8 would be 0 1 2 3
        for (let i = 0; i < orderedKeys.length; i++) {
            WindowManagerClass.lastSelected.set(orderedKeys[i], i)
        }

        // set the z index for each window
        for (const [key, value] of WindowManagerClass.lastSelected.entries()) {
            if (WindowManagerClass.openWindows[key] != null) {
                WindowManagerClass.openWindows[key].window.style.zIndex = (900 - value).toString()
                if (WindowManagerClass.openWindows[key] !== this) {
                    WindowManagerClass.openWindows[key].makeInactive()
                }

            }
        }
    }

    async createWindow() {
        // create main window div
        this.window = document.createElement('div')
        this.window.className = 'window'
        this.window.setAttribute('type', this.type)

        const size = 70
        const topLeft = 50 / (100 / (100 - size))

        this.window.style.top = `${topLeft}%`
        this.window.style.left = `${topLeft}%`
        this.window.style.width = `${size}%`
        this.window.style.height = `${size}%`

        this.window.style.border = '1px solid var(--dark-transparent)'
        this.window.style.zIndex = '901'

        // this will be the top bar that holds title and exit buttons etc
        this.topBar = document.createElement('div')
        this.topBar.className = 'window-top-bar'
        this.topBar.style.backgroundColor = ColorsClass.darkTransparent
        this.window.appendChild(this.topBar)

        // the left part that holds title name
        this.topBarLeft = document.createElement('div')
        this.topBarLeft.className = 'window-top-bar-left'
        this.topBar.appendChild(this.topBarLeft)

        // the right part that holds buttons
        const topBarRight = document.createElement('div')
        topBarRight.className = 'window-top-bar-right'
        this.topBar.appendChild(topBarRight)

        // button that maximizes/returns to size
        const maximizeButton = document.createElement('button')
        maximizeButton.className = 'window-maximize-button'
        topBarRight.appendChild(maximizeButton)

        maximizeButton.innerHTML = `
            <svg width='28' height='28' xmlns='http://www.w3.org/2000/svg'>
                <rect x='4' y='4' width='20' height='20' stroke='grey' fill='none' stroke-width='2'/>
            </svg>`

        // this is the main part under the top bar that holds content
        this.windowMain = document.createElement('div')
        this.windowMain.className = 'window-main'
        this.window.appendChild(this.windowMain)

        MainClass.registerClick(maximizeButton, () => {
            this.maximizeWindow()
        })

        // button that closes the window
        const exitButton = document.createElement('button')
        exitButton.className = 'window-exit-button'
        topBarRight.appendChild(exitButton)

        exitButton.innerHTML = `
            <svg width='28' height='28' xmlns='http://www.w3.org/2000/svg'>
              <line x1='4' y1='4' x2='24' y2='24' stroke='grey' stroke-width='2'/>
              <line x1='4' y1='24' x2='24' y2='4' stroke='grey' stroke-width='2'/>
            </svg>`

        // register the exit button
        MainClass.registerClick(exitButton, () => {
            this.deleteWindow()
        })

        // and finally add it to html
        document.body.appendChild(this.window)

        // add listeners for moving mouse and releasing mouse button
        this.topBarLeft.addEventListener('mousedown', this.handleMouseDown)

        // this listener makes it possible to select active window
        this.window.addEventListener('mousedown', this.handleSelectWindow)

        switch (this.type) {
            case 'user-settings':
                this.topBarLeft.textContent = Translation.get('userSettings')
                await this.createSettingsLeftSide(this.windowMain, this.type, 0)
                break
            case 'server-settings':
                this.topBarLeft.textContent = Translation.get('serverSettings')
                await this.createSettingsLeftSide(this.windowMain, this.type, this.id)
                break
            case 'channel-settings':
                this.topBarLeft.textContent = Translation.get('channelSettings')
                await this.createSettingsLeftSide(this.windowMain, this.type, this.id)
                break
        }
    }

    async pictureUploader(settings, pictureType, serverID) {
        settings.innerHTML += `  
    <div>
        <label>${Translation.get('maximum')}: 1,5 MB</label>
        <br>
        <input type='file' accept='.jpg,.png,.gif' name='image' class='pic-uploader' accept='image/*' style='display: none'>
        <button class='select-pic'></button>
        <br>
        <button class='button send-pic-button noHover' disabled>${Translation.get('applyPicture')}</button>
        <br>
        <label class='pic-response-label'></label>
    </div>`

        let picture
        if (pictureType === 'server-pic') {
            picture = window.getComputedStyle(document.getElementById(serverID)).backgroundImage
        } else if (pictureType === 'profile-pic') {
            picture = `url(${MainClass.myProfilePic})`
        } else {
            console.error('Unknown picture type provided for picture loader')
            return
        }
        settings.querySelector('.select-pic').style.backgroundImage = picture


        const sendButton = settings.querySelector('.send-pic-button')
        const responseLabel = settings.querySelector('.pic-response-label')
        // clicked on the pic
        settings.querySelector('.select-pic').addEventListener('click', async (event) => {
            settings.querySelector('.pic-uploader').click()

            responseLabel.style.color = ''
            responseLabel.textContent = ''
        })

        let previousPicture

        // added a pic
        const picUploader = settings.querySelector('.pic-uploader')
        picUploader.addEventListener('change', async (event) => {
            if (picUploader.files.length === 0) {
                console.log('No picture has been selected')
                // picPreview.style.backgroundImage = `url(${ownProfilePic})`
            } else {
                MainClass.setButtonActive(sendButton, true)
                console.log('Picture selected')
                const reader = new FileReader()
                reader.readAsDataURL(picUploader.files[0])

                reader.onload = function (e) {
                    const picPreview = settings.querySelector('.select-pic')
                    previousPicture = picPreview.style.backgroundImage
                    picPreview.style.backgroundImage = `url(${e.target.result})`
                }
            }
        })

        // upload the pic
        sendButton.addEventListener('click', async (event) => {
            console.log(`Uploading ${pictureType}...`)
            event.preventDefault()

            if (picUploader.files.length === 0) {
                console.warn('No new profile was attached')
                return
            }

            const formData = new FormData()

            formData.append(`${pictureType}`, picUploader.files[0])
            if (pictureType === 'server-pic') {
                formData.append('serverID', serverID)
            }

            // Initialize a new XMLHttpRequest object
            const uploadRequest = new XMLHttpRequest()

            uploadRequest.upload.onprogress = function (e) {
                if (e.lengthComputable) {
                    var percent = (e.loaded / e.total) * 100
                    responseLabel.textContent = Math.round(percent) + ' %'
                    console.log(responseLabel.textContent)
                }
            }

            uploadRequest.onload = function () {
                if (uploadRequest.status === 200) {
                    const successText = 'Picture was uploaded successfully'
                    console.log(successText)
                    responseLabel.style.color = 'green'
                    responseLabel.textContent = successText
                } else {
                    console.error(uploadRequest.responseText)
                    responseLabel.style.color = 'red'
                    responseLabel.textContent = uploadRequest.responseText
                    settings.querySelector('.select-pic').style.backgroundImage = previousPicture
                    MainClass.setButtonActive(sendButton, false)
                }
            }

            // Open the request and send the FormData
            uploadRequest.open('POST', `/upload-${pictureType}`, true)
            uploadRequest.send(formData)
        })
    }

    createSettingsLeftSide = async (windowMain, type, value) => {
        const leftSide = document.createElement('div')
        leftSide.className = 'settings-left'
        const rightSide = document.createElement('div')
        rightSide.className = 'settings-right'

        windowMain.appendChild(leftSide)
        windowMain.appendChild(rightSide)

        const addElementsLeftSide = (elements) => {
            const settingsLeft = windowMain.querySelector('.settings-left')

            const top = document.createElement('div')
            top.className = 'settings-left-top'

            settingsLeft.appendChild(top)

            const settingsList = document.createElement('div')
            settingsList.className = 'settings-list'

            for (let i = 0; i < elements.length; i++) {
                const button = document.createElement('button')
                button.className = 'left-side-button'

                const textDiv = document.createElement('div')
                textDiv.textContent = Translation.get(elements[i].text)
                button.appendChild(textDiv)

                const settingsRight = windowMain.querySelector('.settings-right')

                MainClass.registerClick(button, async () => {
                    // reset selection of items on left
                    for (const b of settingsList.children) {
                        if (b !== button) {
                            b.removeAttribute('style')
                        } else {
                            b.style.backgroundColor = ColorsClass.mainColor
                        }
                    }

                    // reset the right side
                    settingsRight.textContent = ''

                    console.log('Selected my', elements[i].text)

                    const topRight = document.createElement('div')
                    topRight.className = 'settings-right-top'

                    const label = document.createElement('div')
                    label.className = 'settings-right-label'
                    label.textContent = Translation.get(elements[i].text)
                    topRight.appendChild(label)

                    settingsRight.appendChild(topRight)

                    const mainRight = document.createElement('div')
                    mainRight.className = 'settings-right-main'

                    settingsRight.appendChild(mainRight)

                    switch (elements[i].text) {
                        case 'profile':
                            const profileSettings = document.createElement('div')
                            profileSettings.className = 'profile-settings'
                            mainRight.appendChild(profileSettings)

                            profileSettings.innerHTML = `
                            <div>
                                <label class='input-label'>${Translation.get('displayName')}:</label>
                                <input class='change-display-name' maxlength='32' value='${MainClass.myDisplayName}'>
                                <br>
                                <label class='input-label'>${Translation.get('pronouns')}:</label>
                                <input class='change-pronoun' placeholder='they/them' maxlength='16' value='${MainClass.myPronouns}'>
                                <br>
                                <label class='input-label'>${Translation.get('statusText')}:</label>
                                <input class='change-status' placeholder='Was passiert?' value='${MainClass.myStatusText}' maxlength='32'>
                                <br>
                                <button class='button update-account-data'>${Translation.get('apply')}</button>
                                <br>
                                <label class='update-data-response-label'></label>
                            </div>`

                            await this.pictureUploader(profileSettings, 'profile-pic', '')

                            // applying username, pronouns, etc
                            profileSettings.querySelector('.update-account-data').addEventListener('click', function () {
                                const newDisplayName = profileSettings.querySelector('.change-display-name').value
                                let displayNameChanged = false
                                if (newDisplayName !== MainClass.myDisplayName) {
                                    displayNameChanged = true
                                }

                                const newPronouns = profileSettings.querySelector('.change-pronoun').value
                                let pronounsChanged = false
                                if (newPronouns !== MainClass.myPronouns) {
                                    pronounsChanged = true
                                }

                                const newStatusText = profileSettings.querySelector('.change-status').value
                                let statusTextChanged = false

                                if (newStatusText !== MainClass.myStatusText) {
                                    statusTextChanged = true
                                }

                                if (!displayNameChanged && !pronounsChanged && !statusTextChanged) {
                                    console.warn('No user settings was changed')
                                    return
                                }

                                console.log('Sending updated user settings to server')

                                const updatedUserData = {
                                    DisplayName: newDisplayName,
                                    Pronouns: newPronouns,
                                    StatusText: newStatusText,
                                    NewDN: displayNameChanged,
                                    NewP: pronounsChanged,
                                    NewST: statusTextChanged
                                }

                                WindowManagerClass.currentUpdateUserDataLabel = profileSettings.querySelector('.update-data-response-label')
                                WebsocketClass.requestUpdateUserData(updatedUserData)

                            })
                            break
                        case 'account':
                            const accountSettings = document.createElement('div')
                            accountSettings.className = 'account-settings'
                            mainRight.appendChild(accountSettings)

                            accountSettings.innerHTML = `
                            <div>
                                <label class='input-label'>${Translation.get('currentPassword')}:</label>
                                <input type='password' class='change-password-current' maxlength='32'>
                                <br>
                                <label class='input-label'>${Translation.get('newPassword')}:</label>
                                <input type='password' class='change-password-first' maxlength='32' minlength='4' required>
                                <br>
                                <label class='input-label'>${Translation.get('confirmNewPassword')}:</label>
                                <input type='password' class='change-password-second' maxlength='32' minlength='4' required>
                                <br>
                                <button class='button update-password'>Apply</button>
                            </div>`
                            break
                        case 'server':
                            const serverSettings = document.createElement('div')
                            serverSettings.className = 'server-settings'
                            mainRight.appendChild(serverSettings)

                            const serverName = ServerListClass.getServerName(value)

                            serverSettings.innerHTML = `
                            <div>
                                <label class='input-label'>Server name:</label>
                                <input class='change-server-name' maxlength='32' value='${serverName}'>
                                <br>
                                <button class='button update-server-data'>Apply</button>
                                <br>
                                <label class='update-data-response-label'></label>
                            </div>`

                            console.log(`Changing picture of server ID [${value}]`)
                            await this.pictureUploader(serverSettings, 'server-pic', value)

                            // updating server data
                            serverSettings.querySelector('.update-server-data').addEventListener('click', function () {
                                const newServerName = serverSettings.querySelector('.change-server-name').value
                                let serverNameChanged = false
                                if (newServerName !== ServerListClass.getServerName(value)) {
                                    serverNameChanged = true
                                }

                                if (!serverNameChanged) {
                                    console.warn('No server settings were changed')
                                    return
                                }

                                const updatedServerData = {
                                    ServerID: value,
                                    Name: newServerName,
                                    NewSN: serverNameChanged,
                                }

                                WindowManagerClass.currentUpdateUserDataLabel = serverSettings.querySelector('.update-data-response-label')
                                WebsocketClass.requestUpdateServerData(updatedServerData)
                            })
                            break
                        case 'channel':
                            const channelSettings = document.createElement('div')
                            channelSettings.className = 'channel-settings'
                            mainRight.appendChild(channelSettings)

                            const channelName = ChannelListClass.getChannelName(value)

                            channelSettings.innerHTML = `
                            <div>
                                <label class='input-label'>Channel name:</label>
                                <input class='change-channel-name' maxlength='16' value='${channelName}'>
                                <br>
                                <button class='button update-channel-data'>Apply</button>
                                <br>
                                <label class='update-data-response-label'></label>
                            </div>`


                            // updating channel data
                            channelSettings.querySelector('.update-channel-data').addEventListener('click', function () {
                                const newChannelName = channelSettings.querySelector('.change-channel-name').value
                                let channelNameChanged = false
                                if (newChannelName !== ChannelListClass.getChannelName(value)) {
                                    channelNameChanged = true
                                }

                                if (!channelNameChanged) {
                                    console.warn('No channel settings were changed')
                                    return
                                }

                                const updatedChannelData = {
                                    ChannelID: value,
                                    Name: newChannelName,
                                    NewCN: channelNameChanged,
                                }

                                WindowManagerClass.currentUpdateUserDataLabel = channelSettings.querySelector('.update-data-response-label')
                                WebsocketClass.requestUpdateChannelData(updatedChannelData)
                            })
                            break
                        case 'language':
                            const leftSideContent = []

                            leftSideContent.push({text: 'Deutsch', code: 'de'})
                            leftSideContent.push({text: 'Español', code: 'es'})
                            leftSideContent.push({text: 'English', code: 'en'})
                            leftSideContent.push({text: 'Русский', code: 'ru'})
                            leftSideContent.push({text: 'Magyar', code: 'hu'})
                            leftSideContent.push({text: 'Türkçe', code: 'tr'})
                            leftSideContent.push({text: '汉语', code: 'zh'})
                            leftSideContent.push({text: '日本語', code: 'jp'})

                            const settingsRightMain = windowMain.querySelector('.settings-right-main')

                            const container = document.createElement('div')
                            container.className = 'language-button-list'
                            container.style.display = 'flex'
                            // container.style.width = '100%'
                            settingsRightMain.appendChild(container)


                            for (let i = 0; i < leftSideContent.length; i++) {
                                const button = document.createElement('button')


                                if (LocalStorageClass.getLanguage() === null) {
                                    if (navigator.language.split('-')[0] === leftSideContent[i].code) {
                                        button.style.backgroundColor = ColorsClass.selectedColor
                                    }
                                } else if (LocalStorageClass.getLanguage() === leftSideContent[i].code) {
                                    button.style.backgroundColor = ColorsClass.selectedColor
                                }

                                const flag = document.createElement('img')
                                flag.className = 'flag'

                                switch (leftSideContent[i].code) {
                                    case 'es':
                                        flag.src = `/static/flags/${leftSideContent[i].code}.webp`
                                        break
                                    default:
                                        flag.src = `/static/flags/${leftSideContent[i].code}.svg`
                                        break
                                }


                                button.appendChild(flag)

                                const label = document.createElement('span')
                                label.textContent = leftSideContent[i].text
                                button.appendChild(label)

                                button.addEventListener('click', function () {
                                    LocalStorageClass.setLanguage(leftSideContent[i].code)
                                    window.location.reload()
                                })

                                container.appendChild(button)
                            }

                            break
                    }
                })

                settingsList.appendChild(button)
            }

            settingsLeft.appendChild(settingsList)
        }

        const leftSideContent = []
        // add these elements to the left side
        switch (type) {
            case 'user-settings':
                leftSideContent.push({text: 'profile'})
                // leftSideContent.push({text: 'account'})
                leftSideContent.push({text: 'language'})
                break
            case 'server-settings':
                leftSideContent.push({text: 'server'})
                break
            case 'channel-settings':
                leftSideContent.push({text: 'channel'})
        }
        addElementsLeftSide(leftSideContent)

        leftSide.querySelector('.settings-list').firstElementChild.click()
    }
}
