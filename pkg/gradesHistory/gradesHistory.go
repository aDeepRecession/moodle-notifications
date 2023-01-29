package gradeshistory

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	moodleapi "github.com/aDeepRecession/moodle-scrapper/pkg/moodleAPI"
	moodlegrades "github.com/aDeepRecession/moodle-scrapper/pkg/moodleGrades"
)

type CourseGrades struct {
	Course moodleapi.Course
	Grades []moodlegrades.GradeReport
}

type SaveConfig struct {
	LastGradesPath    string
	GradesHistoryPath string
}

type GradesHistory struct {
	cfg SaveConfig
	log *log.Logger
}

func NewGradesHistory(cfg SaveConfig, log *log.Logger) GradesHistory {
	return GradesHistory{cfg, log}
}

func (gh GradesHistory) UpdateGradesHistory(newGrades []CourseGrades) error {
	oldGrades, err := gh.getOldGrades()
	if err != nil {
		return err
	}

	gc := newGradesComparator(gh.log)
	gc.compareCourseGrades(oldGrades, newGrades)

	err = gh.saveGrades(newGrades)
	if err != nil {
		return err
	}
	gh.log.Printf("saved new grades to %v\n", gh.cfg.GradesHistoryPath)

	return nil
}

func (gh GradesHistory) getOldGrades() ([]CourseGrades, error) {
	courseGradesFile, err := os.OpenFile(gh.cfg.LastGradesPath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read old grades from \"%v\": %v",
			gh.cfg.LastGradesPath,
			err,
		)
	}
	defer courseGradesFile.Close()

	courseGradesReader, err := io.ReadAll(courseGradesFile)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read old grades from \"%v\": %v",
			gh.cfg.LastGradesPath,
			err,
		)
	}

	var gradesJSON []CourseGrades
	err = json.Unmarshal(courseGradesReader, &gradesJSON)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read old grades from \"%v\": %v",
			gh.cfg.LastGradesPath,
			err,
		)
	}

	return gradesJSON, nil
}

func (gh GradesHistory) saveGrades(grades []CourseGrades) error {
	stream, err := json.MarshalIndent(grades, "", "\t")
	if err != nil {
		panic(err)
	}

	f, err := os.OpenFile(gh.cfg.LastGradesPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to save grades to file \"%v\": %v", gh.cfg.LastGradesPath, err)
	}
	defer f.Close()

	_, err = f.Write(stream)
	if err != nil {
		return fmt.Errorf("failed to save grades to file \"%v\": %v", gh.cfg.LastGradesPath, err)
	}

	return nil
}
