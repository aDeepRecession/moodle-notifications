package formatter

import (
	"fmt"
	"strings"

	"golang.org/x/exp/slices"
)

type Formatter struct {
	cfg FormatConfig
}

func NewFormatter(cfg FormatConfig) Formatter {
	return Formatter{cfg}
}

func (f Formatter) ConvertUpdatesToString(gradesChanges []CourseGradesChange) ([]string, error) {
	messages := []string{}
	for _, courseChange := range gradesChanges {
		msg := strings.Builder{}

		courseTitle := f.getCourseTitle(courseChange.Course.Fullname)
		msg.WriteString(courseTitle)
		msg.WriteString("\n\n")

		gradesChanges, err := f.convertGradeTableToString(courseChange.GradesTableChange)
		if err != nil {
			return nil, fmt.Errorf("failed to convert updates for print: %v", err)
		}
		msg.WriteString(gradesChanges)
		msg.WriteString("\n\n")

		messages = append(messages, msg.String())
	}

	return messages, nil
}

func (f Formatter) FilterGradesChanges(courseChanges []CourseGradesChange) []CourseGradesChange {
	filteredCourseChange := []CourseGradesChange{}
	for _, courseChange := range courseChanges {

		filteredGradeChange := f.filterGradeRows(courseChange.GradesTableChange)

		if len(filteredGradeChange) == 0 {
			continue
		}

		newCourseGradesChange := CourseGradesChange{
			Course:            courseChange.Course,
			GradesTableChange: filteredGradeChange,
		}
		filteredCourseChange = append(filteredCourseChange, newCourseGradesChange)
	}

	return filteredCourseChange
}

func (f Formatter) filterGradeRows(gradeRows []GradeRowChange) []GradeRowChange {
	filteredGradeChange := []GradeRowChange{}
	for _, gradeChange := range gradeRows {
		isUpdateNotTracked := gradeChange.Type == "update" &&
			!f.doesContainSomeUpdateToCheck(gradeChange)

		if isUpdateNotTracked {
			continue
		}

		filteredGradeChange = append(filteredGradeChange, gradeChange)
	}

	return filteredGradeChange
}

func (f Formatter) doesContainSomeUpdateToCheck(gradeChange GradeRowChange) bool {
	for _, changedField := range f.cfg.UpdatesToCheck {
		if slices.Contains(gradeChange.Fields, changedField) {
			return true
		}
	}

	return false
}

func (f Formatter) getCourseTitle(courseName string) string {
	return fmt.Sprintf("%s:", courseName)
}

func (f Formatter) convertGradeTableToString(gradeChanges []GradeRowChange) (string, error) {
	changesStr := strings.Builder{}

	for _, gradeChange := range gradeChanges {
		gradeChange, err := f.convertGradeChangeToString(gradeChange)
		if err != nil {
			return "", err
		}
		changesStr.WriteString(gradeChange)

		changesStr.WriteString("\n\n")
	}

	return changesStr.String(), nil
}

func (f Formatter) convertGradeChangeToString(rowChanges GradeRowChange) (string, error) {
	changesStr := strings.Builder{}

	for _, fieldToPrint := range f.cfg.ToPrint {
		changed := slices.Contains(rowChanges.Fields, fieldToPrint)

		fieldChange, err := f.convertGradeField(
			rowChanges.From,
			rowChanges.To,
			fieldToPrint,
			changed,
		)
		if err != nil {
			return "", err
		}

		changesStr.WriteString(fieldChange)

		changesStr.WriteString("\n")
	}

	return changesStr.String(), nil
}

func (f Formatter) convertGradeField(
	from, to GradeReport,
	field string,
	changed bool,
) (string, error) {
	fieldValueTo, err := f.getFieldValue(field, to)
	if err != nil {
		return "", err
	}

	fieldValueFrom, err := f.getFieldValue(field, from)
	if err != nil {
		return "", err
	}

	fieldChangeMsg := ""
	if !changed {
		fieldChangeMsg = fmt.Sprintf("%s:  %q", field, fieldValueTo)
	} else {
		fieldChangeMsg = fmt.Sprintf("%s:  %q  ->  %q", field, fieldValueFrom, fieldValueTo)
	}

	return fieldChangeMsg, nil
}

func (f Formatter) getFieldValue(field string, gradeReport GradeReport) (string, error) {
	switch field {
	case "Title":
		return gradeReport.Title, nil
	case "Grade":
		return gradeReport.Grade, nil
	case "Persentage":
		return gradeReport.Persentage, nil
	case "Feedback":
		return gradeReport.Feedback, nil
	case "Range":
		return gradeReport.Range, nil
	case "Weight":
		return gradeReport.Weight, nil
	case "Contribution":
		return gradeReport.Contribution, nil
	default:
		return "", fmt.Errorf("bad field name to print = %q", field)
	}
}

type FormatConfig struct {
	ToPrint        []string
	UpdatesToCheck []string
	ToCheckCreates bool
}

type CourseGradesChange struct {
	Course            Course
	GradesTableChange []GradeRowChange
}

type Course struct {
	Fullname string `json:"fullname"`
}

type GradeRowChange struct {
	Type   string
	Fields []string
	From   GradeReport
	To     GradeReport
}

type GradeReport struct {
	ID           int
	Title        string
	Grade        string
	Persentage   string
	Feedback     string
	Contribution string
	Range        string
	Weight       string
}
