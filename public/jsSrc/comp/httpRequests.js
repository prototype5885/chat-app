async function sendPostRequest(url, struct) {
    const response = await fetch(window.location.origin + url, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(struct)
    })
    const result = await response.json()
    return result

    // console.log('Success:', result.Success)
    // console.log('Message:', result.Message)
}

// async function requestChannelList() {
//     const response = await fetch(`/channels/${currentChannelID}`);
//     const data = await response.text();
// }