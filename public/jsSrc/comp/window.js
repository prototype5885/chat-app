var openWindows = [] // this stores every open windows as hashmap by type value
var lastSelected = new Map()

// this is called when something creates as new window
function addWindow(type) {
    openWindows.push(new Window(type))
}

class Window {
    constructor(type) {
        this.window
        this.topBar
        this.topBarLeft
        this.main
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
            if (openWindows[i] == this) {
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

            // clamn so it can leave the window
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
            if (openWindows[i] == this) {
                lastSelected.set(i, 0)
                break
            }
        }

        // add + 1 for the order value of each other windows
        // also look for highest value among them
        let highestValue = 0
        for (let i = 0; i < openWindows.length; i++) {
            if (openWindows[i] != this) {
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
                if (value == i) {
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
                if (openWindows[key] != this) {
                    openWindows[key].makeInactive()
                }

            }
        }
    }

    createSettingsWindowArea() {
        const leftSide = document.createElement("div")
        leftSide.className = "settings-left"
        const rightSide = document.createElement("div")
        rightSide.className = "settings-right"

        this.main.appendChild(leftSide)
        this.main.appendChild(rightSide)
    }

    addElementsLeftSide(elements) {

    }

    createWindow() {
        // create main window div
        this.window = document.createElement("div")
        this.window.className = "window"
        this.window.setAttribute("type", this.type)

        const size = 50
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
        this.main = document.createElement("div")
        this.main.className = "window-main"
        this.window.appendChild(this.main)

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

        const leftSide = []

        switch (this.type) {
            case "user-settings":
                this.topBarLeft.textContent = "User settings"
                this.createSettingsWindowArea()
                this.addElementsLeftSide(["wtf", "XDDDD"])
                break
            case "server-settings":
                this.topBarLeft.textContent = "Server settings"
                this.createSettingsWindowArea()
                this.addElementsLeftSide(leftSide)
                break
        }

    }
}