class UserPanelClass {
    static init(main) {
        const UserSettingsButton = document.getElementById('user-settings-button')
        const RefreshButton = document.getElementById('refresh-button')

        // user settings button
        MainClass.registerHover(UserSettingsButton, () => {
            BubbleClass.createBubble(UserSettingsButton, Translation.get('userSettings'), 'up', 15)
        }, () => {
            BubbleClass.deleteBubble()
        })
        UserSettingsButton.addEventListener('click', () => {
            WindowManagerClass.addWindow('user-settings', 0)
        })

        // refresh button
        MainClass.registerHover(RefreshButton, () => {
            BubbleClass.createBubble(RefreshButton, 'Force refresh', 'up', 15)
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

