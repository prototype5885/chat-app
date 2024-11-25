function fadeOutLoading() {
    const loading = document.getElementById("loading")
    setTimeout(() => {
        loading.style.display = "none"
    }, 250)

    loading.style.pointerEvents = "none"
    loading.style.opacity = "0%"
}

function fadeInLoading() {
    const loading = document.getElementById("loading")
    loading.style.display = "block"
    loading.style.opacity = "100%"
    loading.style.pointerEvents = "auto"
    loading.querySelector("div").innerText = "Connection lost, reconnecting..."
}