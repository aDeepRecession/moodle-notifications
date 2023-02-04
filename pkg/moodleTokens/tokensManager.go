package moodletokensmanager

import (
	"fmt"
	"log"

	moodleapi "github.com/aDeepRecession/moodle-scrapper/pkg/moodleAPI"
)

type MoodleToken string

func GetTokens(credentialsPath string, logger *log.Logger) (MoodleToken, error) {
	cookieRequestManager, err := newCookieRequestManager(credentialsPath, logger)
	if err != nil {
		return "", err
	}

	loginCredentials, err := cookieRequestManager.credentials.GetLoginCredentials()
	if err != nil {
		return "", fmt.Errorf("failed to get old cookies: %v", err)
	}
	oldToken := MoodleToken(loginCredentials.Token)

	if isTokenGood(oldToken, logger) {
		return oldToken, nil
	}

	tokens, err := cookieRequestManager.requestNewTokens()
	if err != nil {
		return "", err
	}

	return tokens, nil
}

func isTokenGood(token MoodleToken, logger *log.Logger) bool {
	api, err := moodleapi.NewMoodleAPI(string(token), logger)
	if err != nil {
		return false
	}
	return api.IsTokenGood()
}
