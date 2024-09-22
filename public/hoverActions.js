function registerHover(element, callbackIn, callbackOut) {
    element.addEventListener('mouseover', (event) => {
        // console.log('hovering over', element)
        callbackIn()
    })

    element.addEventListener('mouseout', (event) => {
        // console.log('hovering out', element)
        callbackOut()
    })
}

function serverWhiteThingSize(thing, newSize) {
    thing.style.height = `${newSize}px`
}