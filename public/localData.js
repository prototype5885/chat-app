const currentServerIDString = 'currentServerID'
const currentChannelIDString = 'currentChannelID'

// currentChannelID
function getCurrentChannelID() {
    const channelID = sessionStorage.getItem(currentChannelIDString)
    if (channelID != null) {
        return BigInt(channelID)
    } else {
        console.log(currentChannelIDString, 'is null')
        return null
    }
}

function setCurrentChannelID(channelID) {
    sessionStorage.setItem(currentChannelIDString, channelID.toString())
}

// currentServerID
function getCurrentServerID() {
    const serverID = sessionStorage.getItem(currentServerIDString)
    if (serverID != null) {
        return BigInt(serverID)
    } else {
        console.log(currentServerIDString, 'is null')
        return null
    }
}

function setCurrentServerID(serverID) {
    sessionStorage.setItem(currentServerIDString, serverID.toString())
}