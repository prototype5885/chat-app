
function getCurrentChannelID() {
    return BigInt(localStorage.getItem('currentChannelID'))
}

function setCurrentChannelID(channelID) {
    localStorage.setItem("currentChannelID", channelID.toString())
}

function getCurrentServerID() {
    return BigInt(localStorage.getItem('currentServerID'))
}

function setCurrentServerID(serverID) {
    localStorage.setItem("currentServerID", serverID.toString())
}