class NotificationClass {
    static NotificationSound = document.getElementById('notification-sound')

    static init() {
        const enabled = this.checkIfNotificationsEnabled()
        if (!enabled) {
            console.warn('Notifications are not enabled, requesting permission...')
            Notification.requestPermission().then(r => {
                console.log('Notification have been enabled')
            })
        } else {
            console.log('Notifications are enabled')
        }
    }

    static sendNotification(userID, message) {
        if (message === '') {
            message = 'Sent something...'
        }
        const userInfo = MemberListClass.getUserInfo(userID)
        if (this.checkIfNotificationsEnabled()) {
            new Notification(userInfo.displayName, {
                body: message,
                icon: userInfo.pic
            })
            // NotificationClass.NotificationSound.play()
        }
    }

    static checkIfNotificationsEnabled() {
        if (Notification.permission !== 'granted') {
            return false
        } else {
            return true
        }
    }
}