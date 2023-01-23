package moodletokensmanager


type MoodleCookies struct {
    Secret1 string
    Token string
    Secret2 string
}

func GetTokens() (MoodleCookies, error) {

    // get old cookies
    // check if they are good

    // get new cookies if they are bod
    // save new cokies


    cookieRequestManager, err := newCookieRequestManager()
    if err != nil {
        return MoodleCookies{}, err
    }

    tokens, err := cookieRequestManager.requestNewTokens()
    if err != nil {
        return MoodleCookies{}, err
    }
    return tokens, nil
}

