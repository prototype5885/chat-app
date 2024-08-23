let loginOrRegister = 'login'

const loginText = 'Login'
const registerText = 'Register'
const accountText = 'Account?'
const noAccountText = 'No account?'

document.getElementById('login-register-form').addEventListener('submit', function (event) {
    event.preventDefault();

    const username = document.getElementById('username-input').value;
    const password = document.getElementById('first-password-input').value;
    // const password = document.getElementById('first-password-input').value;

    let dataToSend
    if (loginOrRegister == 'login') {
        dataToSend = {
            type: 'login',
            username: username,
            password: password,
        }
    } else {
        dataToSend = {
            type: 'register',
            username: username,
            password: password,
        }
    }
    var xhr = new XMLHttpRequest();
    xhr.open('POST', window.location.origin + '/logging_in', true);
    xhr.setRequestHeader('Content-Type', 'application/json');


    xhr.onreadystatechange = function () {
        if (xhr.readyState === 4) {
            if (xhr.status === 200) {
                console.log(JSON.parse(xhr.responseText));
            } else {
                console.error('Error:', xhr.statusText);
            }
        }
    };

    xhr.send(JSON.stringify(dataToSend));

    // fetch('https://127.0.0.1:3000/logging_in', {
    //     method: 'POST',
    //     headers: {
    //         'Content-Type': 'application/json',
    //     },
    //     body: JSON.stringify(dataToSend),
    // })
    //     .then(response => response.json())
    //     .then(json => console.log(json))
    //     .catch(error => console.error('Error:', error));
});

document.addEventListener('DOMContentLoaded', (event) => {
    const button = document.getElementById('login-or-register')
    const secondPassword = document.getElementById('second-password')
    const invitationCode = document.getElementById('invitation-code')
    const submitButton = document.getElementById('submit-button')

    // sets names in script initially
    button.textContent = noAccountText
    submitButton.textContent = loginText

    // hide registration stuff
    secondPassword.style.visibility = 'hidden'
    invitationCode.style.visibility = 'hidden'

    button.addEventListener('click', () => {
        if (loginOrRegister == 'login') {
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

    });
});    