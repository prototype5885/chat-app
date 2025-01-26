class LoadingClass {
    static loading = document.getElementById('loading')
    static isLoading = true

    static fadeOutLoading() {
        setTimeout(() => {
            LoadingClass.loading.style.display = 'none'
        }, 250)

        LoadingClass.loading.style.pointerEvents = 'none'
        LoadingClass.loading.style.opacity = '0%'
        this.isLoading = false
    }

    static fadeInLoading() {
        LoadingClass.loading.style.display = 'block'
        LoadingClass.loading.style.opacity = '100%'
        LoadingClass.loading.style.pointerEvents = 'auto'
        LoadingClass.setLoadingText('Connection lost, reconnecting...')
    }

    static setLoadingText(text) {
        LoadingClass.loading.querySelector('div').innerText = text
    }
}
