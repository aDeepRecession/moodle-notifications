package main

import (
	"log"
	"os"
	"time"

	gradeshistory "github.com/aDeepRecession/moodle-scrapper/pkg/gradesHistory"
	moodleapi "github.com/aDeepRecession/moodle-scrapper/pkg/moodleAPI"
	moodletokensmanager "github.com/aDeepRecession/moodle-scrapper/pkg/moodleTokens"
	"github.com/aDeepRecession/moodle-scrapper/pkg/notifyer"
	"github.com/aDeepRecession/moodle-scrapper/pkg/notifyer/formatter"
	telegramnotifyer "github.com/aDeepRecession/moodle-scrapper/pkg/notifyer/telegram"
)

var (
	logger *log.Logger = log.New(
		os.Stdout,
		"",
		log.Ldate|log.Ltime|log.Lshortfile,
	)
	lastTimeNotifyedFilePath string = "./last_time_notifyed_time"
)

func main() {
	tokens, err := moodletokensmanager.GetTokens()
	if err != nil {
		logger.Println(err)
		os.Exit(1)
	}

	logger.Println("initializing moodleAPI...")
	moodleAPI, err := moodleapi.NewMoodleAPI(tokens.Token)
	if err != nil {
		logger.Println(err)
		os.Exit(1)
	}

	notifyer := getNotifyer()

	for {
		logger.Println("getting moodle courses...")
		courses, err := moodleAPI.GetCourses()
		if err != nil {
			logger.Println(err)
			os.Exit(1)
		}

		coursesGrades := getGradesForNonHiddenCourses(moodleAPI, courses)

		saveCfg := gradeshistory.SaveConfig{
			LastGradesPath:    "./last_grades.json",
			GradesHistoryPath: "./grades_history.json",
		}
		gradesHistory := gradeshistory.NewGradesHistory(saveCfg, logger)

		err = gradesHistory.UpdateGradesHistory(coursesGrades)
		if err != nil {
			logger.Println(err)
			os.Exit(1)
		}

		messagesSended, err := sendNotificationOnLastUpdates(&notifyer, gradesHistory)
		if err != nil {
			logger.Println(err)
		}
		logger.Printf("sended %v messages\n", messagesSended)

		checkIntervalDuraion := time.Minute * 10
		logger.Printf(
			"Waiting... next check at %q",
			getNextCheckTime(checkIntervalDuraion).Format(time.Layout),
		)
		time.Sleep(checkIntervalDuraion)
	}
}

func sendNotificationOnLastUpdates(
	notifyer *notifyer.Notifyer,
	gradesHistory gradeshistory.GradesHistory,
) (int, error) {
	timeNotifyed := time.Now()

	lastTimeNotifyed, err := notifyer.GetLastTimeNotifyed()
	if err != nil {
		lastTimeNotifyed = time.Now()
	}

	updatesHistory, err := gradesHistory.GetGradesHistoryFromDate(lastTimeNotifyed)
	if err != nil {
		return 0, err
	}

	messagesSended, err := notifyer.SendUpdates(convertCourseGradesChange(updatesHistory))
	if err != nil {
		return 0, err
	}

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

func getNotifyer() notifyer.Notifyer {
	tgService := telegramnotifyer.NewTelegramNotifyer(
		"5908469215:AAGB6TUPSN0aStQaGNZDjZRucOJHpj76G2E",
		788782273,
	)

	cfg := formatter.FormatConfig{
		UpdatesToCheck: []string{"Grade", "Persentage", "Feedback"},
		ToPrint:        []string{"Title", "Grade", "Persentage", "Feedback"},
		ToCheckCreates: true,
	}
	fmter := formatter.NewFormatter(cfg)

	return notifyer.NewNotifyer(fmter, tgService, lastTimeNotifyedFilePath)
}

func getGradesForNonHiddenCourses(
	moodleAPI moodleapi.MoodleAPI,
	courses []moodleapi.Course,
) []gradeshistory.CourseGrades {
	coursesGrades := []gradeshistory.CourseGrades{}
	for _, course := range courses {
		if course.Hidden {
			continue
		}

		grades, err := moodleAPI.GetCourseGrades(course)
		if err != nil {
			logger.Println(err)
			os.Exit(1)
		}

		coursesGrades = append(coursesGrades, gradeshistory.CourseGrades{
			Course: course,
			Grades: grades,
		})
	}
	return coursesGrades
}
