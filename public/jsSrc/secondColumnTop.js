class SecondColumnTopClass {
}

class ServerBannerClass {
    static #ServerNameContainer = document.getElementById('server-name-container')
    static #ServerName = document.getElementById('server-name')


    static init() {
        const serverNameButton = document.getElementById('server-name-button')
        ContextMenuClass.registerLeftClickContextMenu(serverNameButton, () => {
            const rect = serverNameButton.getBoundingClientRect()

            const absoluteBottom = (rect.bottom + window.scrollY) + 8
            const absoluteCenter = (rect.left + window.scrollX) + (rect.width / 2) - 75

            const serverID = MainClass.getCurrentServerID()
            ContextMenuClass.serverCtxMenu(serverID, ServerListClass.getServerOwned(serverID), absoluteCenter, absoluteBottom)
        })
    }

    static setName(name) {
        this.#ServerName.textContent = name
    }

    static setPicture(serverID, picPath) {
        if (serverID !== MainClass.getCurrentServerID()) {
            console.warn(`Won't set banner for server ID [${serverID}] as it's not the currently selected server`)
            return
        }

        if (picPath !== '') {
            picPath = '/content/banners/' + picPath
            const img = new Image();
            img.src = picPath

            img.onload = () => {
                const newHeight = (this.#ServerNameContainer.offsetWidth / img.width) * img.height;
                this.#ServerNameContainer.style.height = `${newHeight}px`
                this.#ServerNameContainer.style.backgroundImage = `url(${picPath})`
                this.#ServerName.style.textShadow = '1px 1px 1px black'
                document.getElementById('server-name-button-container').style.backgroundColor = 'rgba(0, 0, 0, 0.0)'
            }
        } else {
            this.#ServerNameContainer.style.height = ''
            this.#ServerNameContainer.style.backgroundImage = ''
            // this.#ServerNameButton.style.backgroundColor = ColorsClass.bitDarkerColor
            this.#ServerName.style.textShadow = ''
            document.getElementById('server-name-button-container').style.backgroundColor = ColorsClass.bitDarkerColor
        }
    }
}