class ChatInputClass {
    constructor(main) {
        this.main = main
        // this.AttachmentContainer = document.getElementById("attachment-container")
        this.AttachmentList = document.getElementById("attachment-list")

        this.typing = false

        this.ChatInput = document.getElementById("chat-input")
        this.ChatInput.addEventListener("keydown", this.chatEnterPressed.bind(this))

        this.ChatInput.addEventListener("input", () => {
            this.resizeChatInput()
            this.checkIfTyping()
        })

        document.getElementById("attachment-button").addEventListener("click", () => {
            this.AttachmentInput.click()
        })

        // this is when user clicks on attachment button and uploads files from there
        this.AttachmentInput = document.getElementById("attachment-input")
        this.AttachmentInput.addEventListener("change", () => {
            for (let i = 0; i < this.AttachmentInput.files.length; i++) {
                this.addAttachment(this.AttachmentInput.files[i])
            }
        })

        this.fileDropZone = document.getElementById("file-drop-zone")
        this.fileDropMsg = document.getElementById("file-drop-msg")

        document.addEventListener("dragover", e => {
            e.preventDefault()
            this.fileDropZone.style.display = "flex"
            this.fileDropMsg.textContent = "Upload to:\n\n" + this.main.currentChannelID
        })

        // this when user drags files into webpage
        this.fileDropZone.addEventListener("drop", e => {
            e.preventDefault()
            console.log("dropped file")

            for (let i = 0; i < e.dataTransfer.items.length; i++) {
                const file = e.dataTransfer.items[i]
                if (e.dataTransfer.items[i].kind === "file") {
                    const file = e.dataTransfer.items[i].getAsFile();
                    this.addAttachment(file)
                }
            }
            this.hideFileDropUI()
        })

        this.fileDropZone.addEventListener("dragenter", e => {
            e.preventDefault()
            console.log("Started dragging a file into window")
        })

        this.fileDropZone.addEventListener("dragleave", e => {
            e.preventDefault()
            console.log("Cancelled file dragging")
            this.hideFileDropUI()
        })

        document.addEventListener("paste", e => {
            const items = e.clipboardData.items
            if (items) {
                for (let i = 0; i < items.length; i++) {
                    const item = items[i]

                    // Only handle files
                    if (item.kind === 'file') {
                        const file = item.getAsFile()
                        this.addAttachment(file)
                    }
                }
            }
        })

        this.maxFiles = 5
        this.files = []

        this.resizeChatInput()
    }

    // dynamically resize the chat input textarea to fit the text content
    // runs whenever the chat input textarea content changes
    // or pressed enter
    resizeChatInput() {
        this.ChatInput.style.height = "auto"
        this.ChatInput.style.height = this.ChatInput.scrollHeight + "px"
    }

    checkIfTyping() {
        if (this.ChatInput.value !== "" && !this.typing) {
            this.typing = true
            console.log("started typing")
            WebsocketClass.startedTyping(true)
        }
        if (this.ChatInput.value === "") {
            this.typing = false
            console.log("stopped typing")
            WebsocketClass.startedTyping(false)
        }
    }

    // send the text message on enter
    async chatEnterPressed(event) {
        if (event.key === "Enter" && !event.shiftKey) {
            event.preventDefault()
            let attachmentToken = null
            if (this.files.length !== 0) {
                console.log(`Chat message has [${this.files.length}] attachments, sending those first...`)

                // these hashes are of the attachments that already exist on server, no need to upload them
                const existingHashes = await this.checkAttachments()

                attachmentToken = await this.sendAttachment(existingHashes)
                console.log("http response to uploading attachment:", attachmentToken)
            }

            if (this.ChatInput.value || attachmentToken !== null) {
                if (attachmentToken !== null) {
                    await WebsocketClass.sendChatMessage(this.ChatInput.value, main.currentChannelID, attachmentToken.AttToken)
                } else {
                    await WebsocketClass.sendChatMessage(this.ChatInput.value, main.currentChannelID, null)
                }
                console.log("Resetting chat input and attachment input values")
                this.ChatInput.value = ""
                this.AttachmentInput.value = ""
                this.resizeChatInput()
                // this.checkIfTyping()
            }
            this.typing = false
        }
    }


    async checkAttachments() {
        console.log("Checking if prepare attachments already exist on server")

        let hashes = []
        for (let i = 0; i < this.AttachmentInput.files.length; i++) {
            const hash = await MainClass.calculateSHA256(this.AttachmentInput.files[i])
            hashes.push(hash)
        }

        const xhr = new XMLHttpRequest()

        return new Promise((resolve, reject) => {
            xhr.onload = function () {
                if (xhr.status === 200) {
                    const existingHashes = JSON.parse(xhr.responseText)
                    if (existingHashes === null) {
                        console.log("All attachments need to be uploaded")
                        resolve(null)
                    } else {
                        console.log(`[${existingHashes.length}] attachments don't need to be uploaded`)
                        resolve(existingHashes)
                    }

                } else {
                    console.error("Failed asking the server if given attachment hashes exist")
                    reject(null)
                }
            }


            xhr.onerror = function () {
                console.error("Error asking the server if given attachment hashes exist")
                reject(null)
            }

            xhr.open("POST", "/check-attachment")
            xhr.setRequestHeader('Content-Type', 'application/json')
            xhr.send(JSON.stringify(hashes))
        })
    }

    async sendAttachment(existingHashes) {
        console.log("Sending attachments to server")
        const formData = new FormData()

        // loops through added attachments
        for (let i = 0; i < this.files.length; i++) {
            if (i > this.maxFiles - 1) {
                console.warn("Too many attachments, ignoring those after 4th...")
                continue
            }

            console.log(`Preparing attachment index [${i}] called [${this.files[i].name}] for sending`)
            const hash = await MainClass.calculateSHA256(this.files[i])

            let exists = false
            if (existingHashes === null) {
                console.warn(`existingHashes is null, uploading attachment index [${i}]`)
                exists = false
            } else {
                for (let h = 0; h < existingHashes.length; h++) {
                    console.log(`Comparing [${hash}] with [${existingHashes[h]}]`)
                    if (MainClass.areArraysEqual(hash, existingHashes[h])) {
                        exists = true
                        break
                    }
                }
            }

            if (!exists) {
                console.log(`Attachment index [${i}] doesn't exist on server, uploading...`)
                formData.append("a", this.files[i])
            } else {
                console.log(`Attachment index [${i}] exists on server, sending hash only...`)
                const name = this.files[i].name
                const jsonString = JSON.stringify({Hash: hash, Name: name})
                formData.append("h", jsonString)
            }
        }

        const xhr = new XMLHttpRequest()

        return new Promise((resolve, reject) => {
            xhr.onload = () => {
                if (xhr.status === 200) {
                    const attachmentToken = JSON.parse(xhr.responseText)
                    console.log("Attachment was uploaded successfully")
                    this.resetAttachments()
                    this.calculateAttachments()
                    resolve(attachmentToken)
                } else {
                    console.error("Failed asking the server if given attachment hashes exist")
                    reject(null)
                }
            }

            xhr.onloadstart = function () {
                console.log("Starting upload...")
            }
            xhr.onloadend = function () {
                console.log("Finished upload")
            }

            xhr.upload.onprogress = async function (e) {
                console.log(e.loaded, e.total)
                if (e.lengthComputable) {
                    const indicator = document.getElementById("upload-percentage")
                    let percent = (e.loaded / e.total) * 100

                    percent = Math.round(percent)
                    indicator.textContent = percent.toString() + " %"
                    if (percent >= 100) {
                        indicator.textContent = ""
                    }
                }
            }


            xhr.onerror = function () {
                console.error("Error asking the server if given attachment hashes exist")
                reject(null)
            }

            xhr.open("POST", "/upload-attachment")
            xhr.send(formData)
        })

    }

    resetAttachments() {
        console.log("Resetting attachments")
        this.AttachmentList.innerHTML = ""
        this.files = []
    }

    addAttachment(entry) {
        if (this.files.length >= this.maxFiles) {
            console.warn("Too many attachments, ignoring those after 4th...")
            return
        }
        this.files.push(entry)
        console.log(`Added attachment [${entry.name}], current attachment count: [${this.files.length}]`)

        const reader = new FileReader()
        reader.readAsDataURL(entry)

        // when the file is loaded into the browser
        reader.onload = (e) => {
            const attachmentContainer = document.createElement("div")
            this.AttachmentList.appendChild(attachmentContainer)

            // when clicked on the attachment, it removes it
            attachmentContainer.addEventListener("click", () => {
                attachmentContainer.remove()
                this.removeAttachment(entry)
                if (this.AttachmentList.length <= 0) {
                    this.AttachmentInput.value = ""
                }
                console.log(`Removed attachment [${entry.name}], current attachment count: [${this.files.length}]`)
                this.calculateAttachments()
            })

            const text = false

            const attachmentPreview = document.createElement("div")
            attachmentPreview.className = "attachment-preview"
            if (text) {
                attachmentContainer.style.height = "224px"
            } else {
                attachmentContainer.style.height = "200px"
            }
            const imgElement = document.createElement("img")
            imgElement.src = e.target.result
            imgElement.style.display = 'block'
            attachmentPreview.appendChild(imgElement)
            attachmentContainer.appendChild(attachmentPreview)

            if (text) {
                const attachmentName = document.createElement("div")
                attachmentName.className = "attachment-name"
                attachmentName.textContent = "test.jpg"
                attachmentContainer.appendChild(attachmentName)
            }
            this.calculateAttachments()
        }
    }

    removeAttachment(entry) {
        this.files.splice(this.files.indexOf(entry), 1)
    }

    hideFileDropUI() {
        this.fileDropZone.style.display = "none"
    }

    calculateAttachments() {
        const count = this.AttachmentList.children.length

        const ChatInputForm = document.getElementById("chat-input-form")

        if (count > 0 && this.AttachmentList.style.display !== "flex") {
            this.AttachmentList.style.display = "flex"
            ChatInputForm.style.borderTopLeftRadius = "0px"
            ChatInputForm.style.borderTopRightRadius = "0px"
            ChatInputForm.style.borderTopStyle = "solid"
        } else if (count <= 0 && this.AttachmentList.style.display === "flex") {
            this.AttachmentList.style.display = "none"
            ChatInputForm.style.borderTopLeftRadius = "12px"
            ChatInputForm.style.borderTopRightRadius = "12px"
            ChatInputForm.style.borderTopStyle = "none"
        }
    }

    disableChatInput() {
        document.getElementById("chat-input-container").style.display = "none"
    }

    enableChatInput() {
        document.getElementById("chat-input-container").style.display = "block"
    }


}