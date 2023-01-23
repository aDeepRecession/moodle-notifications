package moodleapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

var errGradeRowIsEmpty error = errors.New("grade row is empty")

type MoodleAPI struct {
	token  string
	userid string
}

func NewMoodleAPI(token string) (MoodleAPI, error) {
	moodleAPI := MoodleAPI{token: token}
	userid, err := moodleAPI.getUserID()
	if err != nil {
		return MoodleAPI{}, err
	}
	moodleAPI.userid = userid

	return moodleAPI, nil
}

func (api MoodleAPI) GetCourses() ([]map[string]interface{}, error) {
	data := map[string]string{
		"userid": api.userid,
	}

	coursesRes, err := api.moodleAPIRequest("core_enrol_get_users_courses", data)
	if err != nil {
		return nil, fmt.Errorf("failed to get courses: %v", err)
	}

	var courses []map[string]interface{}
	err = json.Unmarshal([]byte(coursesRes), &courses)
	if err != nil {
		return nil, fmt.Errorf("failed to get courses: %v", err)
	}

	return courses, nil
}

type GradeReport struct {
	Title        string
	Grade        string
	Persentage   string
	Feedback     string
	Contribution string
	Range        string
	Weight       string
}

func (api MoodleAPI) GetCourseGrades(courseid string) ([]GradeReport, error) {
	data := map[string]string{
		"userid":   api.userid,
		"courseid": courseid,
	}

	gradesRes, err := api.moodleAPIRequest("gradereport_user_get_grades_table", data)
	if err != nil {
		return nil, fmt.Errorf("failed to get course grades: %v", err)
	}

	gradesJSON := string(gradesRes)
	grades := api.parseGradeTable(string(gradesJSON))

	return grades, nil
}

func (api MoodleAPI) parseGradeTable(gradesJSON string) []GradeReport {
	gradeRows := gjson.Get(gradesJSON, "tables.0.tabledata").Array()
	gradesReport := make([]GradeReport, len(gradeRows))

	for rowInx, gradeRow := range gradeRows {

		gradeReport, err := api.parseGradeRow(gradeRow)
		if errors.Is(err, errGradeRowIsEmpty) {
			continue
		}
		if err != nil {
			panic(err)
		}

		gradesReport[rowInx] = gradeReport
	}

	return gradesReport
}

func (api MoodleAPI) parseGradeRow(gradeRow gjson.Result) (GradeReport, error) {
	titleUnparced := gradeRow.Get("itemname.content").String()
	if titleUnparced == "" {
		return GradeReport{}, errGradeRowIsEmpty
	}
	title := api.parseTitle(titleUnparced)

	grade := gradeRow.Get("grade.content").String()

	percentage := gradeRow.Get("percentage.content").String()

	weight := gradeRow.Get("weight.content").String()

	contributionToCourse := gradeRow.Get("contributiontocoursetotal.content").String()

	rangeUnparced := gradeRow.Get("range.content").String()
	gradeRange := api.parseRange(rangeUnparced)

	feedbackUnparced := gradeRow.Get("feedback.content").String()
	feedback := api.parseFeedback(feedbackUnparced)

	gradeReport := GradeReport{
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

func (api MoodleAPI) parseFeedback(feedbackUnparced string) string {
	feedback := api.getStringBetween(feedbackUnparced, "<div class=\"text_to_html\">", "</div>")

	return strings.ReplaceAll(feedback, "&ndash;", "-")
}

func (api MoodleAPI) parseRange(rangeUnparced string) string {
	return strings.ReplaceAll(rangeUnparced, "&ndash;", "-")
}

func (api MoodleAPI) parseTitle(unparsedTitle string) string {
	title := api.getStringBetween(unparsedTitle, "title=\"", "\"")
	return title
}

func (api MoodleAPI) getStringBetween(str string, startS string, endS string) string {
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

func (api MoodleAPI) getUserID() (string, error) {
	log.Println("getting moodle user id...")

	info, err := api.getCoreWebsiteInfo()
	if err != nil {
		return "", err
	}

	var infoJSON map[string]interface{}
	err = json.Unmarshal([]byte(info), &infoJSON)
	if err != nil {
		return "", err
	}

	userid := fmt.Sprint(infoJSON["userid"])
	return userid, nil
}

func (api MoodleAPI) getCoreWebsiteInfo() ([]byte, error) {
	info, err := api.moodleAPIRequest("core_webservice_get_site_info", nil)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (api MoodleAPI) moodleAPIRequest(
	requestFunction string,
	dataArgs map[string]string,
) ([]byte, error) {
	moodleURL := "https://moodle.innopolis.university/webservice/rest/server.php?moodlewsrestformat=json&wsfunction=" + requestFunction
	data := url.Values{
		"moodlewssettingfilter":  {"True"},
		"moodlewssettingfileurl": {"False"},
		"wsfunction":             {requestFunction},
		"wstoken":                {api.token},
	}
	for k, v := range dataArgs {
		data.Add(k, v)
	}

	client := http.Client{
		Timeout: time.Second * 5,
	}
	moodleReq, err := http.NewRequest(http.MethodPost, moodleURL, strings.NewReader(data.Encode()))
	moodleReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		return nil, fmt.Errorf("failed to make a request to moodle: %v", err)
	}

	response, err := client.Do(moodleReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make a request to moodle: %v", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to make a request to moodle: %v", err)
	}

	return body, nil
}
