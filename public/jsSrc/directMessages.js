class DirectMessagesClass {
    static DirectMessages = document.getElementById('direct-messages')
    static DmChatList = document.getElementById('dm-chat-list')

    static init() {
        const dmFriendsButton = document.getElementById('dm-friends-button')
        MainClass.registerClick(dmFriendsButton, async () => {
            await ChannelListClass.selectChannel(chatID, true)
        })
    }

    static addDirectMessages(json) {
        const dmChatIDs = json
        for (let i = 0, len = dmChatIDs.length; i < len; i++) {
            this.addDirectMessage(dmChatIDs[i])
        }
        console.log(LocalStorageClass.selectLastChannel())
        // ChannelListClass.selectChannel(LocalStorageClass.selectLastChannel())
    }

    static addDirectMessage(chatID) {
        const dmButton = document.createElement('button')
        dmButton.id = chatID
        dmButton.textContent = chatID
        this.DmChatList.appendChild(dmButton)

        MainClass.registerClick(dmButton, async () => {
            await ChannelListClass.selectChannel(chatID, true)
        })
    }
}