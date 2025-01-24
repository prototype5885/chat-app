class UserPanelClass {
    static init(main) {
        const UserSettingsButton = document.getElementById('user-settings-button')
        const ToggleMicrophoneButton = document.getElementById('toggle-microphone-button')

        // user settings button
        MainClass.registerHover(UserSettingsButton, () => {
            BubbleClass.createBubble(UserSettingsButton, 'User Settings', 'up', 15)
        }, () => {
            BubbleClass.deleteBubble()
        })
        UserSettingsButton.addEventListener('click', () => {
            WindowManagerClass.addWindow('user-settings', 0)
        })

        // toggle microphone button
        MainClass.registerHover(ToggleMicrophoneButton, () => {
            BubbleClass.createBubble(ToggleMicrophoneButton, 'Toggle Microphone', 'up', 15)
        }, () => {
            BubbleClass.deleteBubble()
        })
    }

    static setUserPanelName(name) {
        document.getElementById('user-panel-name').textContent = name
    }

    static setUserPanelPic(pic) {
        document.getElementById('user-panel-pfp').src = pic
    }

    static setUserPanelStatusText(statusText) {
        document.getElementById('user-panel-status-text').textContent = statusText
    }
}

