class FriendListClass {
    static friendListContainer = document.getElementById('friend-list-container')
    static friendList = this.friendListContainer.querySelector('ul')
    static friendCount = this.friendListContainer.querySelector('label')

    // static enableFriendList() {
    //     if (this.friendList.innerHTML !== '') {
    //         console.warn('Friend list is already loaded')
    //         return
    //     }
    //     document.getElementById('dm-friends-button').style.backgroundColor = ColorsClass.mainColor
    //     this.friendListContainer.style.display = 'flex'
    //     this.addCurrentFriends()
    //     ChatMessageListClass.disableChat()
    //     ChannelListClass.selectNoChannel(true)
    // }
    //
    // static disableFriendList() {
    //     this.friendListContainer.style.display = 'none'
    //     this.friendList.innerHTML = ''
    // }
    //
    // static addCurrentFriends() {
    //     this.updateFriendCount()
    //     for (let f = 0; f < MainClass.myFriends.length; f++) {
    //         this.addFriend(MainClass.myFriends[f], "name", "username", '/content/static/default_profilepic.webp', true, 1, "test status")
    //     }
    // }

    static addFriend(userID, displayName, username, picture, online, status, statusText) {
        const friendStr =
            `<li friend-id="${userID}">
                <div class="profile-pic-container" style="width: 32px; height: 32px">
                    <img class="profile-pic" src="${picture}">
                    <div class="user-status"></div>
                </div>   
                <div class="user-data">
                    <span class="user-name">${userID}</span>
                    <div class="user-status-text">${statusText}</div>
                </div>
            </li>`

        this.friendList.insertAdjacentHTML('beforeend', friendStr)

        const friend = this.friendList.querySelector(`[friend-id="${userID}"]`)
        ContextMenuClass.registerContextMenu(friend, (pageX, pageY) => {
            ContextMenuClass.userCtxMenu(userID, pageX, pageY)
        })
    }

    static updateFriendCount() {
        this.friendCount.textContent = `friends - ${MainClass.myFriends.length}`
    }
}