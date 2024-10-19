// dynamically resize the chat input textarea to fit the text content
function resizeChatInput() {
    ChatInput.style.height = "auto"
    ChatInput.style.height = ChatInput.scrollHeight + "px"
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
    // if (AttachmentInput.files.length > 0) {
    AttachmentPreviewContainer.style.display = "block"
    ChatInputForm.style.borderTopLeftRadius = "0px"
    ChatInputForm.style.borderTopRightRadius = "0px"
    ChatInputForm.style.borderTopStyle = "solid"

    // for (let i = 0; i < AttachmentInput.files.length; i++) {

    const selectedFile = AttachmentInput.files[0]
    const reader = new FileReader()
    reader.readAsDataURL(selectedFile) // Read the file as a data URL

    reader.onload = function (e) {
        const imgContainer = document.createElement("div")

        const imgElement = document.createElement("img")
        imgElement.src = e.target.result
        imgElement.style.display = 'block'
        imgContainer.appendChild(imgElement)
        AttachmentPreview.appendChild(imgContainer)

        console.log(selectedFile.type)
    }


    // }
    // } else if (AttachmentInput.files.length == 0) {
    //     AttachmentPreviewContainer.style.display = "none"
    //     ChatInputForm.style.borderTopLeftRadius = "12px"
    //     ChatInputForm.style.borderTopRightRadius = "12px"
    //     ChatInputForm.style.borderTopStyle = "none"
    // }
}