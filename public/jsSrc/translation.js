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
            // language = navigator.language.split('-')[0]
            localStorage.setItem('language', 'en')
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
            reply: 'Reply',
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
            leaveServer: 'Leave server',
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
            mentionUser: 'Mention user',
            online: 'online',
            offline: 'offline',
            replyingTo: 'Replying to',
            retryingIn: 'Retrying in',
            connecting: 'Connecting...'
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
            leaveServer: 'Server verlassen',
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
            mentionUser: "Nutzer erwähnen",
            online: "online",
            offline: "offline"
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
            leaveServer: 'Kilépés a szerverből',
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
            mentionUser: "Felhasználó megemlítése",
            online: "elérhető",
            offline: "nem elérhető"
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
            leaveServer: 'Покинуть сервер',
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
            mentionUser: "Упоминание пользователя",
            online: "в сети",
            offline: "не в сети"
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
            leaveServer: 'Dejar el servidor',
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
            mentionUser: "Mencionar usuario",
            online: "en línea",
            offline: "desconectado"
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
            leaveServer: 'Sunucudan ayrıl',
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
            mentionUser: "Kullanıcıyı belirt",
            online: "çevrimiçi",
            offline: "çevrimdışı"
        },
        zh: {
            today: "今天",
            yesterday: "昨天",
            dm: "私信",
            addServer: "添加服务器",
            message: "消息",
            textChannels: "文本频道",
            copyChatMessage: "复制消息",
            editChatMessage: "编辑消息",
            deleteChatMessage: "删除消息",
            userSettings: "用户设置",
            profile: "个人资料",
            account: "账户",
            language: "语言",
            server: "服务器",
            channel: "频道",
            displayName: "显示名称",
            pronouns: "代词",
            statusText: "状态文本",
            apply: "应用",
            applyPicture: "应用图片",
            maximum: "最大值",
            currentPassword: "当前密码",
            newPassword: "新密码",
            confirmNewPassword: "确认新密码",
            channelSettings: "频道设置",
            deleteChannel: "删除频道",
            serverSettings: "服务器设置",
            createInviteLink: "创建邀请链接",
            leaveServer: "离开服务器",
            deleteServer: "删除服务器",
            copyUserID: "复制用户ID",
            openInBrowser: "在浏览器中打开",
            username: "用户名",
            password: "密码",
            toRegistration: "注册",
            toLogin: "返回",
            passwordAgain: "再次输入密码",
            inviteKey: "邀请码",
            login: "登录",
            register: "注册",
            mentionUser: "提及用户",
            online: "在线",
            offline: "离线"
        },
        jp: {
            today: "今日の",
            yesterday: "昨日の",
            dm: "ダイレクトメッセージ",
            addServer: "サーバーを追加",
            message: "メッセージ",
            textChannels: "テキストチャンネル",
            copyChatMessage: "メッセージをコピー",
            editChatMessage: "メッセージを編集",
            deleteChatMessage: "メッセージを削除",
            userSettings: "ユーザー設定",
            profile: "プロフィール",
            account: "アカウント",
            language: "言語",
            server: "サーバー",
            channel: "チャンネル",
            displayName: "表示名",
            pronouns: "代名詞",
            statusText: "ステータス",
            apply: "適用",
            applyPicture: "画像を適用",
            maximum: "最大",
            currentPassword: "現在のパスワード",
            newPassword: "新しいパスワード",
            confirmNewPassword: "新しいパスワードを再度入力",
            channelSettings: "チャンネル設定",
            deleteChannel: "チャンネルを削除",
            serverSettings: "サーバー設定",
            createInviteLink: "招待リンクを作成",
            leaveServer: "サーバーから退出",
            deleteServer: "サーバーを削除",
            copyUserID: "ユーザーIDをコピー",
            openInBrowser: "ブラウザで開く",
            username: "ユーザー名",
            password: "パスワード",
            toRegistration: "登録へ",
            toLogin: "戻る",
            passwordAgain: "パスワード再入力",
            inviteKey: "招待キー",
            login: "ログイン",
            register: "登録",
            mentionUser: "ユーザーに言及する",
            online: "オンライン",
            offline: "オフライン"
        }
    }
}

