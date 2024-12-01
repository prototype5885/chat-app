let listOfAttachments = []

// dynamically resize the chat input textarea to fit the text content
// runs whenever the chat input textarea content changes
function resizeChatInput() {
    ChatInput.style.height = "auto"
    ChatInput.style.height = ChatInput.scrollHeight + "px"
}

// send the text message on enter
async function sendChatEnter(event) {
    if (event.key === "Enter" && !event.shiftKey) {
        event.preventDefault()
        listOfAttachments = []
        if (AttachmentInput.files.length !== 0) {
            for (let i = 0; i < AttachmentInput.files.length; i++) {
                listOfAttachments.push(AttachmentInput.files[i].name)
            }
            await sendAttachment()
        }
        readChatInput()
        AttachmentInput.value = ""
    }
}

// read the text message for sending
function readChatInput() {
    if (ChatInput.value || listOfAttachments.length !== 0) {
        console.log("list:", listOfAttachments)
        sendChatMessage(ChatInput.value, currentChannelID, listOfAttachments)
        ChatInput.value = ""
        resizeChatInput()
    }
}

function uploadAttachment() {
    AttachmentInput.click()
}

async function sendAttachment() {
    const formData = new FormData()

    formData.append("attachment", AttachmentInput.files[0])

    const response = await fetch('/upload-attachment', {
        method: "POST",
        body: formData
    })

    if (!response.ok) {
        console.error("Attachment upload failed")
        return
    }

    console.log("Attachment was uploaded successfully")
    AttachmentList.innerHTML = ""
    calculateAttachments()
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