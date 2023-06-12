package notifyer

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aDeepRecession/moodle-scrapper/pkg/config"
	"github.com/aDeepRecession/moodle-scrapper/pkg/notifyer/formatter"
	"github.com/aDeepRecession/moodle-scrapper/pkg/notifyer/telegram"
)

type Notifyer struct {
	service                  Service
	formatter                Formatter
	lastTimeNotifyedFilePath string
}

func NewTelegramNotifyer(cfg config.Config) Notifyer {
	tgService := telegram.NewTelegramService(
		cfg.TelegramBotKey,
		cfg.TelegramChatID,
	)

	formatterConfig := formatter.FormatConfig{
		UpdatesToCheck:   cfg.UpdatesToCheck,
		ToPrintOnUpdates: cfg.ToPrintOnUpdates,
		ToPrint:          cfg.ToPrint,
		ToCheckCreates:   false,
		ToCheckRemoves:   false,
	}
	fmter := formatter.NewFormatter(formatterConfig)

	return NewNotifyer(tgService, fmter, cfg.LastTimeNotifyedPath)
}

func NewNotifyer(
	service Service,
	formatter Formatter,
	lastTimeNotifyedFilePath string,
) Notifyer {
	return Notifyer{service, formatter, lastTimeNotifyedFilePath}
}

func (tn *Notifyer) SaveLastTimeNotifyed(timeNotifyed time.Time) error {
	f, err := os.OpenFile(tn.lastTimeNotifyedFilePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("couldn't save last time notifyed: %v", err)
	}
	defer f.Close()

	timeNotifyedStr := timeNotifyed.Format(time.Layout)
	_, err = f.Write([]byte(timeNotifyedStr))
	if err != nil {
		return fmt.Errorf("couldn't save last time notifyed: %v", err)
	}

	return nil
}

func (tn Notifyer) GetLastTimeNotifyed() (time.Time, error) {
	f, err := os.OpenFile(tn.lastTimeNotifyedFilePath, os.O_RDONLY, 0644)
	if err != nil {
		return time.Time{}, fmt.Errorf("couldn't get last time notifyed: %v", err)
	}
	defer f.Close()

	timeByte, err := io.ReadAll(f)
	if err != nil {
		return time.Time{}, fmt.Errorf("couldn't get last time notifyed: %v", err)
	}

	lastTimeNotifyedTime, err := time.Parse(time.Layout, string(timeByte))
	if err != nil {
		return time.Time{}, fmt.Errorf("couldn't get last time notifyed: %v", err)
	}

	return lastTimeNotifyedTime, nil
}

func (tn *Notifyer) SendUpdates(updates []formatter.CourseGradesChange) (int, error) {
	filteredUpdates := tn.formatter.FilterGradesChanges(updates)

	messages, err := tn.formatter.ConvertUpdatesToString(filteredUpdates, 4096)
	if err != nil {
		return 0, fmt.Errorf("failed to send updates: %v", err)
	}

	for _, msg := range messages {
		err = tn.service.Send(msg)
		if err != nil {
			return 0, fmt.Errorf("failed to send updates: %v", err)
		}
	}

	return len(messages), nil
}

type Service interface {
	Send(msg string) error
}

type Formatter interface {
	ConvertUpdatesToString(
		gradesChange []formatter.CourseGradesChange,
		maxMsgLen int,
	) ([]string, error)
	FilterGradesChanges(courseChanges []formatter.CourseGradesChange) []formatter.CourseGradesChange
}
