const loginUsername = document.getElementById("login-username")
const loginPassword = document.getElementById("login-password")

const username = document.getElementById("register-username")
const firstPassword = document.getElementById("register-password-first")
const secondPassword = document.getElementById("register-password-second")
const inviteKey = document.getElementById("invite-key")

const errorMessage = document.getElementById("error-message")

Translation.setLanguage()
document.querySelector(`label[for="${loginUsername.id}"]`).textContent = Translation.get('username') + ":"


function toLogin() {
    errorMessage.textContent = ""
    document.getElementById("register-container").style.display = "none"
    document.getElementById("login-container").style.display = "block"
}

function toRegister() {
    errorMessage.textContent = ""
    document.getElementById("register-container").style.display = "block"
    document.getElementById("login-container").style.display = "none"
}

async function sendLogin(event) {
    event.preventDefault()

    const json = {
        Username: loginUsername.value,
        Password: await hashPassword(loginPassword.value)
    }

    const response = await fetch("/login", {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(json)
    })
    if (response.ok) {
        if (response.redirected) {
            window.location.href = response.url
        } else {
            errorMessage.textContent = await response.text()
        }
    }
}

async function sendRegister(event) {
    event.preventDefault()
    
    if (firstPassword.value !== secondPassword.value) {
        errorMessage.textContent = "Passwords don't match"
        return
    }

    const json = {
        Username: username.value,
        Password: await hashPassword(firstPassword.value),
        InviteKey: inviteKey.value
    }

    const response = await fetch("/register", {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(json)
    })
    if (response.ok) {
        console.log(response)
        if (response.redirected) {
            console.log("Redirecting to: ", response.url)
            window.location.href = response.url
        } else {
            errorMessage.textContent = await response.text()
        }
    }
}

async function hashPassword(password) {
    const passwordBuffer = new TextEncoder().encode(password)
    const hashedPassword = await crypto.subtle.digest('SHA-512', passwordBuffer)
    return btoa(String.fromCharCode.apply(null, new Uint8Array(hashedPassword)))
}

function asciiOnly(event) {
    event.target.value = event.target.value.replace(/[^\x00-\x7F]/g, '')
}