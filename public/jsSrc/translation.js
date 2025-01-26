class Translation {
    static xmlData = `
    <translations>
        <en>
            <today>Today at</today>
            <yesterday>Yesterday at</yesterday>
            <dm>Direct Messages</dm>
            <add-server>Add Server</add-server>
            <message>Message</message>
            <text-channels>text channels</text-channels>
            <copy-chat-message>Copy Message</copy-chat-message>
            <edit-chat-message>Edit Message</edit-chat-message>
            <delete-chat-message>Delete Message</delete-chat-message>
            <user-settings>User Settings</user-settings>
        </en>
        <de>
            <today>heute um</today>
            <yesterday>gestern um</yesterday>
            <dm>Direktnachrichten</dm>
            <add-server>Server hinzufügen</add-server>
            <message>Nachricht an</message>
            <text-channels>textkanäle</text-channels>
            <copy-chat-message>Text kopieren</copy-chat-message>
            <edit-chat-message>Nachricht bearbeiten</edit-chat-message>
            <delete-chat-message>Nachricht löschen</delete-chat-message>
            <user-settings>Benutzereinstellungen</user-settings>
        </de>
        <hu>
            <today>Ma </today>
        </hu>
        <ru>
            <today>Сегодня, в </today>
        </ru>
        <es>
            <today>hoy a las </today>
            <yesterday>ayer a las</yesterday>
            <dm>Mensajes directos</dm>
            <add-server>Añadir un servidor</add-server>
            <message>Enviar mensaje a</message>
            <text-channels>canales de texto</text-channels>
            <copy-chat-message>Copiar texto</copy-chat-message>
            <edit-chat-message>Editar mensaje</edit-chat-message>
            <delete-chat-message>Denunciar mensaje</delete-chat-message>
            <user-settings>Ajustes de usuario</user-settings>
        </es>
    </translations>`

    static #lang = ''

    static #parser = new DOMParser();
    static #xmlDoc = this.#parser.parseFromString(this.xmlData, "text/xml");

    static get(key) {
        if (this.#lang === '') {
            this.#lang = 'en'
        }
        const element = this.#xmlDoc.querySelector(`${this.#lang} ${key}`)
        if (element !== null) {
            return element.textContent
        } else {
            return this.#xmlDoc.querySelector(`en ${key}`).textContent
        }
    }


    static setLanguage(language) {
        this.#lang = language.split('-')[0];
        console.log('Language set to: ' + language)
    }
}

