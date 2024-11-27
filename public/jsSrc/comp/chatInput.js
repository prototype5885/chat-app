// dynamically resize the chat input textarea to fit the text content
// runs whenever the chat input textarea content changes
function resizeChatInput() {
    ChatInput.style.height = "auto"
    ChatInput.style.height = ChatInput.scrollHeight + "px"
}

// send the text message on enter
function sendChatEnter(event) {
    if (event.key === "Enter" && !event.shiftKey) {
        event.preventDefault()
        readChatInput()
    }
}

// read the text message for sending
function readChatInput() {
    if (ChatInput.value) {
        sendChatMessage(ChatInput.value, currentChannelID)
        ChatInput.value = ""
        resizeChatInput()
    }
}

function uploadAttachment() {
    AttachmentInput.click()


    // const response = await fetch(url, {
    //     method: 'POST',
    //     headers: {
    //         'Content-Type': 'application/json'
    //     },
    //     body: JSON.stringify(dataToSend)
    // })
    // const result = await response.json()
}

function attachmentAdded() {
    for (i = 0; i < AttachmentInput.files.length; i++) {
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

    // }
    // } else if (AttachmentInput.files.length == 0) {
    //     AttachmentPreviewContainer.style.display = "none"
    //     ChatInputForm.style.borderTopLeftRadius = "12px"
    //     ChatInputForm.style.borderTopRightRadius = "12px"
    //     ChatInputForm.style.borderTopStyle = "none"
    // }
}

function calculateAttachments() {
    const count = AttachmentList.children.length
    console.log("attachments:", count)

    if (count > 0 && AttachmentContainer.style.display != "block") {
        AttachmentContainer.style.display = "block"
        ChatInputForm.style.borderTopLeftRadius = "0px"
        ChatInputForm.style.borderTopRightRadius = "0px"
        ChatInputForm.style.borderTopStyle = "solid"
    } else if (count == 0 && AttachmentContainer.style.display == "block") {
        AttachmentContainer.style.display = "none"
        ChatInputForm.style.borderTopLeftRadius = "12px"
        ChatInputForm.style.borderTopRightRadius = "12px"
        ChatInputForm.style.borderTopStyle = "none"
    }
}