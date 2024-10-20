function initNotification() {
    if (Notification.permission !== "granted") {
        console.warn("Notifications are not enabled, requesting permission...")
        Notification.requestPermission()
    } else {
        console.log("Notifications are enabled")
    }
}

function sendNotification(userID, message) {
    const userInfo = getUserInfo(userID)
    if (Notification.permission === "granted") {
        new Notification(userInfo.username, {
            body: message,
            icon: userInfo.pic // Optional icon
        })
    }
}