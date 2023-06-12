package main

import (
	"fmt"

	"github.com/aDeepRecession/moodle-scrapper/pkg/config"
	"github.com/aDeepRecession/moodle-scrapper/pkg/course"
	"github.com/aDeepRecession/moodle-scrapper/pkg/moodle"
	"github.com/aDeepRecession/moodle-scrapper/pkg/notifyer"
	"github.com/aDeepRecession/moodle-scrapper/pkg/notifyer/formatter"
	"github.com/aDeepRecession/moodle-scrapper/pkg/terminal"
)

var configPath string = "./config.json"

func main() {

	cfg := config.GetConfigFromPath(configPath)

	notifyer := notifyer.NewTelegramNotifyer(cfg)

	output := terminal.NewTerminal(cfg)

	saveCfg := course.SaveConfig{
		LastGradesPath:    cfg.LastGradesPath,
		GradesHistoryPath: cfg.GradesHistoryPath,
	}

	for {
		token, err := moodle.GetTokens(cfg.MoodleCredentialsPath, cfg.Logger)
		if err != nil {
			output.PrintError(err)
			output.WaitFailedRequestRepeatInterval()

			continue
		}

		output.PrintMsg("initializing moodleAPI...")
		moodleAPI, err := moodle.NewMoodle(token, cfg.Logger)
		if err != nil {
			output.PrintError(err)
			output.WaitFailedRequestRepeatInterval()

			continue
		}

		output.PrintMsg("getting moodle grades...")
		coursesGrades, err := moodleAPI.GetNonHiddenCourses()
		if err != nil {
			output.PrintError(err)
			output.WaitFailedRequestRepeatInterval()

			continue
		}

		grades := course.NewGrades(saveCfg, cfg.Logger)

		gradeChanges, err := grades.Compare(coursesGrades)
		if err != nil {
			output.PrintError(err)
			output.WaitFailedRequestRepeatInterval()

			continue
		}

		output.PrintMsg(fmt.Sprintf("found %v changes\n", len(gradeChanges)))

		grades.Save(coursesGrades)

		messagesSended, err := notifyer.SendUpdates(
			formatter.ConvertCourseGradesChange(gradeChanges),
		)
		if err != nil {
			output.PrintError(err)
		}
		output.PrintMsg(fmt.Sprintf("sended %v messages\n", messagesSended))

		output.WaitUntilNextCheck()
	}
}
