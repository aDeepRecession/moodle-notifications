package ssocredentials

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type SsoCredentialsManager struct {
	CredentialsPath string
}

type Credentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

func NewCredentialsManager(credentialsPath string) SsoCredentialsManager {
	return SsoCredentialsManager{credentialsPath}
}

func (cm SsoCredentialsManager) GetLoginCredentials() (Credentials, error) {
	credentialsFile, err := os.OpenFile(cm.CredentialsPath, os.O_RDONLY, 0644)
	if err != nil {
		return Credentials{}, fmt.Errorf("failed to get credentials")
	}

	credentialsJSON, err := io.ReadAll(credentialsFile)
	if err != nil {
		return Credentials{}, fmt.Errorf("failed to get credentials")
	}

	var credentials Credentials
	err = json.Unmarshal(credentialsJSON, &credentials)
	if err != nil {
		return Credentials{}, fmt.Errorf("failed to get credentials")
	}

	return credentials, nil
}

func (cm SsoCredentialsManager) SaveToken(newToken string) error {
	newCredentials, err := cm.GetLoginCredentials()
	if err != nil {
		return fmt.Errorf("failed to save credentials")
	}

	newCredentials.Token = newToken

	credentialsFile, err := os.OpenFile(
		cm.CredentialsPath,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0644,
	)
	if err != nil {
		return fmt.Errorf("failed to save credentials")
	}

	credentialsJSON, err := json.Marshal(newCredentials)
	if err != nil {
		return fmt.Errorf("failed to save credentials")
	}

	_, err = credentialsFile.Write(credentialsJSON)
	if err != nil {
		return fmt.Errorf("failed to save credentials")
	}

	return nil
}
