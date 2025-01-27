class Translation {
    static lang = ''

    static get(key) {
        if (this.lang === '') {
            this.lang = 'en'
        }
        const txt = this.#translations[this.lang]?.[key]
        if (txt !== undefined) {
            return txt
        } else {
            return this.#translations['en']?.[key]
        }
    }

    static setLanguage() {
        this.#checkMissingTranslations()

        let language = localStorage.getItem('language')
        if (language === null) {
            language = navigator.language.split('-')[0]
            localStorage.setItem('language', language)
        }
        this.lang = language
        console.log('Language set to: ' + language)
    }

    static #checkMissingTranslations() {
        const referenceKeys = Object.keys(this.#translations.en)

        let issue = false

        for (const language in this.#translations) {
            const lang = this.#translations[language]
            const currentKeys = Object.keys(lang)

            if (!this.#compareArrays(referenceKeys, currentKeys)) {
                const missingKeys = referenceKeys.filter(key => !currentKeys.includes(key))
                const extraKeys = currentKeys.filter(key => !referenceKeys.includes(key))

                let message = `Language [${language}] has different keys.`
                if (missingKeys.length > 0) {
                    message += ` Missing: ${missingKeys.join(', ')}.`
                    issue = true
                }
                if (extraKeys.length > 0) {
                    message += ` Extra: ${extraKeys.join(', ')}.`
                    issue = true
                }
                console.warn(message)
            }
        }
        if (!issue) {
            console.log('There were no issues reading translation data')
        }
    }

    static #compareArrays(arr1, arr2) {
        if (arr1.length !== arr2.length) {
            return false
        }
        const sortedArr1 = [...arr1].sort()
        const sortedArr2 = [...arr2].sort()
        for (let i = 0; i < sortedArr1.length; i++) {
            if (sortedArr1[i] !== sortedArr2[i]) {
                return false
            }
        }
        return true
    }

    static #translations = {
        en: {
            today: 'Today at',
            yesterday: 'Yesterday at',
            dm: 'Direct messages',
            addServer: 'Add server',
            message: 'Message',
            textChannels: 'text channels',
            copyChatMessage: 'Copy message',
            editChatMessage: 'Edit message',
            deleteChatMessage: 'Delete message',
            userSettings: 'User settings',
            profile: 'Profile',
            account: 'Account',
            language: 'Language',
            server: 'Server',
            channel: 'Channel',
            displayName: 'Display name',
            pronouns: 'Pronouns',
            statusText: 'Status text',
            apply: 'Apply',
            applyPicture: 'Apply picture',
            maximum: 'Maximum',
            currentPassword: 'Current password',
            newPassword: 'New password',
            confirmNewPassword: 'Confirm new password',
            channelSettings: 'Channel settings',
            deleteChannel: 'Delete channel',
            serverSettings: 'Server settings',
            createInviteLink: 'Create invite link',
            deleteServer: 'Delete server',
            copyUserID: 'Copy user ID',
            openInBrowser: 'Open in browser',
            username: 'Username',
            password: 'Password',
            toRegistration: 'Registration',
            toLogin: 'Back',
            passwordAgain: 'Password again',
            inviteKey: 'Invite key',
            login: 'Login',
            register: 'Register',
        },
        de: {
            today: 'heute um',
            yesterday: 'gestern um',
            dm: 'Direktnachrichten',
            addServer: 'Server hinzufügen',
            message: 'Nachricht an',
            textChannels: 'textkanäle',
            copyChatMessage: 'Text kopieren',
            editChatMessage: 'Nachricht bearbeiten',
            deleteChatMessage: 'Nachricht löschen',
            userSettings: 'Benutzereinstellungen',
            profile: 'Profil',
            account: 'Konto',
            language: 'Sprache',
            server: 'Server',
            channel: 'Kanal',
            displayName: 'Anzeigename',
            pronouns: 'Pronomen',
            statusText: 'Status Text',
            apply: 'Anwenden',
            applyPicture: 'Bild anwenden',
            maximum: 'Maximal',
            currentPassword: 'Aktuelles passwort',
            newPassword: 'Neues Passwort',
            confirmNewPassword: 'Neues passwort bestätigen',
            channelSettings: 'Kanaleinstellungen',
            deleteChannel: 'Kanal löschen',
            serverSettings: 'Servereinstellungen',
            createInviteLink: 'Einladungslink erstellen',
            deleteServer: 'Server löschen',
            copyUserID: 'Benutzer-ID kopieren',
            openInBrowser: 'Im Browser öffnen',
            username: 'Benutzername',
            password: 'Passwort',
            toRegistration: 'Registrierung',
            toLogin: 'Zurück',
            passwordAgain: 'Passwort wiederholen',
            inviteKey: 'Einladungsschlüssel',
            login: 'Anmelden',
            register: 'Registrieren',
        },
        hu: {
            today: 'Ma',
            yesterday: 'Tegnap',
            dm: 'Üzenetek',
            addServer: 'Szerver hozzáadása',
            message: 'Üzenet',
            textChannels: 'Szöveges csatornák',
            copyChatMessage: 'Másolás',
            editChatMessage: 'Szerkesztés',
            deleteChatMessage: 'Törlés',
            userSettings: 'Felhasználói beállítások',
            profile: 'Profil',
            account: 'Fiók',
            language: 'Nyelv',
            server: 'Szerver',
            channel: 'Csatorna',
            displayName: 'Név',
            pronouns: 'Névmások',
            statusText: 'Állapot szövege',
            apply: 'Alkalmaz',
            applyPicture: 'Kép feltöltése',
            maximum: 'Maximum',
            currentPassword: 'Jelenlegi jelszó',
            newPassword: 'Új jelszó',
            confirmNewPassword: 'Új jelszó megerősítése',
            channelSettings: 'Csatorna beállításai',
            deleteChannel: 'Csatorna törlése',
            serverSettings: 'Szerver beállításai',
            createInviteLink: 'Meghívó link létrehozása',
            deleteServer: 'Szerver törlése',
            copyUserID: 'Felhasználói azonosító másolása',
            openInBrowser: 'Megnyitás böngészőben',
            username: 'Felhasználónév',
            password: 'Jelszó',
            toRegistration: 'Regisztráció',
            toLogin: 'Vissza',
            passwordAgain: 'Jelszó megismétlése',
            inviteKey: 'Meghívó jelszó',
            login: 'Bejelentkezés',
            register: 'Regisztráció',
        },
        ru: {
            today: 'Сегодня, в',
            yesterday: 'Вчера, в',
            dm: 'Личные сообщения',
            addServer: 'Добавить сервер',
            message: 'Сообщение',
            textChannels: 'Текстовые каналы',
            copyChatMessage: 'Копировать сообщение',
            editChatMessage: 'Редактировать сообщение',
            deleteChatMessage: 'Удалить сообщение',
            userSettings: 'Настройки пользователя',
            profile: 'Профиль',
            account: 'Аккаунт',
            language: 'Язык',
            server: 'Сервер',
            channel: 'Канал',
            displayName: 'Отображаемое имя',
            pronouns: 'Местоимения',
            statusText: 'Текст статуса',
            apply: 'Применить',
            applyPicture: 'Применить изображение',
            maximum: 'Максимум',
            currentPassword: 'Текущий пароль',
            newPassword: 'Новый пароль',
            confirmNewPassword: 'Подтвердить новый пароль',
            channelSettings: 'Настройки канала',
            deleteChannel: 'Удалить канал',
            serverSettings: 'Настройки сервера',
            createInviteLink: 'Создать ссылку-приглашение',
            deleteServer: 'Удалить сервер',
            copyUserID: 'Копировать ID пользователя',
            openInBrowser: 'Открыть в браузере',
            username: 'Имя пользователя',
            password: 'Пароль',
            toRegistration: 'Регистрация',
            toLogin: 'Назад',
            passwordAgain: 'Пароль еще раз',
            inviteKey: 'Ключ приглашения',
            login: 'Войти',
            register: 'Зарегистрироваться',
        },
        es: {
            today: 'hoy a las',
            yesterday: 'ayer a las',
            dm: 'Mensajes directos',
            addServer: 'Añadir un servidor',
            message: 'Enviar mensaje a',
            textChannels: 'canales de texto',
            copyChatMessage: 'Copiar texto',
            editChatMessage: 'Editar mensaje',
            deleteChatMessage: 'Denunciar mensaje',
            userSettings: 'Ajustes de usuario',
            profile: 'Perfil',
            account: 'Cuenta',
            language: 'Idioma',
            server: 'Servidor',
            channel: 'Canal',
            displayName: 'Mostrar nombre',
            pronouns: 'Pronombres',
            statusText: 'Texto de estado',
            apply: 'Aplicar',
            applyPicture: 'Aplicar imagen',
            maximum: 'Máximo',
            currentPassword: 'Contraseña actual',
            newPassword: 'Contraseña nueva',
            confirmNewPassword: 'Confirmación de contraseña nueva',
            channelSettings: 'Ajustes del canal',
            deleteChannel: 'Eliminar canal',
            serverSettings: 'Ajustes del servidor',
            createInviteLink: 'Crear enlace de invitación',
            deleteServer: 'Eliminar servidor',
            copyUserID: 'Copiar ID de usuario',
            openInBrowser: 'Abrir en el navegador',
            username: 'Nombre de usuario',
            password: 'Contraseña',
            toRegistration: 'Registro',
            toLogin: 'Atrás',
            passwordAgain: 'Repetir contraseña',
            inviteKey: 'Clave de invitación',
            login: 'Acceso',
            register: 'Registrarse',
        },
        tr: {
            today: 'Bugün',
            yesterday: 'Dün',
            dm: 'Özel mesajlar',
            addServer: 'Sunucu ekle',
            message: 'Mesaj',
            textChannels: 'Metin kanalları',
            copyChatMessage: 'Mesajı kopyala',
            editChatMessage: 'Mesajı düzenle',
            deleteChatMessage: 'Mesajı sil',
            userSettings: 'Kullanıcı ayarları',
            profile: 'Profil',
            account: 'Hesap',
            language: 'Dil',
            server: 'Sunucu',
            channel: 'Kanal',
            displayName: 'Görünen ad',
            pronouns: 'Zamirler',
            statusText: 'Durum metni',
            apply: 'Uygula',
            applyPicture: 'Resmi uygula',
            maximum: 'Maksimum',
            currentPassword: 'Mevcut şifre',
            newPassword: 'Yeni şifre',
            confirmNewPassword: 'Yeni şifreyi onayla',
            channelSettings: 'Kanal ayarları',
            deleteChannel: 'Kanalı sil',
            serverSettings: 'Sunucu ayarları',
            createInviteLink: 'Davet bağlantısı oluştur',
            deleteServer: 'Sunucuyu sil',
            copyUserID: 'Kullanıcı ID\'sini kopyala',
            openInBrowser: 'Tarayıcıda aç',
            username: 'Kullanıcı adı',
            password: 'Şifre',
            toRegistration: 'Kaydol',
            toLogin: 'Geri',
            passwordAgain: 'Şifreyi tekrar gir',
            inviteKey: 'Davet anahtarı',
            login: 'Giriş yap',
            register: 'Kaydol',
        }
    }
}

