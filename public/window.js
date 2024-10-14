var activeWindow
var activeWindowMaximized = false

var lastTop
var lastLeft
var lastWidth
var lastHeight

function addWindow(type) {
    if (activeWindow == null) {
        // create window if none is open
        createWindow(type)
    } else if (activeWindow.getAttribute("type") != type) {
        // if a window is already open, just swap them
        deleteWindow()
        createWindow(type)
    } else {
        // close if trying to open window of same type if already open
        deleteWindow()
    }
}

function deleteWindow() {
    activeWindow.remove()
    activeWindow = null
}

function maximizeWindow() {
    if (activeWindowMaximized) {
        activeWindow.style.top = lastTop
        activeWindow.style.left = lastLeft
        activeWindow.style.width = lastWidth
        activeWindow.style.height = lastHeight

        activeWindowMaximized = false
    } else {
        lastTop = activeWindow.style.top
        lastLeft = activeWindow.style.left
        lastWidth = activeWindow.style.width
        lastHeight = activeWindow.style.height

        activeWindow.style.top = ""
        activeWindow.style.left = ""
        activeWindow.style.width = "100%"
        activeWindow.style.height = "100%"

        activeWindowMaximized = true
    }
}

function createWindow(type) {
    // create main window div
    activeWindow = document.createElement("div")
    activeWindow.className = "window"
    activeWindow.setAttribute("type", type)

    const size = 80
    const topLeft = 50 / (100 / (100 - size))

    activeWindow.style.top = `${topLeft}%`
    activeWindow.style.left = `${topLeft}%`
    activeWindow.style.width = `${size}%`
    activeWindow.style.height = `${size}%`


    // this will be the top bar that holds title and exit buttons etc
    const topBar = document.createElement("div")
    topBar.className = "window-top-bar"
    activeWindow.appendChild(topBar)

    const mainPart = document.createElement("div")
    mainPart.className = "window-main"
    activeWindow.appendChild(mainPart)

    // the left part that holds title name
    const topBarLeft = document.createElement("div")
    topBarLeft.className = "window-top-bar-left"
    topBar.appendChild(topBarLeft)

    // the right part that holds buttons
    const topBarRight = document.createElement("div")
    topBarRight.className = "window-top-bar-right"
    topBar.appendChild(topBarRight)

    // button that maximizes/returns to size
    const maximizeButton = document.createElement("button")
    maximizeButton.className = "window-maximize-button"
    topBarRight.appendChild(maximizeButton)

    registerClick(maximizeButton, () => { maximizeWindow() })

    // button that closes the window
    const exitButton = document.createElement("button")
    exitButton.className = "window-exit-button"
    topBarRight.appendChild(exitButton)

    // register the exit button
    registerClick(exitButton, () => { deleteWindow() })

    // and finally add it to html
    document.body.appendChild(activeWindow)

    let isDragging = false
    let offsetX
    let offsetY

    topBarLeft.addEventListener('mousedown', (e) => {
        e.preventDefault()
        isDragging = true
        offsetX = e.clientX - activeWindow.getBoundingClientRect().left
        offsetY = e.clientY - activeWindow.getBoundingClientRect().top
        topBarLeft.style.cursor = "grabbing"
    })
    document.addEventListener('mousemove', (e) => {
        if (isDragging) {
            let newPosX = e.clientX - offsetX
            let newPosY = e.clientY - offsetY

            // clamn so it can leave the window
            newPosX = Math.max(0, Math.min(window.innerWidth - activeWindow.clientWidth, newPosX))
            newPosY = Math.max(0, Math.min(window.innerHeight - activeWindow.clientHeight, newPosY))

            activeWindow.style.left = `${newPosX}px`
            activeWindow.style.top = `${newPosY}px`

        }
    })
    document.addEventListener('mouseup', (e) => {
        if (isDragging) {
            isDragging = false
            topBarLeft.style.cursor = ""
        }
    })

    const left = []

    switch (type) {
        case "user-settings":
            topBarLeft.textContent = "User settings"
            break
        case "server-settings":
            topBarLeft.textContent = "Server settings"
            break
    }
}


