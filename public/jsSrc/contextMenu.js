class ContextMenuClass {
    static defaultRightClick = false

    static init() {
        // delete context menu if left-clicked somewhere that's not
        // a context menu list element
        document.addEventListener('click', (event) => {
            this.deleteCtxMenu()
        })

        // delete context menu if right-clicked somewhere that's not registered
        // with context menu listener
        document.addEventListener('contextmenu', (event) => {
            if (!this.defaultRightClick) {
                event.preventDefault()
            }
            this.deleteCtxMenu()
        })
    }

    static registerLeftClickContextMenu(element, callback) {
        element.addEventListener('click', (event) => {
            event.preventDefault()
            this.deleteCtxMenu()
            event.stopPropagation()
            callback(event.pageX, event.pageY)
        })
    }

    static registerContextMenu(element, callback) {
        element.addEventListener('contextmenu', (event) => {
            event.preventDefault()
            this.deleteCtxMenu()
            event.stopPropagation()
            callback(event.pageX, event.pageY)
        })
    }

    static createContextMenu(actions, pageX, pageY) {
        if (actions.length === 0) {
            return
        }

        // create the right click menu
        const rightClickMenu = document.createElement('div')
        rightClickMenu.id = 'ctx-menu'
        document.body.appendChild(rightClickMenu)

        // create ul that holds the menu items
        let ul = document.createElement('ul')
        rightClickMenu.appendChild(ul)

        // add a menu item for each action
        actions.forEach((action) => {
            const li = document.createElement('li')
            li.textContent = action.text
            if (action.color === 'red') {
                li.className = 'cm-red' // to make the text red from css
            }
            // this will assign the for each element
            li.onclick = () => {
                action.func()
            }

            ul.appendChild(li)
        })

        // creates the right click menu on cursor position
        rightClickMenu.style.display = 'block'
        rightClickMenu.style.left = `${pageX}px`
        rightClickMenu.style.top = `${pageY}px`
    }

    static deleteCtxMenu() {
        const existingCtxMenus = document.querySelectorAll('#ctx-menu')
        for (let i = 0; i < existingCtxMenus.length; i++) {
            existingCtxMenus[i].remove()
        }
    }

    static pictureCtxMenu(path, name, pageX, pageY) {
        function openPicture() {
            window.open(path, '_blank');
        }

        function savePicture() {
            const link = document.createElement('a');
            link.href = path
            link.download = name

            // Trigger a click event on the <a> element to start the download
            link.click();
        }

        const actions = [
            {text: 'Save', color: '', func: () => savePicture()},
            {text: 'Open in new tab', color: '', func: () => openPicture()},
        ]

        this.createContextMenu(actions, pageX, pageY)
    }

    static serverCtxMenu(serverID, owned, pageX, pageY) {
        const actions = []

        if (owned) {
            console.log(Translation.get('serverSettings'))
            actions.push({
                text: Translation.get('serverSettings'),
                func: () => WindowManagerClass.addWindow('server-settings', serverID)
            })
        }
        if (owned) {
            actions.push({
                text: Translation.get('createInviteLink'),
                func: () => WebsocketClass.requestInviteLink(serverID)
            })
        }
        if (owned) {
            actions.push({
                text: Translation.get('deleteServer'),
                color: 'red',
                func: () => WebsocketClass.requestDeleteServer(serverID)
            })
        }
        if (!owned) {
            actions.push({
                text: Translation.get('leaveServer'),
                color: 'red',
                func: () => WebsocketClass.requestLeaveServer(serverID)
            })
        }

        this.createContextMenu(actions, pageX, pageY)
    }

    static channelCtxMenu(channelID, owned, pageX, pageY) {
        const actions = []
        if (owned) {
            actions.push({
                text: Translation.get('channelSettings'),
                func: () => WindowManagerClass.addWindow('channel-settings', channelID)
            })
        }
        if (owned) {
            actions.push({
                text: Translation.get('deleteChannel'),
                color: 'red',
                func: async () => deleteChannel(channelID)
            })
        }

        async function deleteChannel(channelID) {
            await WebsocketClass.requestRemoveChannel(channelID)
        }

        this.createContextMenu(actions, pageX, pageY)
    }

    static userCtxMenu(userID, pageX, pageY) {
        // function reportUser(userID) {
        //     console.log('Reporting user', userID)
        // }


        const actions = []

        // if (userID !== MainClass.getOwnUserID()) {
        //     actions.push({
        //         text: 'Message', func: async () => {
        //             console.log(`Messaging user ID ${userID}`)
        //             await WebsocketClass.requestOpenDm(userID)
        //             await ServerListClass.selectServer('1')
        //         }
        //     })
        // }
        //
        // if (!MainClass.myFriends.includes(userID) && userID !== MainClass.getOwnUserID()) {
        //     actions.push({text: 'Add friend', func: async () => WebsocketClass.requestAddFriend(userID)})
        // }
        // if (MainClass.myFriends.includes(userID) && userID !== MainClass.getOwnUserID()) {
        //     actions.push({
        //         text: 'Remove friend',
        //         color: 'red',
        //         func: async () => WebsocketClass.requestUnfriend(userID)
        //     })
        // }
        // if (userID !== MainClass.getOwnUserID()) {
        //     actions.push({text: 'Block', color: 'red', func: async () => WebsocketClass.requestBlockUser(userID)})
        // }
        // if (userID !== MainClass.myUserID) {
        //     actions.push({text: 'Report user', color: 'red', func: () => reportUser(userID)})
        // }
        actions.push({
            text: Translation.get('copyUserID'), func: () => {
                console.log('Copying user ID', userID)
                navigator.clipboard.writeText(userID).then(r => '')
            }
        })
        actions.push({
            text: Translation.get('mentionUser'), func: () => {
                console.log('Mentioning user ID', userID)
                const chatInput = document.getElementById('chat-input')
                chatInput.value = `<@${userID}>`
                chatInput.dispatchEvent(new Event('input'));
                chatInput.focus()
            }
        })

        this.createContextMenu(actions, pageX, pageY)
    }

    static messageCtxMenu(messageID, owner, pageX, pageY) {
        const actions = []
        actions.push({
            text: Translation.get('copyChatMessage'), func: async () => {
                const chatMsg = document.getElementById(messageID).querySelector('.msg-text').textContent
                console.log('Copied to clipboard:', chatMsg)
                await navigator.clipboard.writeText(chatMsg)
            }
        })

        actions.push({
            text: Translation.get('reply'), func: async () => {
                const msg = document.getElementById(messageID)
                console.log(`Replying to message ID [${msg.id}]`)

                ChatInputClass.openReplyContainer(msg.id, msg.getAttribute('user-id'))
            }
        })

        if (owner) {
            actions.push({
                text: Translation.get('editChatMessage'), func: () => {
                    const chatMsg = document.getElementById(messageID).querySelector('.msg-text')
                    const msgData = document.getElementById(messageID).querySelector('.msg-data')

                    // hide the original chat message
                    chatMsg.style.display = 'none'

                    // hide the (edited) text if it exists
                    const msgEdited = msgData.querySelector('.msg-edited')
                    if (msgEdited !== null) {
                        msgEdited.style.display = 'none'
                    }

                    const container = document.createElement('div')
                    container.className = 'edit-chat-msg-container'

                    msgData.appendChild(container)

                    const form = document.createElement('div')
                    form.className = 'edit-chat-msg-form'

                    const textArea = document.createElement('textarea')
                    textArea.className = 'edit-chat-msg'

                    textArea.textContent = chatMsg.textContent
                    form.appendChild(textArea)

                    container.appendChild(form)

                    const label = document.createElement('label')
                    label.style.fontSize = '11px'
                    label.innerHTML = `escape to <a href='#' id='cancel-edit-link''>cancel</a>, enter to <a href='#' id='send-edit-link''>save</a>`

                    container.appendChild(label)

                    const msgTextContainer = msgData.querySelector('.msg-text-container')
                    msgTextContainer.insertAdjacentElement('afterend', container)


                    textArea.focus()
                    let length = textArea.value.length
                    textArea.setSelectionRange(length, length)

                    function resize() {
                        textArea.style.height = 'auto'
                        textArea.style.height = textArea.scrollHeight + 'px'
                    }

                    textArea.addEventListener('input', () => {
                        resize()
                    })

                    async function sendEditedMessage() {
                        if (chatMsg.textContent === textArea.value.trim()) {
                            console.log('Edited message has no difference, cancelling...')
                        } else {
                            await WebsocketClass.requestEditChatMessage(messageID, textArea.value)
                        }
                        reset()
                    }

                    function reset() {
                        container.remove()
                        chatMsg.style.display = 'block'

                        const msgEdited = msgData.querySelector('.msg-edited')
                        if (msgEdited !== null) {
                            msgEdited.style.display = 'block'
                        }
                    }

                    textArea.addEventListener('keydown', async (event) => {
                        if (event.key === 'Enter' && !event.shiftKey) {
                            event.preventDefault()
                            await sendEditedMessage()
                        } else if (event.key === 'Escape') {
                            event.preventDefault()
                            reset()
                        }
                    })

                    document.getElementById('cancel-edit-link').addEventListener('click', (event) => {
                        event.preventDefault();
                        reset()
                    })
                    document.getElementById('send-edit-link').addEventListener('click', async (event) => {
                        event.preventDefault();
                        await sendEditedMessage()
                    })

                    resize()
                }
            })
        }

        if (owner) {
            actions.push({
                text: Translation.get('deleteChatMessage'),
                color: 'red',
                func: () => WebsocketClass.requestDeleteChatMessage(messageID)
            })
        }

        // if (!owner) {
        //     actions.push({text: 'Report message', color: 'red'})
        // }
        this.createContextMenu(actions, pageX, pageY)
    }
}

