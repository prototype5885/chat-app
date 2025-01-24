class Translation {
    static translationJson = {
        'en': {
            'today': 'Today at '
        },
        'de': {
            'today': 'Am heute '
        },
        'hu': {
            'today': 'Ma '
        },
        'ru': {
            'today': 'Сегодня, в '
        },
        'es': {
            'today': 'hoy a las '
        }
    }

    static translation = {}

    static setLanguage(language) {
        language = language.split('-')[0];
        console.log('Language set to: ' + language)
        switch (language) {
            case 'de':
                Translation.translation = Translation.translationJson.de
                break
            case 'hu':
                Translation.translation = Translation.translationJson.hu
                break
            case 'ru':
                Translation.translation = Translation.translationJson.ru
                break
            case 'es':
                Translation.translation = Translation.translationJson.es
                break
            case 'en':
            default:
                Translation.translation = Translation.translationJson.en
                break
        }
    }
}

