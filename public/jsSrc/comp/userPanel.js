function changeUserPanelName() {
    UserPanelName.textContent = ownDisplayName
}

function changeUserPanelPic() {
    document.getElementById("user-panel-pfp").src = getAvatarFullPath(ownProfilePic)
}