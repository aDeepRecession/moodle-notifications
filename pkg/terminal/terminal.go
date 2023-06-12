package terminal

import (
	"log"
	"time"

	"github.com/aDeepRecession/moodle-scrapper/pkg/config"
)

type Terminal struct {
	FailedRequestTimeout time.Duration
	CheckIntervalDelay   time.Duration
	log                  *log.Logger
}

func NewTerminal(cfg config.Config) Terminal {
	return Terminal{
		cfg.FailedRequestRepeatTimeout,
		cfg.CheckInterval,
		cfg.Logger,
	}
}

func (t Terminal) PrintError(err error) {
	t.log.Println(err)
}

func (t Terminal) PrintMsg(msg string) {
	t.log.Println(msg)
}

func (t Terminal) WaitFailedRequestRepeatInterval() {
	nextTimeCheck := t.getNextCheckTime(t.FailedRequestTimeout).Format(time.Layout)

	t.log.Printf(
		"Waiting... next check at %q",
		nextTimeCheck,
	)

	time.Sleep(t.FailedRequestTimeout)
}

func (terminal Terminal) WaitUntilNextCheck() {
	nextTimeCheck := terminal.getNextCheckTime(terminal.CheckIntervalDelay).Format(time.Layout)

	terminal.log.Printf(
		"Waiting... next check at %q",
		nextTimeCheck,
	)

	time.Sleep(terminal.CheckIntervalDelay)
}

func (t Terminal) getNextCheckTime(sleepDuration time.Duration) time.Time {
	return time.Now().Add(sleepDuration)
}
