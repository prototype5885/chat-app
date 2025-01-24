class MemberListClass {
    static MemberList = document.getElementById('fourth-column-main')

    static addMember(userID, displayName, picture, online, status, statusText) {
        // create a <li> that holds the user
        const li = document.createElement('li')
        li.className = 'member'
        li.id = userID

        const picContainer = document.createElement('div')
        picContainer.className = 'profile-pic-container'
        picContainer.style.width = '32px'
        picContainer.style.height = '32px'

        if (picture === '') {
            picture = '/content/static/default_profilepic.webp'
        } else {
            picture = MainClass.getAvatarFullPath(picture)
        }

        // create a <img> that shows profile pic on the left
        const img = document.createElement('img')
        img.className = 'profile-pic'
        img.src = picture
        img.width = 32
        img.height = 32

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
        userNameDiv.className = 'user-name'
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
        this.MemberList.appendChild(li)

        this.changeStatusValueInMemberList(userID, status)
        this.setMemberOnline(userID, online)
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
            const username = member.querySelector('div.user-name').textContent
            return {username: username, pic: pic}
        } else {
            return {username: userID, pic: ''}
        }
    }

    static toggleMemberListView() {
        if (this.MemberList.style.display === 'none') {
            this.showMemberList()
        } else {
            this.hideMemberList()
        }
    }

    static hideMemberList() {
        this.MemberList.style.display = 'none'
    }

    static showMemberList() {
        this.MemberList.style.display = 'block'
    }

    static resetMemberList() {
        this.MemberList.innerHTML = ''
    }

    static changeDisplayNameInMemberList(userID, newDisplayName) {
        try {
            const user = document.getElementById(userID)
            user.querySelector('.user-name').textContent = newDisplayName
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


    static findMember(userID) {
        return document.getElementById(userID)
    }

    static setMemberStatusText(userID, newStatusText) {
        const userStatusText = this.findMember(userID).querySelector('.user-status-text')
        userStatusText.textContent = newStatusText
    }

    static setMemberOnline(userID, online) {
        const userStatus = document.getElementById(userID).querySelector('.profile-pic-container').querySelector('.user-status')
        const member = this.findMember(userID)
        if (online) {
            member.removeAttribute('style')
            userStatus.style.display = 'block'
        } else {
            member.style.filter = 'grayscale(100%)'
            member.style.opacity = '0.5'
            userStatus.style.display = 'none'
        }
    }

    static setMemberDisplayName(userID, displayName) {
        displayName = MainClass.checkDisplayName(displayName)

        if (displayName === '') {
            displayName = userID
        }

        this.changeDisplayNameInMemberList(userID, displayName)
    }

    static getMemberName(userID) {
        return document.getElementById(userID).querySelector('div.user-name').textContent
    }

    static setMemberProfilePic(userID, pic) {
        this.changeProfilePicInMemberList(userID, pic)
        console.log(`User ID [${userID}] changed profile pic to [${pic}]`)
    }
}