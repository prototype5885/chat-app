class ServerListClass {
    static ServerList = document.getElementById('first-column')
    static serverSeparators = ServerListClass.ServerList.querySelectorAll('.servers-separator')
    static AddServerButton = document.getElementById('add-server-button')

    static init() {
        // add bubble when hovering over add server button
        MainClass.registerHover(ServerListClass.AddServerButton, () => {
            BubbleClass.createBubble(ServerListClass.AddServerButton, 'Add Server', 'right', 15)
        }, () => {
            BubbleClass.deleteBubble()
        })

        // hide notification marker as this doesn't use it,
        // but it's needed for formatting reasons
        ServerListClass.AddServerButton.nextElementSibling.style.backgroundColor = 'transparent'

        ServerListClass.AddServerButton.addEventListener('click', async () => {
            await WebsocketClass.requestAddServer('server')
        })
    }

    static createPlaceHolderServers() {
        console.log('Adding placeholder servers')
        ServerListClass.removePlaceholderServers()
        const serverCount = LocalStorageClass.getServerCount()
        if (serverCount !== 0) {
            for (let i = 0; i < serverCount; i++) {
                ServerListClass.addPlaceholderServer()
            }
        }
        ServerListClass.calculateServerAmount()
    }

    static removePlaceholderServers() {
        console.log('Removing placeholder servers')
        // remove placeholder servers
        const placeholderButtons = ServerListClass.ServerList.querySelectorAll('.placeholder-server')
        for (let i = 0; i < placeholderButtons.length; i++) {
            placeholderButtons[i].remove()
        }
    }

    static addPlaceholderServer() {
        const buttonParent = ServerListClass.addServer('', 0, '', '', 'placeholder-server')
        let button = buttonParent.querySelector('button')
        button.nextElementSibling.style.backgroundColor = 'transparent'
        button.textContent = ''
    }

    static addServer(serverID, owned, serverName, picture, className) {
        if (serverName === '') {
            serverName = serverID
        }

        // this li will hold the server and notification thing, which is the span
        const li = document.createElement('li')
        li.className = className
        ServerListClass.ServerList.append(li)

        // create the server button itself
        const button = document.createElement('button')
        button.id = serverID
        button.setAttribute('name', serverName)
        li.append(button)

        // set picture of server
        if (picture !== '') {
            if (serverID !== '1') {
                picture = 'content/avatars/' + picture
            }
            button.style.backgroundImage = `url(${picture})`
        } else {
            if (serverName !== '') {
                button.textContent = serverName[0].toUpperCase()
            }
        }

        const span = document.createElement('span')
        span.className = 'server-notification'
        li.append(span)

        button.setAttribute('owned', owned)

        MainClass.registerClick(button, async () => {
            await this.selectServer(serverID)
        })
        ContextMenuClass.registerContextMenu(button, (pageX, pageY) => {
            ContextMenuClass.serverCtxMenu(serverID, owned, pageX, pageY)
        })
        MainClass.registerHover(button, () => {
            if (serverID !== MainClass.getCurrentServerID()) {
                button.style.borderRadius = '35%'
                button.style.backgroundColor = '#5865F2'
                span.style.height = '24px'
            }
            BubbleClass.createBubble(button, ServerListClass.getServerName(serverID), 'right', 15)
        }, () => {
            if (serverID !== MainClass.getCurrentServerID()) {
                button.style.borderRadius = '50%'
                button.style.backgroundColor = ''
                span.style.height = '8px'
            }
            BubbleClass.deleteBubble()
        })

        // this check needs to be made else adding placeholder servers will break serverCount value,
        // as it would reset the serverCount value while adding placeholders, as fix serverSeparatorVisibility
        // is run manually only after creating each placeholder servers on startup
        if (className === 'server') {
            ServerListClass.calculateServerAmount()
        }

        return li
    }

    static removeServers() {
        document.querySelectorAll('.server').forEach(server => {
            server.remove()
        })
    }

    static async selectServer(serverID) {
        // if (serverID === '1') {
        //     console.log('Selected direct messages')
        // } else {
        //     console.log('Selected server ID', serverID, ', requesting list of channels...')
        // }


        // check if selected server is already the current one
        if (serverID === MainClass.getCurrentServerID()) {
            console.warn('Selected server is already the current one')
            return
        }


        // DirectMessagesClass.DmChatList.innerHTML = ''
        // ChannelListClass.ChannelList.innerHTML = ''

        // if (serverID === '1') {
        //     DirectMessagesClass.DirectMessages.removeAttribute('style')
        //     ChannelListClass.Channels.style.display = 'none'
        //     document.getElementById('channel-name-top').textContent = 'Friends'
        // } else {
        //     ChannelListClass.Channels.removeAttribute('style')
        //     DirectMessagesClass.DirectMessages.style.display = 'none'
        // }

        MainClass.memberListLoaded = false

        // this will reset the previously selected server's visuals
        const previousServerButton = document.getElementById(MainClass.getCurrentServerID())
        if (previousServerButton != null) {
            previousServerButton.nextElementSibling.style.height = '8px'
            previousServerButton.style.backgroundColor = ''
            previousServerButton.style.borderRadius = '50%'
        }

        MainClass.setCurrentServerID(serverID)

        // serverButton.nextElementSibling.style.height = '36px'

        // if (serverID !== '1') {
        //     const owned = ServerListClass.getServerOwned(serverID)
        //
        //     // hide add channel button if server isn't own
        //     if (owned === 'true') {
        //         ChannelListClass.AddChannelButton.style.display = 'block'
        //     } else {
        //         ChannelListClass.AddChannelButton.style.display = 'none'
        //     }
        // }


        // if (dm) {
        //     this.memberList.hideMemberList()
        // } else {
        //     this.memberList.showMemberList()
        // }

        SecondColumnMainClass.reset()
        ChatMessageListClass.resetChatMessages()
        MemberListClass.resetMemberList()

        ChannelListClass.createChannelList()

        if (serverID !== '1') {
            await WebsocketClass.requestChannelList()
            await WebsocketClass.requestMemberList()
        } else {
            await WebsocketClass.requestDmList()
        }
        const serverButton = document.getElementById(serverID)

        ServerBannerClass.setName(serverButton.getAttribute('name'))

        LocalStorageClass.setLastServer(serverID)

        // const bannerUrl = 'https://cdn.discordapp.com/banners/1267683587902279742/adb469683ec356db30b42f0e5bccba01.webp?size=480'
        const bannerUrl = ''

        ServerBannerClass.setPicture(bannerUrl)

        // TouchControlsClass.goRight()
    }

    static deleteServer(serverID) {
        console.log('Deleting server ID:', serverID)
        // check if class is correct
        document.getElementById(serverID).parentNode.remove()
        ServerListClass.calculateServerAmount()
    }

    static setServerPicture(serverID, picture) {
        const serverButton = document.getElementById(serverID)
        if (serverButton == null) {
            console.error(`Server ID ${serverID} button doesn't exist, can't set server picture`)
            return
        }
        picture = 'public/content/avatars/' + picture
        serverButton.style.backgroundImage = `url('${picture}')`
    }

    static setServerName(serverID, name) {
        console.log(`Changing name of server ID [${serverID}] to [${name}]`)
        const server = document.getElementById(serverID)
        server.setAttribute('name', name)
        if (serverID === MainClass.getCurrentServerID()) {
            ServerBannerClass.setName(name)
        }
        if (server.style.backgroundImage === '') {
            server.textContent = name[0].toUpperCase()
        }
    }

    static getServerName(serverID) {
        return document.getElementById(serverID).getAttribute('name')
    }

    static calculateServerAmount() {
        const allServers = ServerListClass.ServerList.querySelectorAll('.server, .placeholder-server')
        const servers = ServerListClass.ServerList.querySelectorAll('.server')
        LocalStorageClass.setServerCount(servers.length)

        if (allServers.length !== 0) {
            ServerListClass.serverSeparators.forEach((separator) => {
                separator.style.display = 'block'
            })
        } else {
            ServerListClass.serverSeparators.forEach((separator) => {
                separator.style.display = 'none'
            })
        }
    }

    static getServerOwned(serverID) {
        console.log(`Getting if server ID [${serverID}] is owned by me`)
        const server = document.getElementById(serverID)
        if (server === null) {
            console.error(`Server ID [${serverID}] doesn't exist`)
        } else {
            return server.getAttribute('owned')
        }

    }

    serverWhiteThingSize(thing, newSize) {
        thing.style.height = `${newSize}px`
    }
}