package moodle

import (
	"fmt"
	"log"
)

func GetTokens(credentialsPath string, logger *log.Logger) (MoodleToken, error) {
	cookieRequestManager, err := newCookieRequest(credentialsPath, logger)
	if err != nil {
		return "", err
	}

	loginCredentials, err := cookieRequestManager.requestNewTokens()
	if err != nil {
		return "", fmt.Errorf("failed to get old cookies: %v", err)
	}
	oldToken := MoodleToken(loginCredentials)

	if check(oldToken, logger) {
		return oldToken, nil
	}

	tokens, err := cookieRequestManager.requestNewTokens()
	if err != nil {
		return "", err
	}

	return tokens, nil
}

func check(token MoodleToken, logger *log.Logger) bool {
	api, err := NewMoodle(string(token), logger)
	if err != nil {
		return false
	}
	return api.isTokenGood()
}
