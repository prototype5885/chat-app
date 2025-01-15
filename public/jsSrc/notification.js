class NotificationClass {
    constructor() {
        this.NotificationSound = document.getElementById("notification-sound")

        if (Notification.permission !== "granted") {
            console.warn("Notifications are not enabled, requesting permission...")
            Notification.requestPermission()
        } else {
            console.log("Notifications are enabled")
        }
    }

    sendNotification(userID, message) {
        const userInfo = MemberListClass.getUserInfo(userID)
        if (Notification.permission === "granted" && document.hidden) {
            new Notification(userInfo.username, {
                body: message,
                icon: userInfo.pic // Optional icon
            })
        }
    }
}