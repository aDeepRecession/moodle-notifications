package config

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

type Config struct {
	Logger                     *log.Logger
	UpdatesToCheck             []string
	ToPrint                    []string
	ToPrintOnUpdates           []string
	TelegramBotKey             string
	TelegramChatID             int
	FailedRequestRepeatTimeout time.Duration
	CheckInterval              time.Duration
	LastGradesPath             string
	GradesHistoryPath          string
	MoodleCredentialsPath      string
	TelegramCredentialsPath    string
	LastTimeNotifyedPath       string
}

type telegramCredentialsJSON struct {
	TelegramBotKey string `json:"telegramBotKey"`
	TelegramChatID int    `json:"telegramChatID"`
}

type configJSON struct {
	UpdatesToCheck             []string `json:"updatesToCheck"`
	ToPrint                    []string `json:"toPrint"`
	ToPrintOnUpdates           []string `json:"toPrintOnUpdates"`
	FailedRequestRepeatTimeout int      `json:"failedRequestRepeatTimeout"`
	CheckInterval              int      `json:"checkInterval"`
	LastGradesPath             string   `json:"lastGradesPath"`
	GradesHistoryPath          string   `json:"gradesHistoryPath"`
	MoodleCredentialsPath      string   `json:"moodleCredentialsPath"`
	TelegramCredentialsPath    string   `json:"telegramCredentialsPath"`
	LastTimeNotifyedPath       string   `json:"lastTimeNotifyedPath"`
}

var logger *log.Logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)

func NewConfig(cfgReader io.Reader) (Config, error) {
	cfgJSON, err := getConfigJSON(cfgReader)
	if err != nil {
		return Config{}, err
	}

	telegramCredentials, err := getTelegramCredentials(cfgJSON.TelegramCredentialsPath)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		Logger:                     logger,
		UpdatesToCheck:             cfgJSON.UpdatesToCheck,
		ToPrint:                    cfgJSON.ToPrint,
		ToPrintOnUpdates:           cfgJSON.ToPrintOnUpdates,
		TelegramBotKey:             telegramCredentials.TelegramBotKey,
		TelegramChatID:             telegramCredentials.TelegramChatID,
		FailedRequestRepeatTimeout: time.Duration(cfgJSON.FailedRequestRepeatTimeout) * time.Second,
		CheckInterval:              time.Duration(cfgJSON.CheckInterval) * time.Second,
		LastGradesPath:             cfgJSON.LastGradesPath,
		GradesHistoryPath:          cfgJSON.GradesHistoryPath,
		MoodleCredentialsPath:      cfgJSON.MoodleCredentialsPath,
		TelegramCredentialsPath:    cfgJSON.MoodleCredentialsPath,
		LastTimeNotifyedPath:       cfgJSON.LastTimeNotifyedPath,
	}

	return cfg, nil
}

func getConfigJSON(cfgReader io.Reader) (configJSON, error) {
	cfgJSON := configJSON{}

	cfgByte, err := io.ReadAll(cfgReader)
	if err != nil {
		return configJSON{}, fmt.Errorf("failed to get config: %v", err)
	}

	err = json.Unmarshal(cfgByte, &cfgJSON)
	if err != nil {
		return configJSON{}, fmt.Errorf("failed to get config: %v", err)
	}
	return cfgJSON, nil
}

func getTelegramCredentials(credentialsPath string) (telegramCredentialsJSON, error) {
	credentialsFile, err := os.OpenFile(credentialsPath, os.O_RDONLY, 0644)
	if err != nil {
		return telegramCredentialsJSON{}, fmt.Errorf("failed to get config: %v", err)
	}
	defer credentialsFile.Close()

	credentialsByte, err := io.ReadAll(credentialsFile)
	if err != nil {
		return telegramCredentialsJSON{}, fmt.Errorf("failed to get config: %v", err)
	}

	credentials := telegramCredentialsJSON{}
	err = json.Unmarshal(credentialsByte, &credentials)
	if err != nil {
		return telegramCredentialsJSON{}, fmt.Errorf("failed to get config: %v", err)
	}

	return credentials, nil
}
