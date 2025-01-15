class UserPanelClass {
    constructor(main) {
        this.main = main;

        this.UserSettingsButton = document.getElementById("user-settings-button")
        this.ToggleMicrophoneButton = document.getElementById("toggle-microphone-button")

        // user settings button
        MainClass.registerHover(this.UserSettingsButton, () => {
            BubbleClass.createBubble(this.UserSettingsButton, "User Settings", "up", 15)
        }, () => {
            BubbleClass.deleteBubble()
        })
        this.UserSettingsButton.addEventListener("click", (e) => {
            WindowManagerClass.addWindow(main, "user-settings", 0)
        })

        // toggle microphone button
        MainClass.registerHover(this.ToggleMicrophoneButton, () => {
            BubbleClass.createBubble(this.ToggleMicrophoneButton, "Toggle Microphone", "up", 15)
        }, () => {
            BubbleClass.deleteBubble()
        })


    }

    static setUserPanelName(name) {
        document.getElementById("user-panel-name").textContent = name
    }

    static setUserPanelPic(pic) {
        document.getElementById("user-panel-pfp").src = pic
    }

    static setUserPanelStatusText(statusText) {
        document.getElementById("user-panel-status-text").textContent = statusText
    }
}

