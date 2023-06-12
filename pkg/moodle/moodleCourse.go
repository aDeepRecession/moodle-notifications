package moodle

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
)

var errGradeRowIsNotGradeRow error = errors.New("row is not grade row")

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

type MoodleUser struct {
	api    moodleApi
	userid string
}

func NewMoodleUser(api moodleApi, userid string) MoodleUser {
	return MoodleUser{api, userid}
}

type moodleApi interface {
	MoodleAPIRequest(string, map[string]string) ([]byte, error)
}

func (mg MoodleUser) GetCourseGrades(courseid string) ([]GradeReport, error) {
	data := map[string]string{
		"userid":   mg.userid,
		"courseid": courseid,
	}

	gradesRes, err := mg.api.MoodleAPIRequest("gradereport_user_get_grades_table", data)
	if err != nil {
		return nil, fmt.Errorf("failed to get course grades: %v", err)
	}

	gradesJSON := string(gradesRes)
	grades := mg.parseGradeTable(string(gradesJSON))

	return grades, nil
}

func (mg MoodleUser) parseGradeTable(gradesJSON string) []GradeReport {
	gradeRows := gjson.Get(gradesJSON, "tables.0.tabledata").Array()
	gradesReport := []GradeReport{}

	for _, gradeRow := range gradeRows {

		gradeReport, err := mg.parseGradeRow(gradeRow)
		if errors.Is(err, errGradeRowIsNotGradeRow) {
			continue
		}
		if err != nil {
			panic(err)
		}

		gradesReport = append(gradesReport, gradeReport)
	}

	return gradesReport
}

func (mg MoodleUser) parseGradeRow(gradeRow gjson.Result) (GradeReport, error) {
	titleUnparced := gradeRow.Get("itemname.content").String()

	if !mg.isRowContainsGrade(titleUnparced) {
		return GradeReport{}, errGradeRowIsNotGradeRow
	}

	title := mg.parseTitle(titleUnparced)

	isTitleEmpty := strings.TrimSpace(title) == ""
	if isTitleEmpty {
		return GradeReport{}, errGradeRowIsNotGradeRow
	}

	idStr := gradeRow.Get("itemname.id").String()
	id, err := strconv.Atoi(mg.getStringBetween(idStr, "_", "_"))
	if err != nil {
		panic(err)
	}

	unparsedGrade := gradeRow.Get("grade.content").String()
	grade := mg.parseGrade(unparsedGrade)

	unparsedPercentage := gradeRow.Get("percentage.content").String()
	percentage := mg.parsePersentage(unparsedPercentage)

	weight := gradeRow.Get("weight.content").String()

	contributionToCourse := gradeRow.Get("contributiontocoursetotal.content").String()

	rangeUnparced := gradeRow.Get("range.content").String()
	gradeRange := mg.parseRange(rangeUnparced)

	feedbackUnparced := gradeRow.Get("feedback.content").String()
	feedback := mg.parseFeedback(feedbackUnparced)

	gradeReport := GradeReport{
		ID:           id,
		Title:        title,
		Grade:        grade,
		Persentage:   percentage,
		Feedback:     feedback,
		Contribution: contributionToCourse,
		Range:        gradeRange,
		Weight:       weight,
	}

	return gradeReport, nil
}

func (mg MoodleUser) parseGrade(unparsedGrade string) string {
	grade := mg.removeTags(unparsedGrade)

	if grade == "Error" || grade == "-" {
		return ""
	}

	return grade
}

func (mg MoodleUser) parsePersentage(unparsedPersentage string) string {
	persentage := mg.removeTags(unparsedPersentage)

	if persentage == "Error" || persentage == "-" {
		return ""
	}

	return persentage
}

func (mg MoodleUser) parseFeedback(feedbackUnparced string) string {
	feedback := mg.removeTags(feedbackUnparced)

	feedback = strings.ReplaceAll(feedback, "&ndash;", "")
	return strings.ReplaceAll(feedback, "&nbsp;", "")
}

func (mg MoodleUser) parseRange(rangeUnparced string) string {
	return strings.ReplaceAll(rangeUnparced, "&ndash;", "-")
}

func (mg MoodleUser) isRowContainsGrade(row string) bool {
	return strings.Contains(row, "class=\"gradeitemheader\"")
}

func (mg MoodleUser) parseTitle(unparsedTitle string) string {
	title := mg.getStringBetween(unparsedTitle, "title=\"", "\"")
	return title
}

func (mg MoodleUser) getStringBetween(str string, startS string, endS string) string {
	s := strings.Index(str, startS)
	if s == -1 {
		return ""
	}
	newS := str[s+len(startS):]
	e := strings.Index(newS, endS)
	if e == -1 {
		return ""
	}
	result := newS[:e]
	return result
}

func (mg MoodleUser) removeTags(str string) string {
	for strings.Contains(str, "<") {
		start := strings.Index(str, "<")
		end := strings.Index(str, ">")

		before := str[:start]
		after := str[end+1:]

		str = before + after
	}

	return str
}
