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
    AttachmentPreviewContainer.style.display = "block"
    ChatInputForm.style.borderTopLeftRadius = "0px"
    ChatInputForm.style.borderTopRightRadius = "0px"
    ChatInputForm.style.borderTopStyle = "solid"

    for (i = 0; i < AttachmentInput.files.length; i++) {
        const reader = new FileReader()
        reader.readAsDataURL(AttachmentInput.files[i]) // Read the file as a data URL

        reader.onload = function (e) {
            const imgContainer = document.createElement("div")

            const imgElement = document.createElement("img")
            imgElement.src = e.target.result
            imgElement.style.display = 'block'
            imgContainer.appendChild(imgElement)
            AttachmentPreview.appendChild(imgContainer)
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