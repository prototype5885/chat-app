class TouchControlsClass {
    static currentIndex = 0
    static nextIndex = 0

    static pastHalf = false
    static direction = ''

    static boxes = document.getElementById('pages').children

    static offsetX = 0
    static touchStartX = 0
    static swiping = false

    static init() {
        document.addEventListener('touchstart', e => {
            this.touchStartX = e.changedTouches[0].clientX
            console.log('started swiping')
        })

        document.addEventListener('touchmove', e => {
            if (!this.swiping) {
                if (e.changedTouches[0].clientX < this.touchStartX) {
                    this.direction = 'left'
                } else {
                    this.direction = 'right'
                }

                if (this.direction === 'left') {
                    this.nextIndex = this.currentIndex + 1
                } else if (this.direction === 'right') {
                    this.nextIndex = this.currentIndex - 1
                }

                this.nextIndex = Math.max(0, Math.min(this.nextIndex, this.boxes.length - 1))
                this.offsetX = e.changedTouches[0].clientX - this.boxes[this.nextIndex].getBoundingClientRect().left

                this.swiping = true
            }

            if (this.currentIndex === 0 && this.nextIndex === 0) {
                return
            }

            const width = this.boxes[this.nextIndex].getBoundingClientRect().width

            let posX = e.changedTouches[0].clientX - this.offsetX
            posX = Math.max(0, Math.min(posX, width))


            if (this.direction === 'left') {
                this.boxes[this.nextIndex].style.left = `${posX}px`
            } else if (this.direction === 'right') {
                this.boxes[this.currentIndex].style.left = `${posX}px`
            }

            const halfPointX = width / 2

            if (posX < halfPointX) {
                this.pastHalf = true
            } else {
                this.pastHalf = false
            }
        })

        document.addEventListener('touchend', e => {
            this.swipe()
        })


    }

    static swipe() {
        // if ((Date.now() - this.startedSwipingTime) < 500) {
        //     if (this.direction === 'left') {
        //         this.pastHalf = true
        //     } else if (this.direction === 'right') {
        //         this.pastHalf = false
        //     }
        // }

        if (this.direction === 'left') {
            if (this.pastHalf) {
                this.pastHalf = false
                this.animateElement(this.boxes[this.nextIndex], 0)
                this.currentIndex = this.nextIndex
            } else {
                const nextElement = this.boxes[this.nextIndex]
                const nextElementWidth = this.boxes[this.nextIndex].getBoundingClientRect().width
                this.animateElement(nextElement, nextElementWidth)
            }
        } else if (this.direction === 'right') {
            if (!this.pastHalf) {
                this.pastHalf = false

                const currentElement = this.boxes[this.currentIndex]
                const currentElementWidth = this.boxes[this.currentIndex].getBoundingClientRect().width
                this.animateElement(currentElement, currentElementWidth)

                this.animateElement(this.boxes[this.nextIndex], 0)
                this.currentIndex = this.nextIndex
            } else {
                this.animateElement(this.boxes[this.nextIndex], 0)
                this.animateElement(this.boxes[this.currentIndex], 0)
            }
        }
        console.log('Next index: ', this.nextIndex)
        this.swiping = false
        this.direction = ''
        this.offsetX = 0
        this.touchStartX = 0
    }

    static goRight() {

    }

    static goLeft() {

    }

    static animateElement(element, endPosition) {
        const startPosition = parseFloat(getComputedStyle(element).left)
        const start = performance.now()
        const duration = 100


        function move() {
            const elapsedTime = performance.now() - start
            const progress = Math.min(elapsedTime / duration, 1)

            const currentPosition = startPosition + (parseFloat(endPosition) - startPosition) * progress

            element.style.left = `${currentPosition}px`

            if (elapsedTime < duration) {
                requestAnimationFrame(move)
            }
        }

        requestAnimationFrame(move)
    }
}