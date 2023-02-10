package main

import (
	"fmt"
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
	failedRequestRepeatTimeout               = time.Minute
	UpdatesToCheck             []string      = []string{"Grade", "Persentage", "Feedback"}
	ToPrint                    []string      = []string{"Title", "Grade", "Persentage", "Feedback"}
	telegramBotKey             string        = "5908469215:AAGB6TUPSN0aStQaGNZDjZRucOJHpj76G2E"
	telegramChatID             int           = 788782273
	checkInterval              time.Duration = time.Minute * 10
	LastGradesPath             string        = "./last_grades.json"
	GradesHistoryPath          string        = "./grades_history.json"
	credentialsPath            string        = "./credentials.json"
	lastTimeNotifyedPath       string        = "./last_time_notifyed_time"
)

func main() {
	notifyer := getNotifyer()

	for {
		token, err := moodletokensmanager.GetTokens(credentialsPath, logger)
		if err != nil {
			logger.Println(err)
			os.Exit(1)
		}

		logger.Println("initializing moodleAPI...")
		moodleAPI, err := moodleapi.NewMoodleAPI(string(token), logger)
		if err != nil {
			logger.Println(err)
			os.Exit(1)
		}

		logger.Println("getting moodle courses...")
		courses, err := moodleAPI.GetCourses()
		if err != nil {
			logger.Println(err)
			os.Exit(1)
		}

		coursesGrades, err := getGradesForNonHiddenCourses(moodleAPI, courses)
		if err != nil {
			logger.Println(err)
			logger.Printf(
				"Waiting... next check at %q",
				getNextCheckTime(failedRequestRepeatTimeout).Format(time.Layout),
			)
			time.Sleep(failedRequestRepeatTimeout)
		}

		saveCfg := gradeshistory.SaveConfig{
			LastGradesPath:    LastGradesPath,
			GradesHistoryPath: GradesHistoryPath,
		}
		gradesHistory := gradeshistory.NewGradesHistory(saveCfg, logger)

		updatesNum, err := gradesHistory.UpdateGradesHistory(coursesGrades)
		if err != nil {
			logger.Println(err)
			continue
		}
		log.Printf(
			"have %v new updates",
			updatesNum,
		)

		messagesSended, err := sendNotificationOnLastUpdates(&notifyer, gradesHistory)
		if err != nil {
			logger.Println(err)
		}
		logger.Printf("sended %v messages\n", messagesSended)

		logger.Printf(
			"Waiting... next check at %q",
			getNextCheckTime(checkInterval).Format(time.Layout),
		)
		time.Sleep(checkInterval)
	}
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

func getNotifyer() notifyer.Notifyer {
	tgService := telegramnotifyer.NewTelegramNotifyer(
		telegramBotKey,
		telegramChatID,
	)

	cfg := formatter.FormatConfig{
		UpdatesToCheck: UpdatesToCheck,
		ToPrint:        ToPrint,
		ToCheckCreates: false,
		ToCheckRemoves: false,
	}
	fmter := formatter.NewFormatter(cfg)

	return notifyer.NewNotifyer(fmter, tgService, lastTimeNotifyedPath)
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
