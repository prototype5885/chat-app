
// dynamically resize the chat input textarea to fit the text content
// runs whenever the chat input textarea content changes
function resizeChatInput() {
    ChatInput.style.height = "auto"
    ChatInput.style.height = ChatInput.scrollHeight + "px"
}

// send the text message on enter
async function chatEnterPressed(event) {
    if (event.key === "Enter" && !event.shiftKey) {
        event.preventDefault()
        await readChatInput()
    }
}

// read the text message for sending
async function readChatInput() {
    let attachmentToken = null
    if (AttachmentInput.files.length !== 0) {
        await checkAttachments()
        attachmentToken = await sendAttachment()
        console.log("http response to uploading attachment:", attachmentToken)
    }

    if (ChatInput.value || attachmentToken !== null) {
        if (attachmentToken !== null) {
            await sendChatMessage(ChatInput.value, currentChannelID, attachmentToken.AttToken)
        } else {
            await sendChatMessage(ChatInput.value, currentChannelID, null)
        }
        ChatInput.value = ""
        AttachmentInput.value = ""
        resizeChatInput()
    }
}

function uploadAttachment() {
    AttachmentInput.click()
}

async function checkAttachments() {
    const hash = await calculateSHA256(AttachmentInput.files[0])

    let hashes = []
    for (let i = 0; i < AttachmentInput.files.length; i++) {
        const hash = await calculateSHA256(AttachmentInput.files[0])
        hashes.push(hash)
    }

    const checkRequest = new XMLHttpRequest()

    checkRequest.onload = function () {
        if (checkRequest.status === 200) {
            console.log("Response to check attachment request: ", checkRequest.responseText)
        }
    }

    checkRequest.open("POST", "/check-attachment")
    checkRequest.send(hashes)
}

async function sendAttachment() {
    const formData = new FormData()

    for (let i = 0; i < AttachmentInput.files.length; i++) {
        formData.append("attachment[]", AttachmentInput.files[i])
    }

    const response = await fetch('/upload-attachment', {
        method: "POST",
        body: formData
    })

    const attachmentToken = await response.json()

    if (!response.ok) {
        console.error("Attachment upload failed")
        return
    }

    console.log("Attachment was uploaded successfully")
    AttachmentList.innerHTML = ""
    calculateAttachments()
    return attachmentToken
}

function attachmentAdded() {
    for (i = 0; i < AttachmentInput.files.length; i++) {
        if (i >= 4) {
            console.warn("Too many attachments were added")
            continue
        }
        const reader = new FileReader()
        reader.readAsDataURL(AttachmentInput.files[i])

        reader.onload = function (e) {
            const attachmentContainer = document.createElement("div")
            AttachmentList.appendChild(attachmentContainer)

            attachmentContainer.addEventListener("click", function () {
                attachmentContainer.remove()
                calculateAttachments()
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
            calculateAttachments()
        }
    }
}

function calculateAttachments() {
    const count = AttachmentList.children.length
    console.log("Attachments count:", count)

    if (count > 0 && AttachmentContainer.style.display !== "block") {
        AttachmentContainer.style.display = "block"
        ChatInputForm.style.borderTopLeftRadius = "0px"
        ChatInputForm.style.borderTopRightRadius = "0px"
        ChatInputForm.style.borderTopStyle = "solid"
    } else if (count === 0 && AttachmentContainer.style.display === "block") {
        AttachmentContainer.style.display = "none"
        ChatInputForm.style.borderTopLeftRadius = "12px"
        ChatInputForm.style.borderTopRightRadius = "12px"
        ChatInputForm.style.borderTopStyle = "none"
    }
}