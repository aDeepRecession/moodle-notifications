package main

import (
	"log"
	"os"
	"time"

	gradeshistory "github.com/aDeepRecession/moodle-scrapper/pkg/gradesHistory"
	moodleapi "github.com/aDeepRecession/moodle-scrapper/pkg/moodleAPI"
	moodletokensmanager "github.com/aDeepRecession/moodle-scrapper/pkg/moodleTokens"
)

var logger *log.Logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)

var checkIntervalDuraion = time.Minute * 10

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

		logger.Printf(
			"Waiting... next check at %q",
			getNextCheckTime(checkIntervalDuraion).Format(time.Layout),
		)

		time.Sleep(checkIntervalDuraion)
	}
}

func getNextCheckTime(sleepDuration time.Duration) time.Time {
	return time.Now().Add(sleepDuration)
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
