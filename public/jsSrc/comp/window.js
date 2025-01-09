let openWindows = [] // this stores every open windows as hashmap by type value
let lastSelected = new Map()

// this is called when something creates as new window
function addWindow(type, id) {
    openWindows.push(new Window(type, id))
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
        this.topBarLeft.removeEventListener("mousedown", this.handleMouseDown)
        this.window.removeEventListener("mousedown", this.handleSelectWindow)
        // remove html element before deleting from openWindows array
        this.window.remove()
        // remove from lastSelected

        // find and delete from array
        for (let i = 0; i < openWindows.length; i++) {
            if (openWindows[i] === this) {
                openWindows.splice(i, 1)
                lastSelected.delete(i)
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

            this.window.style.top = ""
            this.window.style.left = ""
            this.window.style.width = "100%"
            this.window.style.height = "100%"

            this.maximized = true

            this.topBar.style.backgroundColor = darkNonTransparent
            this.window.style.border = ""
        }
    }

    makeActive() {
        this.topBar.style.backgroundColor = darkTransparent
        this.window.style.border = "1px solid var(--dark-transparent)"
    }

    makeInactive() {
        this.topBar.style.backgroundColor = brighterTransparent
        this.window.style.border = "1px solid var(--brighter-transparent)"
    }

    // this runs when the top bar of window is held to move the window
    mouseDown(e) {
        document.addEventListener('mousemove', this.handleMouseMove)
        document.addEventListener('mouseup', this.handleMouseUp)
        e.preventDefault()
        this.isDragging = true
        this.offsetX = e.clientX - this.window.getBoundingClientRect().left
        this.offsetY = e.clientY - this.window.getBoundingClientRect().top
        this.topBarLeft.style.cursor = "grabbing"
    }

    mouseMove(e) {
        if (this.isDragging) {
            let newPosX = e.clientX - this.offsetX
            let newPosY = e.clientY - this.offsetY

            // clamp so it can leave the window
            newPosX = Math.max(0, Math.min(window.innerWidth - this.window.clientWidth, newPosX))
            newPosY = Math.max(0, Math.min(window.innerHeight - this.window.clientHeight, newPosY))

            this.window.style.left = `${newPosX}px`
            this.window.style.top = `${newPosY}px`
        }
    }

    mouseUp(e) {
        if (this.isDragging) {
            this.isDragging = false
            this.topBarLeft.style.cursor = ""
            // remove event listeners when stopped moving window
            document.removeEventListener("mousemove", this.handleMouseMove)
            document.removeEventListener("mouseup", this.handleMouseUp)
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
        for (let i = 0; i < openWindows.length; i++) {
            if (openWindows[i] === this) {
                lastSelected.set(i, 0)
                break
            }
        }

        // add + 1 for the order value of each other windows
        // also look for highest value among them
        let highestValue = 0
        for (let i = 0; i < openWindows.length; i++) {
            if (openWindows[i] !== this) {
                const value = lastSelected.get(i) + 1
                lastSelected.set(i, value)
                if (value > highestValue) {
                    highestValue = value
                }
            }
        }

        // order the values here
        const orderedKeys = []
        for (let i = 0; i < highestValue + 1; i++) {
            for (const [key, value] of lastSelected.entries()) {
                if (value === i) {
                    orderedKeys.push(key)
                }
            }
        }
        // then trim the array
        // for example 0 1 6 8 would be 0 1 2 3
        for (let i = 0; i < orderedKeys.length; i++) {
            lastSelected.set(orderedKeys[i], i)
        }

        // set the z index for each window
        for (const [key, value] of lastSelected.entries()) {
            if (openWindows[key] != null) {
                openWindows[key].window.style.zIndex = (900 - value).toString()
                if (openWindows[key] !== this) {
                    openWindows[key].makeInactive()
                }

            }
        }
    }

    createWindow() {
        // create main window div
        this.window = document.createElement("div")
        this.window.className = "window"
        this.window.setAttribute("type", this.type)

        const size = 70
        const topLeft = 50 / (100 / (100 - size))

        this.window.style.top = `${topLeft}%`
        this.window.style.left = `${topLeft}%`
        this.window.style.width = `${size}%`
        this.window.style.height = `${size}%`

        this.window.style.border = "1px solid var(--dark-transparent)"
        this.window.style.zIndex = "901"

        // this will be the top bar that holds title and exit buttons etc
        this.topBar = document.createElement("div")
        this.topBar.className = "window-top-bar"
        this.topBar.style.backgroundColor = darkTransparent
        this.window.appendChild(this.topBar)

        // the left part that holds title name
        this.topBarLeft = document.createElement("div")
        this.topBarLeft.className = "window-top-bar-left"
        this.topBar.appendChild(this.topBarLeft)

        // the right part that holds buttons
        const topBarRight = document.createElement("div")
        topBarRight.className = "window-top-bar-right"
        this.topBar.appendChild(topBarRight)

        // button that maximizes/returns to size
        const maximizeButton = document.createElement("button")
        maximizeButton.className = "window-maximize-button"
        topBarRight.appendChild(maximizeButton)

        // this is the main part under the top bar that holds content
        this.windowMain = document.createElement("div")
        this.windowMain.className = "window-main"
        this.window.appendChild(this.windowMain)

        registerClick(maximizeButton, () => { this.maximizeWindow() })

        // button that closes the window
        const exitButton = document.createElement("button")
        exitButton.className = "window-exit-button"
        topBarRight.appendChild(exitButton)

        // register the exit button
        registerClick(exitButton, () => { this.deleteWindow() })

        // and finally add it to html
        document.body.appendChild(this.window)

        // add listeners for moving mouse and releasing mouse button
        this.topBarLeft.addEventListener('mousedown', this.handleMouseDown)

        // this listener makes it possible to select active window
        this.window.addEventListener("mousedown", this.handleSelectWindow)

        switch (this.type) {
            case "user-settings":
                this.topBarLeft.textContent = "User settings"
                createSettingsLeftSide(this.windowMain, this.type)
                break
            case "server-settings":
                this.topBarLeft.textContent = "Server settings"
                createSettingsLeftSide(this.windowMain, this.type, this.id)
                break
        }
    }
}

function pictureUploader(settings, pictureType, serverID) {
    const sendButon = settings.querySelector(".send-pic-button")
    const responseLabel = settings.querySelector(".pic-response-label")
    // clicked on the pic
    settings.querySelector(".select-pic").addEventListener("click", async (event) => {
        settings.querySelector(".pic-uploader").click()
        
        responseLabel.style.color = ""
        responseLabel.textContent = ""
    })

    // added a pic
    const picUploader = settings.querySelector(".pic-uploader")
    picUploader.addEventListener("change", async (event) => {
        const picPreview = settings.querySelector(".select-pic")
        if (picUploader.files.length === 0) {
            console.log("No picture has been selected")
            picPreview.style.backgroundImage = `url(${ownProfilePic})`
        } else {
            setButtonActive(sendButon, true)
            console.log("Picture selected")
            const reader = new FileReader()
            reader.readAsDataURL(picUploader.files[0])

            reader.onload = function (e) {
                picPreview.style.backgroundImage = `url(${e.target.result})`
            }
        }
    })

    // upload the pic
    sendButon.addEventListener("click", async (event) => {
        console.log(`Uploading ${pictureType}...`)
        event.preventDefault()

        if (picUploader.files.length === 0) {
            console.warn("No new profile was attached")
            return
        }

        const formData = new FormData()

        formData.append(`${pictureType}`, picUploader.files[0])
        if (pictureType === "server-pic") {
            formData.append("serverID", serverID)
        }
        

        const response = await fetch(`/upload-${pictureType}`, {
            method: "POST",
            body: formData
        })

        const respText = await response.text()

        if (response.ok) {
            const successText = "Picture was uploaded successfully"
            console.log(successText)
            responseLabel.style.color = "green"
            responseLabel.textContent = successText
        } else {
            console.error(respText)
            responseLabel.style.color = "red"
            responseLabel.textContent = respText
        }
    })
}

function createSettingsLeftSide(windowMain, type, value) {
    const leftSide = document.createElement("div")
    leftSide.className = "settings-left"
    const rightSide = document.createElement("div")
    rightSide.className = "settings-right"

    windowMain.appendChild(leftSide)
    windowMain.appendChild(rightSide)

    function addElementsLeftSide(elements) {
        const settingsLeft = windowMain.querySelector(".settings-left")

        const top = document.createElement("div")
        top.className = "settings-left-top"

        settingsLeft.appendChild(top)

        const settingsList = document.createElement("div")
        settingsList.className = "settings-list"

        for (let i = 0; i < elements.length; i++) {
            const button = document.createElement("button")
            button.className = "left-side-button"

            const textDiv = document.createElement("div")
            button.textContent = elements[i].text
            button.appendChild(textDiv)

            const settingsRight = windowMain.querySelector(".settings-right")

            registerClick(button, () => {
                // reset selection of items on left
                for (const b of settingsList.children) {
                    if (b !== button) {
                        b.removeAttribute("style")
                    } else {
                        b.style.backgroundColor = mainColor
                    }
                }

                // reset the right side
                settingsRight.textContent = ""

                console.log("Selected my", elements[i].text)
                const mainRight = createSettingsRightSideMyAccount(elements[i].text, settingsRight)
                switch (elements[i].text) {
                    case "Profile":
                        const profileSettings = document.createElement("div")
                        profileSettings.className = "profile-settings"
                        mainRight.appendChild(profileSettings)

                        profileSettings.innerHTML = `
                            <div>
                                <label class="input-label">Display name:</label>
                                <input class="change-display-name" maxlength="32" value="${ownDisplayName}">
                                <br>
                                <label class="input-label">Pronouns:</label>
                                <input class="change-pronoun" placeholder="they/them" maxlength="16" value="${ownPronouns}">
                                <br>
                                <label class="input-label">Status:</label>
                                <input class="change-status" placeholder="Was passiert?" value="${ownStatusText}">
                                <br>
                                <button class="button update-account-data">Apply</button>
                            </div>
                            <div>
                                <input type="file" name="image" class="pic-uploader" accept="image/*" style="display: none">
                                <button class="select-pic" style="background-image: url(${ownProfilePic})"></button>
                                <br>
                                <button class="button send-pic-button noHover" disabled>Apply Picture</button>
                                <br>
                                <label class="pic-response-label"></label>
                            </div>`

                        // applying username, pronouns, etc
                        profileSettings.querySelector(".update-account-data").addEventListener('click', function () {
                            const newDisplayName = profileSettings.querySelector(".change-display-name").value
                            let displayNameChanged = true
                            if (newDisplayName === ownDisplayName) {
                                displayNameChanged = false
                            }

                            const newPronouns = profileSettings.querySelector(".change-pronoun").value
                            let pronounsChanged = true
                            if (newPronouns === ownPronouns) {
                                pronounsChanged = false
                            }

                            const newStatusText = profileSettings.querySelector(".change-status").value
                            let statusTextChanged = true

                            if (newStatusText === ownStatusText) {
                                statusTextChanged = false
                            }

                            if (newDisplayName === ownDisplayName && newPronouns === ownPronouns && newStatusText === ownStatusText) {
                                console.warn("No user settings was changed")
                                return
                            }

                            const updatedUserData = {
                                DisplayName: newDisplayName,
                                Pronouns: newPronouns,
                                StatusText: newStatusText,
                                NewDN: displayNameChanged,
                                NewP: pronounsChanged,
                                NewST: statusTextChanged
                            }

                            requestUpdateUserData(updatedUserData)
                        })

                        pictureUploader(profileSettings, "profile-pic", "")
                        break
                    case "Account":
                        const accountSettings = document.createElement("div")
                        accountSettings.className = "account-settings"
                        mainRight.appendChild(accountSettings)

                        accountSettings.innerHTML = `
                            <div>
                                <label class="input-label">Current Password:</label>
                                <input type="password" class="change-password-current" maxlength="32">
                                <br>
                                <label class="input-label">New Password:</label>
                                <input type="password" class="change-password-first" maxlength="32" minlength="4" required>
                                <br>
                                <label class="input-label">New Password Again:</label>
                                <input type="password" class="change-password-second" maxlength="32" minlength="4" required>
                                <br>
                                <button class="button update-password">Apply</button>
                            </div>`
                        break
                    case "Server":
                        const serverSettings = document.createElement("div")
                        serverSettings.className = "server-settings"
                        mainRight.appendChild(serverSettings)


  

                        serverSettings.innerHTML = `
                            <div>
                                <label class="input-label">Server name:</label>
                                <input class="change-server-name" maxlength="32" value="${value}">
                                <br>
                                <button class="button update-server-data">Apply</button>
                            </div>
                            <div>
                                <input type="file" name="image" class="pic-uploader" accept="image/*" style="display: none">
                                <button class="select-pic"></button>
                                <br>
                                <button class="button send-pic-button noHover" disabled>Apply Picture</button>
                                <br>
                                <label class="pic-response-label"></label>
                            </div>`

                        const serverPic = window.getComputedStyle(document.getElementById(value)).backgroundImage
                        serverSettings.querySelector(".select-pic").style.backgroundImage = serverPic

                         // applying server name
                        serverSettings.querySelector(".update-server-data").addEventListener('click', function () {
                            const newServerName = profileSettings.querySelector(".change-server-name").value
                            let serverNameChanged = false
                            if (newServerName !== this.id) {
                                serverNameChanged = true
                            }



                            if (newServerName ===  this.id) {
                                console.warn("No server settings was changed")
                                return
                            }

                            const updatedServerData = {
                                ServerID: serverID,
                                Servername: newServerName,
                                NewSN: serverNameChanged,
                            }

                            requestUpdateServerData(updatedServerData)
                        })

                        console.log(`Changing picture of server ID [${value}]`)
                        pictureUploader(serverSettings, "server-pic", value)
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
        case "user-settings":
            leftSideContent.push({ text: "Profile" })
            leftSideContent.push({ text: "Account" })
            leftSideContent.push({ text: "1" })
            leftSideContent.push({ text: "2" })
            leftSideContent.push({ text: "3" })
            break
        case "server-settings":
            leftSideContent.push({ text: "Server"})
            leftSideContent.push({ text: "Extra"})
            break
    }
    addElementsLeftSide(leftSideContent)

    leftSide.querySelector(".settings-list").firstElementChild.click()
}

function createSettingsRightSideMyAccount(labelText, settingsRight) {
    const topRight = document.createElement("div")
    topRight.className = "settings-right-top"

    const label = document.createElement("div")
    label.className = "settings-right-label"
    label.textContent = labelText
    topRight.appendChild(label)

    settingsRight.appendChild(topRight)

    const mainRight = document.createElement("div")
    mainRight.className = "settings-right-main"

    settingsRight.appendChild(mainRight)

    return mainRight
}