
const wsClient = new WebSocket('wss://' + window.location.host);
const messages = document.getElementById('chat-message-list')

// const friendList = document.getElementById('first-column-main-container')

function getCurrentTime() {
    const currentDate = new Date()

    // const options = {
    //     year: 'numeric',
    //     month: 'numeric',
    //     day: 'numeric',
    //     hour: '2-digit',
    //     minute: '2-digit',
    //     second: undefined,
    //     hour12: undefined 
    // }

    let localDateTime = currentDate.toLocaleString()

    return localDateTime
}

wsClient.onopen = function (event) {
    console.log('Connected to WebSocket successfully.');
};

// when server sends a message
wsClient.onmessage = function (event) {
    const packetString = event.data.toString()
    const packetParsed = JSON.parse(packetString)

    const type = packetParsed.type
    const data = JSON.parse(packetParsed.data)

    console.log('received: ' + packetString);

    switch (type) {
        case 'serverChatMsg':
            addChatMessage(data)
            messages.scrollTo({
                top: messages.scrollHeight,
                behavior: 'smooth'
            })
            break
    }
}

function sendPacket(type, data) {
    if (wsClient.readyState === WebSocket.OPEN) {
        const packet = {
            type: type,
            data: JSON.stringify(data)
        }
        const strPacket = JSON.stringify(packet)
        console.log('sending: ' + strPacket)
        wsClient.send(strPacket);
    }
}

// socket.on('user connected', () => {
//     const messageData = {
//         userID: 0,
//         msgID: 0,
//         msg: 'New user has joined'
//     }

//     addChatMessage(messageData)
// })

// function main() {
//     // for (i = 0; i < 20; i++) {
//     // socket.emit('chat message', 'test')
//     // addMember(5)
//     // addFriend(5)
//     // }
// }

// main()