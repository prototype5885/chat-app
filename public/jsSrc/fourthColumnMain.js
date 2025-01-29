class FourthColumnClass {

}

class MemberListClass {
    static create() {
        const fourthColumnMain = document.getElementById('fourth-column-main')
        fourthColumnMain.innerHTML = `
            <ul id="member-list">
                <label id="online-members" style="order: 0">online</label>
                <label id="offline-members" style="order: 99">offline</label>
            </ul>`
    }

    static addMember(userID, displayName, picture, online, status, statusText) {
        // create a <li> that holds the user
        const li = document.createElement('li')
        li.className = 'member'
        li.id = userID

        const picContainer = document.createElement('div')
        picContainer.className = 'profile-pic-container'

        if (picture === '') {
            picture = MainClass.defaultProfilePic
        } else {
            picture = MainClass.getAvatarFullPath(picture)
        }

        // create a <img> that shows profile pic on the left
        const img = document.createElement('img')
        img.className = 'profile-pic'
        img.src = picture

        // create a <div> that will be a circle in the corner of profile pic to show online status
        const statusDiv = document.createElement('div')
        statusDiv.className = 'user-status'

        picContainer.appendChild(statusDiv)
        picContainer.appendChild(img)

        // create a nested <div> that will contain username and status
        const userDataDiv = document.createElement('div')
        userDataDiv.className = 'user-data'

        if (displayName === '') {
            displayName = userID
        }

        // create <div> that will hold the user's message
        const userNameDiv = document.createElement('div')
        userNameDiv.className = 'display-name'
        userNameDiv.textContent = displayName
        userNameDiv.style.color = ColorsClass.grayTextColor

        // now create a <div> under name that display status text
        const userStatusDiv = document.createElement('div')
        userStatusDiv.className = 'user-status-text'
        userStatusDiv.textContent = statusText

        // append both name/date <div> and msg <div> to msgDatDiv
        userDataDiv.appendChild(userNameDiv)
        userDataDiv.appendChild(userStatusDiv)

        // append both the profile pic and message data to the <li>
        li.appendChild(picContainer)
        li.appendChild(userDataDiv)

        // and finally append the message to the message list
        document.getElementById('member-list').appendChild(li)

        this.changeStatusValueInMemberList(userID, status)
        this.setMemberOnline(userID, online)

        ContextMenuClass.registerContextMenu(li, (pageX, pageY) => {
            ContextMenuClass.userCtxMenu(userID, pageX, pageY)
        })
    }

    static removeMember(userID) {
        const element = document.getElementById(userID)
        if (element.className === 'member') {
            element.remove()
        } else {
            console.log(`Trying to remove member ID [${userID}] but the element is not member class: [${element.className}]`)
        }
    }

    static getUserInfo(userID) {
        const member = document.getElementById(userID)
        if (member != null) {
            const pic = member.querySelector('img.profile-pic').src
            const username = member.querySelector('div.display-name').textContent
            return {displayName: username, pic: pic}
        } else {
            return {displayName: userID, pic: ''}
        }
    }

    static toggleMemberListView() {
        const memberList = document.getElementById('member-list')
        if (memberList.style.display === 'none') {
            this.showMemberList()
        } else {
            this.hideMemberList()
        }
    }

    static hideMemberList() {
        const memberList = document.getElementById('member-list')
        memberList.style.display = 'none'
    }

    static showMemberList() {
        const memberList = document.getElementById('member-list')
        memberList.style.display = 'block'
    }

    static resetMemberList() {
        const memberList = document.getElementById('member-list')
        // memberList.innerHTML = ''
    }

    static changeDisplayNameInMemberList(userID, newDisplayName) {
        try {
            const user = document.getElementById(userID)
            user.querySelector('.display-name').textContent = newDisplayName
        } catch {
            console.error(`Failed changing display name of member ID [${userID}], there is no member list loaded`)
        }
    }

    static changeProfilePicInMemberList(userID, pic) {
        try {
            const user = document.getElementById(userID)
            user.querySelector('.profile-pic').src = pic
        } catch {
            console.error(`Failed changing profile pic of member ID [${userID}], there is no member list loaded`)
        }
    }

    static changeStatusValueInMemberList(userID, newStatus) {
        const container = document.getElementById(userID).querySelector('.profile-pic-container')
        const currentStatus = container.querySelector('.user-status')

        if (currentStatus) {
            currentStatus.remove()
        }

        const status = document.createElement('div')
        status.className = 'user-status'

        switch (newStatus) {
            case 1:
                status.style.backgroundColor = 'limegreen'
                break
            case 2:
                const boolean = document.createElement('div')
                boolean.className = 'orange-status-boolean'
                status.style.backgroundColor = 'orange'
                status.appendChild(boolean)
                break
            case 3:
                status.style.backgroundColor = 'red'
                break
            case 4:
                break
            default:
                status.remove()
                return
        }
        container.appendChild(status)
    }


    static setMemberStatusText(userID, newStatusText) {
        const userStatusText = document.getElementById(userID).querySelector('.user-status-text')
        userStatusText.textContent = newStatusText
    }

    static setMemberOnline(userID, online) {
        const userStatus = document.getElementById(userID).querySelector('.profile-pic-container').querySelector('.user-status')
        const member = document.getElementById(userID)
        member.setAttribute('online', online)
        if (online) {
            member.style.order = 1
            member.removeAttribute('style')
            userStatus.style.display = 'block'
        } else {
            member.style.order = 100
            member.style.filter = 'grayscale(100%)'
            member.style.opacity = '0.5'
            userStatus.style.display = 'none'
        }

        this.calculateOnlineMembers()
    }

    static calculateOnlineMembers() {
        const members = document.getElementById('member-list').querySelectorAll('.member')

        let onlineCount = 0
        for (let i = 0; i < members.length; i++) {
            if (members[i].getAttribute('online') === 'true') {
                onlineCount++
            }
        }

        document.getElementById('online-members').textContent = `${Translation.get('online')} - ${onlineCount}`
        document.getElementById('offline-members').textContent = `${Translation.get('offline')} - ${members.length - onlineCount}`
    }

    static setMemberDisplayName(userID, displayName) {
        displayName = MainClass.checkDisplayName(displayName)

        if (displayName === '') {
            displayName = userID
        }

        this.changeDisplayNameInMemberList(userID, displayName)
    }

    static getMemberName(userID) {
        return document.getElementById(userID).querySelector('div.display-name').textContent
    }

    static setMemberProfilePic(userID, pic) {
        this.changeProfilePicInMemberList(userID, pic)
        console.log(`User ID [${userID}] changed profile pic to [${pic}]`)
    }
}