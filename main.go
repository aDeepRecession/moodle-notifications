package main

import (
	"log"
	"os"

	gradeshistory "github.com/aDeepRecession/moodle-scrapper/pkg/gradesHistory"
	moodleapi "github.com/aDeepRecession/moodle-scrapper/pkg/moodleAPI"
	moodletokensmanager "github.com/aDeepRecession/moodle-scrapper/pkg/moodleTokens"
)

var logger *log.Logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)

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

	logger.Println("getting moodle courses...")
	courses, err := moodleAPI.GetCourses()
	if err != nil {
		logger.Println(err)
		os.Exit(1)
	}

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
}
