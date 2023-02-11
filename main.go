package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aDeepRecession/moodle-scrapper/pkg/config"
	gradeshistory "github.com/aDeepRecession/moodle-scrapper/pkg/gradesHistory"
	moodleapi "github.com/aDeepRecession/moodle-scrapper/pkg/moodleAPI"
	moodletokensmanager "github.com/aDeepRecession/moodle-scrapper/pkg/moodleTokens"
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

	for {
		token, err := moodletokensmanager.GetTokens(cfg.MoodleCredentialsPath, cfg.Logger)
		if err != nil {
			cfg.Logger.Println(err)
			os.Exit(1)
		}

		cfg.Logger.Println("initializing moodleAPI...")
		moodleAPI, err := moodleapi.NewMoodleAPI(string(token), cfg.Logger)
		if err != nil {
			cfg.Logger.Println(err)
			os.Exit(1)
		}

		cfg.Logger.Println("getting moodle courses...")
		courses, err := moodleAPI.GetCourses()
		if err != nil {
			cfg.Logger.Println(err)
			os.Exit(1)
		}

		coursesGrades, err := getGradesForNonHiddenCourses(moodleAPI, courses)
		if err != nil {
			cfg.Logger.Println(err)
			cfg.Logger.Printf(
				"Waiting... next check at %q",
				getNextCheckTime(cfg.FailedRequestRepeatTimeout).Format(time.Layout),
			)
			time.Sleep(cfg.FailedRequestRepeatTimeout)
		}

		saveCfg := gradeshistory.SaveConfig{
			LastGradesPath:    cfg.LastGradesPath,
			GradesHistoryPath: cfg.GradesHistoryPath,
		}
		gradesHistory := gradeshistory.NewGradesHistory(saveCfg, cfg.Logger)

		updatesNum, err := gradesHistory.UpdateGradesHistory(coursesGrades)
		if err != nil {
			cfg.Logger.Println(err)
			continue
		}
		log.Printf(
			"have %v new updates",
			updatesNum,
		)

		messagesSended, err := sendNotificationOnLastUpdates(&notifyer, gradesHistory)
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
	gradesHistory gradeshistory.GradesHistory,
) (int, error) {
	lastTimeNotifyed, err := notifyer.GetLastTimeNotifyed()
	if err != nil {
		lastTimeNotifyed = time.Unix(0, 0)
	}

	updatesHistory, err := gradesHistory.GetGradesHistoryFromDate(lastTimeNotifyed)
	if err != nil {
		return 0, err
	}

	messagesSended, err := notifyer.SendUpdates(convertCourseGradesChange(updatesHistory))
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
	historyCourseGrades []gradeshistory.CourseGradesChange,
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

func getGradesForNonHiddenCourses(
	moodleAPI moodleapi.MoodleAPI,
	courses []moodleapi.Course,
) ([]gradeshistory.CourseGrades, error) {
	coursesGrades := []gradeshistory.CourseGrades{}
	for _, course := range courses {
		if course.Hidden {
			continue
		}

		grades, err := moodleAPI.GetCourseGrades(course)
		if err != nil {
			return nil, fmt.Errorf("failed to get course grades for %q: %v", course.Fullname, err)
		}

		coursesGrades = append(coursesGrades, gradeshistory.CourseGrades{
			Course: course,
			Grades: grades,
		})
	}
	return coursesGrades, nil
}
