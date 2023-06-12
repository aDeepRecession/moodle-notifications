package main

import (
	"fmt"
	"os"
	"time"

	"github.com/aDeepRecession/moodle-scrapper/pkg/config"
	"github.com/aDeepRecession/moodle-scrapper/pkg/course"
	"github.com/aDeepRecession/moodle-scrapper/pkg/moodle"
	"github.com/aDeepRecession/moodle-scrapper/pkg/notifyer"
	"github.com/aDeepRecession/moodle-scrapper/pkg/notifyer/formatter"
	telegramnotifyer "github.com/aDeepRecession/moodle-scrapper/pkg/notifyer/telegram"
)

var configPath string = "./config.json"

func main() {
	cfg, err := getConfig(configPath)
	if err != nil {
		cfg.Logger.Println(err)
		os.Exit(1)
	}

	notifyer := getNotifyer(cfg)

	saveCfg := course.SaveConfig{
		LastGradesPath:    cfg.LastGradesPath,
		GradesHistoryPath: cfg.GradesHistoryPath,
	}

	for {
		token, err := moodle.GetTokens(cfg.MoodleCredentialsPath, cfg.Logger)
		if err != nil {
			cfg.Logger.Println(err)

			time.Sleep(cfg.FailedRequestRepeatTimeout)
			continue
		}

		cfg.Logger.Println("initializing moodleAPI...")
		moodleAPI, err := moodle.NewMoodle(string(token), cfg.Logger)
		if err != nil {
			cfg.Logger.Println(err)
			cfg.Logger.Printf(
				"Waiting... next check at %q",
				getNextCheckTime(cfg.FailedRequestRepeatTimeout).Format(time.Layout),
			)

			time.Sleep(cfg.FailedRequestRepeatTimeout)
			continue
		}

		cfg.Logger.Println("getting moodle grades...")
		coursesGrades, err := moodleAPI.GetNonHiddenCourses()
		if err != nil {
			cfg.Logger.Println(err)
			cfg.Logger.Printf(
				"Waiting... next check at %q",
				getNextCheckTime(cfg.FailedRequestRepeatTimeout).Format(time.Layout),
			)

			time.Sleep(cfg.FailedRequestRepeatTimeout)
			continue
		}

		grades := course.NewGrades(saveCfg, cfg.Logger)

		gradeChanges, err := grades.Compare(coursesGrades)
		if err != nil {
			cfg.Logger.Println(err)

			time.Sleep(cfg.FailedRequestRepeatTimeout)
			continue
		}

		isAnyChange := len(gradeChanges) > 0
		if isAnyChange {
			grades.Save(coursesGrades)
		}

		messagesSended, err := sendNotificationOnLastUpdates(&notifyer, gradeChanges)
		if err != nil {
			cfg.Logger.Println(err)
		}
		cfg.Logger.Printf("sended %v messages\n", messagesSended)

		cfg.Logger.Printf(
			"Waiting... next check at %q",
			getNextCheckTime(cfg.CheckInterval).Format(time.Layout),
		)
		time.Sleep(cfg.CheckInterval)
	}
}

func getConfig(configPath string) (config.Config, error) {
	f, err := os.OpenFile(configPath, os.O_RDONLY, 0644)
	if err != nil {
		return config.Config{}, fmt.Errorf("failed to get configuration: %v", err)
	}
	defer f.Close()

	cfg, err := config.NewConfig(f)
	if err != nil {
		return config.Config{}, fmt.Errorf("failed to get configuration: %v", err)
	}

	return cfg, nil
}

func sendNotificationOnLastUpdates(
	notifyer *notifyer.Notifyer,
	gradesChange []course.CourseGradesChange,
) (int, error) {
	messagesSended, err := notifyer.SendUpdates(convertCourseGradesChange(gradesChange))
	if err != nil {
		return 0, err
	}

	timeNotifyed := time.Now().Add(time.Second * 2)
	err = notifyer.SaveLastTimeNotifyed(timeNotifyed)
	if err != nil {
		return messagesSended, err
	}

	return messagesSended, err
}

func convertCourseGradesChange(
	historyCourseGrades []course.CourseGradesChange,
) []formatter.CourseGradesChange {
	formatterGrades := []formatter.CourseGradesChange{}
	for _, change := range historyCourseGrades {

		formatterTableChange := []formatter.GradeRowChange{}
		for _, historyTableChage := range change.GradesTableChange {

			formatterRowChange := formatter.GradeRowChange{
				Type:   historyTableChage.Type,
				Fields: historyTableChage.Fields,
				From:   formatter.GradeReport(historyTableChage.From),
				To:     formatter.GradeReport(historyTableChage.To),
			}
			formatterTableChange = append(formatterTableChange, formatterRowChange)
		}

		newGradeChange := formatter.CourseGradesChange{
			Course:            formatter.Course{Fullname: change.Course.Fullname},
			GradesTableChange: formatterTableChange,
		}
		formatterGrades = append(formatterGrades, newGradeChange)
	}

	return formatterGrades
}

func getNextCheckTime(sleepDuration time.Duration) time.Time {
	return time.Now().Add(sleepDuration)
}

func getNotifyer(cfg config.Config) notifyer.Notifyer {
	tgService := telegramnotifyer.NewTelegramNotifyer(
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

	return notifyer.NewNotifyer(fmter, tgService, cfg.LastTimeNotifyedPath)
}
