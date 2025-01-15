class BubbleClass {
    static createBubble(element, text, direction, distance) {
        const content = document.createElement("div")
        content.textContent = text

        // create bubble div that will hold the content
        const bubble = document.createElement("div")
        bubble.className = "bubble"
        document.body.appendChild(bubble)

        // add the content into it
        bubble.appendChild(content)

        // center of the element that created the bubble,
        // bubble will be created relative to this
        const rect = element.getBoundingClientRect()

        const center = {
            x: rect.left + rect.width / 2 + window.scrollX,
            y: rect.top + rect.height / 2 + window.scrollY
        }

        const height = bubble.getBoundingClientRect().height
        const width = bubble.getBoundingClientRect().width

        switch (direction) {
            case "right":
                // get how tall the bubble will be, so can
                // offset the Y position to make it appear
                // centered next to the element


                // set the bubble position
                bubble.style.left = `${(center.x + element.clientWidth / 2) + distance}px`
                bubble.style.top = `${center.y - height / 2}px`
                break
            case "up":

                bubble.style.left = `${center.x - width / 2}px`
                bubble.style.top = `${(center.y - element.clientHeight - (element.clientHeight / 2) - distance)}px`
                break
        }
    }

    static deleteBubble() {
        const bubbles = document.querySelectorAll(".bubble")

        if (bubbles.length !== 1) {
            console.warn(`For some reason there are [${bubbles.length}] bubbles to be deleted, while there should be only 1`)
        }

        bubbles.forEach(bubble => {
            bubble.remove()
        })
    }
}

