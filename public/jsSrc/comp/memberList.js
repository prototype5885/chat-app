function addMember(userID, displayName, picture, status, statusText) {
    // create a <li> that holds the user
    const li = document.createElement("li")
    li.className = "member"
    li.id = userID

    const picContainer = document.createElement("div")
    picContainer.className = "profile-pic-container"
    picContainer.style.width = "32px"
    picContainer.style.height = "32px"

    // create a <img> that shows profile pic on the left
    const img = document.createElement("img")
    img.className = "profile-pic"
    img.src = getAvatarFullPath(picture)
    img.width = 32
    img.height = 32

    // create a <div> that will be a circle in the corner of profile pic to show online status
    // const status = document.createElement("div")
    // status.className = "user-status"

    picContainer.appendChild(img)
    // picContainer.appendChild(status)

    // create a nested <div> that will contain username and status
    const userDataDiv = document.createElement("div")
    userDataDiv.className = "user-data"

    // create <div> that will hold the user"s message
    const userNameDiv = document.createElement("div")
    userNameDiv.className = "user-name"
    userNameDiv.textContent = displayName
    userNameDiv.style.color = grayTextColor

    // now create a <div> under name that display statis
    const userStatusDiv = document.createElement("div")
    userStatusDiv.className = "user-status-text"
    userStatusDiv.textContent = statusText

    // append both name/date <div> and msg <div> to msgDatDiv
    userDataDiv.appendChild(userNameDiv)
    userDataDiv.appendChild(userStatusDiv)

    // append both the profile pic and message data to the <li>
    li.appendChild(picContainer)
    li.appendChild(userDataDiv)

    // and finally append the message to the message list
    MemberList.appendChild(li)

    changeStatusValueInMemberList(userID, status)
}

function removeMember(userID) {
    const element = document.getElementById(userID)
    if (element.className === "member") {
        element.remove()
    } else {
        console.log(`Trying to remove member ID [${userID}] but the element is not member class: [${element.className}]`)
    }
}

function getUserInfo(userID) {
    const member = document.getElementById(userID)
    if (member != null) {
        pic = member.querySelector('img.profile-pic').src
        username = member.querySelector('div.user-name').textContent
        return { username: username, pic: pic }
    } else {
        return { username: userID, pic: "" }
    }
}

function toggleMemberListView() {
    if (MemberList.style.display === "none") {
        showMemberList()
    } else {
        hideMemberList()
    }
}

function hideMemberList() {
    MemberList.style.display = "none"
}

function showMemberList() {
    MemberList.style.display = "block"
}

function resetMemberList() {
    MemberList.innerHTML = ""
}

function changeDisplayNameInMemberList(userID, newDisplayName) {
    const user = document.getElementById(userID)
    user.querySelector(".user-name").textContent = newDisplayName
}

function changeProfilePicInMemberList(userID, newPicture) {
    const user = document.getElementById(userID)
    user.querySelector(".profile-pic").src = getAvatarFullPath(newPicture)
}

function changeStatusValueInMemberList(userID, newStatus) {
    const container = document.getElementById(userID).querySelector(".profile-pic-container")
    const currentStatus = container.querySelector(".user-status")

    if (currentStatus) {
        currentStatus.remove()
    }

    const status = document.createElement("div")
    status.className = "user-status"

    switch (newStatus) {
        case 1:
            status.style.backgroundColor = "limegreen"
            break
        case 2:
            const boolean = document.createElement("div")
            boolean.className = "orange-status-boolean"
            status.style.backgroundColor = "orange"
            status.appendChild(boolean)
            break
        case 3:
            status.style.backgroundColor = "red"
            break
        case 4:
            break
        default:
            status.remove()
            return
    }
    container.appendChild(status)
}

function changeStatusTextInMemberList(userID, newStatusText) {
    const userStatusText = document.getElementById(userID).querySelector(".user-status-text")
    userStatusText.textContent = newStatusText
}