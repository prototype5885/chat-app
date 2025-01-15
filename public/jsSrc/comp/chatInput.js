class ChatInputClass {
    constructor() {
        this.AttachmentList = document.getElementById("attachment-list")

        this.AttachmentInput = document.getElementById("attachment-input")
        this.AttachmentInput.addEventListener("change", this.attachmentAdded.bind(this))

        this.AttachmentContainer = document.getElementById("attachment-container")

        this.ChatInput = document.getElementById("chat-input")
        this.ChatInput.addEventListener("keydown", this.chatEnterPressed.bind(this))
        this.ChatInput.addEventListener("input", this.resizeChatInput.bind(this))

        this.AttachmentButton = document.getElementById("attachment-button")
        this.AttachmentButton.addEventListener("click", this.uploadAttachment.bind(this))

        this.AttachmentsToSkip = [false, false, false, false]
    }

    // dynamically resize the chat input textarea to fit the text content
    // runs whenever the chat input textarea content changes
    // or pressed enter
    resizeChatInput() {
        this.ChatInput.style.height = "auto"
        this.ChatInput.style.height = this.ChatInput.scrollHeight + "px"
    }

    // send the text message on enter
    async chatEnterPressed(event) {
        if (event.key === "Enter" && !event.shiftKey) {
            event.preventDefault()
            let attachmentToken = null
            if (this.AttachmentInput.files.length !== 0) {
                console.log(`Chat message has [${this.AttachmentInput.files.length}] attachments, sending those first...`)

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
            }
        }
    }

    uploadAttachment() {
        this.AttachmentInput.click()
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
        for (let i = 0; i < this.AttachmentInput.files.length; i++) {
            // skip attachment that were removed from attachment list above chat input
            if (this.AttachmentsToSkip[i] === false) {
                console.log(`Preparing attachment index [${i}] called [${this.AttachmentInput.files[i].name}] for sending`)
                const hash = await MainClass.calculateSHA256(this.AttachmentInput.files[i])

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
                    formData.append("a", this.AttachmentInput.files[i])
                } else {
                    console.log(`Attachment index [${i}] exists on server, sending hash only...`)
                    const name = this.AttachmentInput.files[i].name
                    const jsonString = JSON.stringify({Hash: hash, Name: name})
                    formData.append("h", jsonString)
                }
            } else {
                console.log(`Skipping attachment index [${i}] called [${this.AttachmentInput.files[i].name}] from sending]`)
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


    attachmentAdded() {
        // reset previously added attachments
        this.resetAttachments()

        if (this.AttachmentInput.files.length >= 4) {
            console.warn("Too many attachments were added, will only use first 4")
        }

        for (let i = 0; i < this.AttachmentInput.files.length; i++) {
            // stop if there are more attachments than 4
            if (i >= 4) {
                break
            }

            const reader = new FileReader()
            reader.readAsDataURL(this.AttachmentInput.files[i])

            // when the file is loaded into the browser
            reader.onload = (e) => {
                const attachmentContainer = document.createElement("div")
                this.AttachmentList.appendChild(attachmentContainer)

                // when clicked on the attachment, it removes it
                attachmentContainer.addEventListener("click", () => {
                    attachmentContainer.remove()
                    this.AttachmentsToSkip[i] = true
                    if (this.AttachmentList.length <= 0) {
                        this.AttachmentInput.value = ""
                    }
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
    }

    calculateAttachments() {
        const count = this.AttachmentList.children.length
        console.log(`Amount of attachments changed to [${count}]`)

        const ChatInputForm = document.getElementById("chat-input-form")

        if (count > 0 && this.AttachmentContainer.style.display !== "block") {
            this.AttachmentContainer.style.display = "block"
            ChatInputForm.style.borderTopLeftRadius = "0px"
            ChatInputForm.style.borderTopRightRadius = "0px"
            ChatInputForm.style.borderTopStyle = "solid"
        } else if (count <= 0 && this.AttachmentContainer.style.display === "block") {
            this.AttachmentContainer.style.display = "none"
            ChatInputForm.style.borderTopLeftRadius = "12px"
            ChatInputForm.style.borderTopRightRadius = "12px"
            ChatInputForm.style.borderTopStyle = "none"

            this.resetAttachments()
        }
    }


    resetAttachments() {
        console.log("Resetting attachments")
        this.AttachmentList.innerHTML = ""
        this.AttachmentsToSkip = [false, false, false, false]
    }
}