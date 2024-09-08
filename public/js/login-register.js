let loginOrRegister = 'login'

const errorMessage = document.getElementById('error-message')
const button = document.getElementById('login-or-register')
const secondPassword = document.getElementById('second-password')
const invitationCode = document.getElementById('invitation-code')
const submitButton = document.getElementById('submit-button')

// set prewritten strings
const loginText = 'Login'
const registerText = 'Register'
const accountText = 'Account?'
const noAccountText = 'No account?'

// sets names in script initially
button.textContent = noAccountText
submitButton.textContent = loginText

button.addEventListener('click', () => {
    if (loginOrRegister === 'login') {
        secondPassword.style.visibility = 'visible'
        invitationCode.style.visibility = 'visible'
        button.textContent = accountText
        submitButton.textContent = registerText
        loginOrRegister = 'register'
    } else {
        secondPassword.style.visibility = 'hidden'
        invitationCode.style.visibility = 'hidden'
        button.textContent = noAccountText
        submitButton.textContent = loginText
        loginOrRegister = 'login'
    }

})

async function submit() {
    const username = document.getElementById('username-input').value
    const password = document.getElementById('first-password-input').value

    // convert password to sha512 in base64 format
    const passwordBuffer = new TextEncoder().encode(password)
    const hashedPassword = await crypto.subtle.digest('SHA-512', passwordBuffer);
    const base64Password = btoa(String.fromCharCode.apply(null, new Uint8Array(hashedPassword)));

    let dataToSend
    let url
    if (loginOrRegister === 'login') {
        dataToSend = {
            Username: username,
            Password: base64Password,
        }
        url = window.location.origin + '/login'
    } else {
        dataToSend = {
            Username: username,
            Password: base64Password,
        }
        url = window.location.origin + '/register'
    }

    const response = await fetch(url, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(dataToSend)
    })
    const result = await response.json()

    // console.log('Success:', result.Success)
    // console.log('Message:', result.Message)

    errorMessage.textContent = result.Message
    if (result.Success) {
        errorMessage.style.color = '#adff2f'
        window.location.href = "/chat.html";
    } else {
        errorMessage.textContent = result.Message
    }
}

document.getElementById('login-register-form').addEventListener('submit', function (event) {
    event.preventDefault()
    submit().then(r => { })
})




