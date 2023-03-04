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

type FormatConfig struct {
	ToPrint          []string
	ToPrintOnUpdates []string
	UpdatesToCheck   []string
	ToCheckCreates   bool
	ToCheckRemoves   bool
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

func (f Formatter) ConvertUpdatesToString(
	allCoursesChanges []CourseGradesChange,
	maxMsgLengh int,
) ([]string, error) {
	messages := []string{}
	for _, courseChange := range allCoursesChanges {
		courseRelatedMessages := make([]string, 0, len(allCoursesChanges))

		courseTitle := f.getCourseTitle(courseChange.Course.Fullname)
		courseRelatedMessages = append(courseRelatedMessages, courseTitle+"\n\n")

		gradesChanges, err := f.parseGradeTable(courseChange.GradesTableChange)
		if err != nil {
			return nil, fmt.Errorf("failed to convert updates for print: %v", err)
		}
		if len(gradesChanges) == 0 {
			continue
		}

		courseRelatedMessages = append(courseRelatedMessages, gradesChanges...)

		resultMessages := f.concatenate(courseRelatedMessages, maxMsgLengh)

		messages = append(messages, resultMessages...)
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

func (f Formatter) concatenate(pieces []string, maxLength int) []string {
	concatenatedMessages := make([]string, 0, 2)

	curMsg := strings.Builder{}
	curMsg.Grow(maxLength)
	for _, piece := range pieces {
		lengthAfterConcatenation := curMsg.Len() + len(piece)
		if lengthAfterConcatenation > maxLength {
			concatenatedMessages = append(concatenatedMessages, curMsg.String())
			curMsg.Reset()
		}

		curMsg.WriteString(piece)
	}

	if curMsg.Len() > 0 {
		concatenatedMessages = append(concatenatedMessages, curMsg.String())
	}

	return concatenatedMessages
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

func (f Formatter) parseGradeTable(gradeChanges []GradeRowChange) ([]string, error) {
	changes := make([]string, 0, len(gradeChanges))

	for _, gradeChange := range gradeChanges {
		gradeChange, err := f.convertGradeChangeToString(gradeChange)
		if err != nil {
			return nil, err
		}
		if gradeChange == "" {
			continue
		}

		changes = append(changes, gradeChange+"\n\n")
	}

	return changes, nil
}

func (f Formatter) convertGradeChangeToString(rowChanges GradeRowChange) (string, error) {
	changesStr := strings.Builder{}

	if rowChanges.Type == "remove" {
		if !f.cfg.ToCheckCreates {
			return "", nil
		}

		changesStr.WriteString("(removed)")
		changesStr.WriteString("\n")
	}

	if rowChanges.Type == "create" {
		if !f.cfg.ToCheckCreates {
			return "", nil
		}

		changesStr.WriteString("(new)")
		changesStr.WriteString("\n")
	}

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

	for _, fieldToPrint := range f.cfg.ToPrintOnUpdates {
		changed := slices.Contains(rowChanges.Fields, fieldToPrint)

		if !changed {
			continue
		}

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

func (f Formatter) checkChangedToPrint() bool {
	return true
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
