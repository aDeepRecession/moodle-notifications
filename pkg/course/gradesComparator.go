package course

import (
	"log"
	"sort"

	"github.com/r3labs/diff/v3"

	"github.com/aDeepRecession/moodle-scrapper/pkg/moodle"
)

type CourseGradesChange struct {
	Course            moodle.Course
	GradesTableChange []GradeRowChange
}

type GradeRowChange struct {
	ID     int
	Type   string
	Fields []string
	From   moodle.GradeReport
	To     moodle.GradeReport
}

type gradesComparator struct {
	log *log.Logger
}

func newGradesComparator(log *log.Logger) gradesComparator {
	return gradesComparator{log}
}

func (gc gradesComparator) compareCourseGrades(from, to []moodle.Course) []CourseGradesChange {
	gc.sortCourses(&from)
	gc.sortCourses(&to)

	courseGradesChange := []CourseGradesChange{}

	fromCourseInx := 0
	toCourseInx := 0
	for fromCourseInx < len(from) && toCourseInx < len(to) {

		fromCourse := from[fromCourseInx]
		toCourse := to[toCourseInx]

		gradesTableChange := []GradeRowChange{}

		newCourseAdded := fromCourse.ID > toCourse.ID
		if newCourseAdded {
			gradesTableChange = gc.compareGradeReports(nil, toCourse.Grades)
			toCourseInx++
		}

		oldCourseRemoved := fromCourse.ID < toCourse.ID
		if oldCourseRemoved {
			gradesTableChange = gc.compareGradeReports(fromCourse.Grades, nil)
			fromCourseInx++
		}

		theSameCourses := fromCourse.ID == toCourse.ID
		if theSameCourses {
			gradesTableChange = gc.compareGradeReports(fromCourse.Grades, toCourse.Grades)
			fromCourseInx++
			toCourseInx++
		}

		noUpdates := len(gradesTableChange) == 0
		if noUpdates {
			continue
		}

		courseGradesChanges := CourseGradesChange{
			Course:            fromCourse,
			GradesTableChange: gradesTableChange,
		}
		courseGradesChange = append(courseGradesChange, courseGradesChanges)
	}

	for fromCourseInx < len(from) {
		fromCourse := from[fromCourseInx]
		gradesTableChange := gc.compareGradeReports(fromCourse.Grades, nil)

		courseGradesChanges := CourseGradesChange{
			Course:            fromCourse,
			GradesTableChange: gradesTableChange,
		}
		courseGradesChange = append(courseGradesChange, courseGradesChanges)

		fromCourseInx++
	}

	for toCourseInx < len(to) {
		toCourse := to[toCourseInx]
		gradesTableChange := gc.compareGradeReports(nil, toCourse.Grades)

		courseGradesChanges := CourseGradesChange{
			Course:            toCourse,
			GradesTableChange: gradesTableChange,
		}
		courseGradesChange = append(courseGradesChange, courseGradesChanges)

		toCourseInx++
	}

	return courseGradesChange
}

func (gc gradesComparator) compareGradeReports(
	from, to []moodle.GradeReport,
) []GradeRowChange {
	gc.sortGradesRows(&from)
	gc.sortGradesRows(&to)

	fromGradeInx := 0
	toGradeInx := 0

	gradesTableChnages := []GradeRowChange{}

	for fromGradeInx < len(from) && toGradeInx < len(to) {
		fromGrade := from[fromGradeInx]
		toGrade := to[toGradeInx]

		newGradeAdded := fromGrade.ID > toGrade.ID
		if newGradeAdded {

			createdGrade := GradeRowChange{
				Type:   "create",
				ID:     fromGrade.ID,
				Fields: []string{},
				To:     fromGrade,
			}
			gradesTableChnages = append(gradesTableChnages, createdGrade)

			toGradeInx++
			continue
		}

		oldGradeRemoved := fromGrade.ID < toGrade.ID
		if oldGradeRemoved {

			removedGrade := GradeRowChange{
				Type:   "remove",
				ID:     toGrade.ID,
				Fields: []string{},
				From:   fromGrade,
			}
			gradesTableChnages = append(gradesTableChnages, removedGrade)

			fromGradeInx++
			continue
		}

		gradeChange := gc.compareGrades(fromGrade, toGrade)
		if gradeChange.Type != "nochange" {
			gradesTableChnages = append(gradesTableChnages, gradeChange)
		}

		fromGradeInx++
		toGradeInx++
	}

	for toGradeInx < len(to) {
		toGrade := to[toGradeInx]
		removedGrade := GradeRowChange{
			Type:   "create",
			ID:     toGrade.ID,
			Fields: []string{},
			To:     toGrade,
		}
		gradesTableChnages = append(gradesTableChnages, removedGrade)

		toGradeInx++
	}

	for fromGradeInx < len(from) {
		fromGrade := from[fromGradeInx]
		createdGrade := GradeRowChange{
			Type:   "remove",
			ID:     fromGrade.ID,
			Fields: []string{},
			From:   fromGrade,
		}
		gradesTableChnages = append(gradesTableChnages, createdGrade)

		fromGradeInx++
	}

	return gradesTableChnages
}

func (gc gradesComparator) compareGrades(from, to moodle.GradeReport) GradeRowChange {
	changes, err := diff.Diff(from, to)
	if err != nil {
		panic(err)
	}

	if len(changes) == 0 {
		return GradeRowChange{Type: "nochange"}
	}

	gradeRowChange := GradeRowChange{
		ID:     from.ID,
		Type:   "update",
		Fields: []string{},
		From:   from,
		To:     to,
	}

	for _, change := range changes {
		if gc.isFieldUpdateUseless(change.To.(string)) {
			continue
		}

		gradeRowChange.Fields = append(gradeRowChange.Fields, change.Path...)
	}

	areNoFieldsUpdated := len(gradeRowChange.Fields) == 0
	if areNoFieldsUpdated {
		return GradeRowChange{Type: "nochange"}
	}

	return gradeRowChange
}

func (gc gradesComparator) isFieldUpdateUseless(to string) bool {
	if to == "Error" || to == "-" {
		return true
	}

	return false
}

func (gc gradesComparator) sortGradesRows(rows *[]moodle.GradeReport) {
	sort.Slice((*rows), func(i, j int) bool {
		return (*rows)[i].ID < (*rows)[j].ID
	})
}

func (gc gradesComparator) sortCourses(grades *[]moodle.Course) {
	sort.Slice((*grades), func(i, j int) bool {
		return (*grades)[i].ID < (*grades)[j].ID
	})
}
