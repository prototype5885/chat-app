function setUserPanelName() {
    document.getElementById("user-panel-name").textContent = ownDisplayName
}

function setUserPanelPic() {
    document.getElementById("user-panel-pfp").src = getAvatarFullPath(ownProfilePic)
}

function setUserPanelStatusText(statusText) {
    document.getElementById("user-panel-status-text").textContent = statusText
}