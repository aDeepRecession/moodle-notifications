package gradeshistory

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	moodleapi "github.com/aDeepRecession/moodle-scrapper/pkg/moodleAPI"
	moodlegrades "github.com/aDeepRecession/moodle-scrapper/pkg/moodleGrades"
)

type CourseGrades struct {
	Course moodleapi.Course
	Grades []moodlegrades.GradeReport
}

type CourseGradesHistoryField struct {
	Time    time.Time
	Updates []CourseGradesChange
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

func (gh GradesHistory) GetGradesHistoryFromDate(
	fromTime time.Time,
) ([]CourseGradesChange, error) {
	history, err := gh.getChangeHistory()
	if err != nil {
		return nil, fmt.Errorf("failed to get last grades history: %v", err)
	}

	newHistory := []CourseGradesChange{}
	for _, field := range history {
		isFieldTooOld := fromTime.After(field.Time)
		if isFieldTooOld {
			continue
		}

		newHistory = append(newHistory, field.Updates...)
	}

	return newHistory, nil
}

func (gh GradesHistory) UpdateGradesHistory(newGrades []CourseGrades) error {
	oldGrades, err := gh.getOldGrades()
	if err != nil {
		gh.log.Println(err)
		oldGrades = []CourseGrades{}
	}

	gc := newGradesComparator(gh.log)

	gradeChanges := gc.compareCourseGrades(oldGrades, newGrades)

	coursesChangedNum := len(gradeChanges)
	if coursesChangedNum > 0 {
		newGradesHistoryField := CourseGradesHistoryField{
			Time:    time.Now(),
			Updates: gradeChanges,
		}

		err = gh.updateChangesHistory(newGradesHistoryField)
		if err != nil {
			return err
		}
		gh.log.Printf(
			"updated %q with %v courses changed\n",
			gh.cfg.GradesHistoryPath,
			coursesChangedNum,
		)
	} else {
		gh.log.Println("no changes")
	}

	err = gh.saveGrades(newGrades)
	if err != nil {
		return err
	}
	gh.log.Printf("saved new grades to %q\n", gh.cfg.LastGradesPath)

	return nil
}

func (gh GradesHistory) updateChangesHistory(newChanges CourseGradesHistoryField) error {
	oldChangeHistory, err := gh.getChangeHistory()
	if err != nil {
		oldChangeHistory = nil
	}
	mergedChanges := append([]CourseGradesHistoryField{newChanges}, oldChangeHistory...)

	changesJSON, err := json.MarshalIndent(mergedChanges, "", "\t")
	if err != nil {
		panic(err)
	}

	f, err := os.OpenFile(gh.cfg.GradesHistoryPath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf(
			"failed to update grade changes file \"%v\": %v",
			gh.cfg.GradesHistoryPath,
			err,
		)
	}
	defer f.Close()

	_, err = f.Write(changesJSON)
	if err != nil {
		return fmt.Errorf(
			"failed to update grade changes file \"%v\": %v",
			gh.cfg.GradesHistoryPath,
			err,
		)
	}

	return nil
}

func (gh GradesHistory) getChangeHistory() ([]CourseGradesHistoryField, error) {
	changeHistoryFile, err := os.OpenFile(gh.cfg.GradesHistoryPath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read old chage history from \"%v\": %v",
			gh.cfg.GradesHistoryPath,
			err,
		)
	}
	defer changeHistoryFile.Close()

	changeHistoryReader, err := io.ReadAll(changeHistoryFile)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read old chage history from \"%v\": %v",
			gh.cfg.GradesHistoryPath,
			err,
		)
	}

	var gradesChanges []CourseGradesHistoryField
	err = json.Unmarshal(changeHistoryReader, &gradesChanges)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read old chage history from \"%v\": %v",
			gh.cfg.GradesHistoryPath,
			err,
		)
	}

	return gradesChanges, nil
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
