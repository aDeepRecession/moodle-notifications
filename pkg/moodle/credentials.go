package moodle

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Credentials struct {
	CredentialsPath string
}

type CredentialsData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

func newCredentials(credentialsPath string) Credentials {
	return Credentials{credentialsPath}
}

func (cm Credentials) get() (CredentialsData, error) {
	credentialsFile, err := os.OpenFile(cm.CredentialsPath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return CredentialsData{}, fmt.Errorf("failed to get credentials")
	}

	credentialsJSON, err := io.ReadAll(credentialsFile)
	if err != nil {
		return CredentialsData{}, fmt.Errorf("failed to get credentials")
	}

	var credentials CredentialsData
	err = json.Unmarshal(credentialsJSON, &credentials)
	if err != nil {
		return CredentialsData{}, fmt.Errorf("failed to get credentials")
	}

	if credentials.Login == "" && credentials.Password == "" && credentials.Token == "" {
		return CredentialsData{}, fmt.Errorf("innopolis credentials are empty")
	}

	return credentials, nil
}

func (cm Credentials) save(newToken string) error {
	newCredentials, err := cm.get()
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
