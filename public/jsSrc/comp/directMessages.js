function selectDirectMessages() {
    console.log("XDD")
    // this is needed because when switching to DM, the user doesn't enter any server channels
    // to overwrite it, so this simply sets it to 0 to fix that problem, else switching back to
    // previous server will just write in console: you are already on current channel
    currentChannelID = "0"
}