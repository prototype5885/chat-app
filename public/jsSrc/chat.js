function sendNotification(userID, message) {
    const userInfo = getUserInfo(userID)
    if (Notification.permission === "granted") {
        new Notification(userInfo.username, {
            body: message,
            icon: userInfo.pic // Optional icon
        })
    }
}

function waitUntilBoolIsTrue(checkFunction, interval = 10) {
    return new Promise((resolve) => {
        const intervalId = setInterval(() => {
            if (checkFunction()) {
                clearInterval(intervalId)
                resolve()
            }
        }, interval)
    })
}