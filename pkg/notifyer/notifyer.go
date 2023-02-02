package notifyer

import (
	"fmt"
	"time"

	"github.com/aDeepRecession/moodle-scrapper/pkg/notifyer/formatter"
)

type Notifyer struct {
	service          Service
	formatter        Formatter
	lastTimeNotifyed time.Time
}

func NewNotifyer(formatter Formatter, service Service, lastTimeNotifyed time.Time) Notifyer {
	return Notifyer{service, formatter, lastTimeNotifyed}
}

func (tn Notifyer) GetLastTimeModifyed() time.Time {
	return tn.lastTimeNotifyed
}

func (tn *Notifyer) SendUpdates(updates []formatter.CourseGradesChange) (int, error) {
	filteredUpdates := tn.formatter.FilterGradesChanges(updates)

	messages, err := tn.formatter.ConvertUpdatesToString(filteredUpdates)
	if err != nil {
		return 0, fmt.Errorf("failed to send updates: %v", err)
	}

	for _, msg := range messages {
		err = tn.service.Send(msg)
		if err != nil {
			return 0, fmt.Errorf("failed to send updates: %v", err)
		}
	}

	tn.lastTimeNotifyed = time.Now()

	return len(messages), nil
}

type Service interface {
	Send(msg string) error
}

type Formatter interface {
	ConvertUpdatesToString(gradesChange []formatter.CourseGradesChange) ([]string, error)
	FilterGradesChanges(courseChanges []formatter.CourseGradesChange) []formatter.CourseGradesChange
}
