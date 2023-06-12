package course

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/aDeepRecession/moodle-scrapper/pkg/moodle"
)

type CourseGradesHistoryField struct {
	Time    time.Time
	Updates []CourseGradesChange
}

type SaveConfig struct {
	LastGradesPath    string
	GradesHistoryPath string
}

type Grades struct {
	cfg SaveConfig
	log *log.Logger
}

func NewGrades(cfg SaveConfig, log *log.Logger) Grades {
	return Grades{cfg, log}
}

func (grades Grades) Compare(
	newGrades []moodle.Course,
) ([]CourseGradesChange, error) {
	oldGrades, err := grades.getSaved()
	if err != nil {
		grades.log.Println(err)
		oldGrades = []moodle.Course{}
	}

	gc := newGradesComparator(grades.log)

	gradeChanges := gc.compareCourseGrades(oldGrades, newGrades)
	return gradeChanges, nil
}

func (grades Grades) Save(gradesToSave []moodle.Course) error {
	stream, err := json.MarshalIndent(gradesToSave, "", "\t")
	if err != nil {
		panic(err)
	}

	f, err := os.OpenFile(grades.cfg.LastGradesPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf(
			"failed to save grades to file \"%v\": %v",
			grades.cfg.LastGradesPath,
			err,
		)
	}
	defer f.Close()

	_, err = f.Write(stream)
	if err != nil {
		return fmt.Errorf(
			"failed to save grades to file \"%v\": %v",
			grades.cfg.LastGradesPath,
			err,
		)
	}

	return nil
}

func (grades Grades) getSaved() ([]moodle.Course, error) {
	courseGradesFile, err := os.OpenFile(grades.cfg.LastGradesPath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read old grades from \"%v\": %v",
			grades.cfg.LastGradesPath,
			err,
		)
	}
	defer courseGradesFile.Close()

	courseGradesReader, err := io.ReadAll(courseGradesFile)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read old grades from \"%v\": %v",
			grades.cfg.LastGradesPath,
			err,
		)
	}

	var gradesJSON []moodle.Course
	err = json.Unmarshal(courseGradesReader, &gradesJSON)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read old grades from \"%v\": %v",
			grades.cfg.LastGradesPath,
			err,
		)
	}

	return gradesJSON, nil
}
