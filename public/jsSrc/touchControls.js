let touchstartX = 0
let touchendX = 0

let position = 1


const first = document.getElementById("left-main-container")
const second = document.getElementById("main-container")
const third = document.getElementById("member-list")

function checkDirection() {
    const distance = Math.abs(touchstartX - touchendX)
    if (distance < 100) {
        return
    }
    if (touchendX < touchstartX) {
        console.log("swiped left!")
        position = position + 1
        if (position >= 2) {
            position = 2
        }

    }
    if (touchendX > touchstartX) {
        console.log("swiped right!")
        position = position - 1
        if (position <= 1) {
            position = 1
        }
    }
    console.log("Pos: " + position)
    if (position === 1) {
        first.style.display = "flex"
        // second.style.display = "none"
        third.style.display = "none"
    } else if (position === 2) {
        first.style.display = "none"
        // second.style.display = "flex"
        third.style.display = "none"
    } else if (position === 3) {
        first.style.display = "none"
        second.style.display = "none"
        third.style.display = "flex"
    }
}

function showMainLeftContainer() {
    // document.getElementById("left-main-container").style.display = "flex"
}

document.addEventListener('touchstart', e => {
    touchstartX = e.changedTouches[0].screenX
})

document.addEventListener('touchend', e => {
    touchendX = e.changedTouches[0].screenX
    checkDirection()
})

// document.addEventListener('touchmove', e => {
//     console.log(e.changedTouches[0].clientX)
// })