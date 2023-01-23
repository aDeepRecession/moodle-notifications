package ssocredentials

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type SsoCredentialsManager struct{}

type Credentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func NewCredentialsManager() SsoCredentialsManager {
	return SsoCredentialsManager{}
}

func (cm SsoCredentialsManager) GetLoginCredentials() (Credentials, error) {
	credentialsFile, err := os.OpenFile("./credentials.json", os.O_RDONLY, 644)
	if err != nil {
		return Credentials{}, fmt.Errorf("failed to get credentials")
	}

	credentialsJSON, err := io.ReadAll(credentialsFile)
	if err != nil {
		return Credentials{}, fmt.Errorf("failed to get credentials")
	}

	var credentials Credentials
	json.Unmarshal(credentialsJSON, &credentials)

	return credentials, nil
}
