package moodleapi

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	moodlegrades "github.com/aDeepRecession/moodle-scrapper/pkg/moodleGrades"
)

type MoodleAPI struct {
	token  string
	userid string
}

type Course struct {
	ID                int    `json:"id"`
	Timemodified      int64  `json:"timemodified"`
	Fullname          string `json:"fullname"`
	Enrolledusercount int    `json:"enrolledusercount"`
	Startdate         int64  `json:"startdate"`
	Enddate           int64  `json:"enddate"`
	Hidden            bool   `json:"hidden"`
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

func (api MoodleAPI) GetCourseGrades(course Course) ([]moodlegrades.GradeReport, error) {
	gradesManager := moodlegrades.NewMoodleGrades(api, api.userid)

	courseGrades, err := gradesManager.GetCourseGrades(fmt.Sprint(course.ID))
	if err != nil {
		return nil, err
	}

	return courseGrades, nil
}

func (api MoodleAPI) GetCourses() ([]Course, error) {
	data := map[string]string{
		"userid": api.userid,
	}

	coursesJSON, err := api.MoodleAPIRequest("core_enrol_get_users_courses", data)
	if err != nil {
		return nil, fmt.Errorf("failed to get courses: %v", err)
	}

	courses, err := api.parseCoursesJSON(coursesJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to get courses: %v", err)
	}

	return courses, nil
}

func (api MoodleAPI) parseCoursesJSON(coursesJSON []byte) ([]Course, error) {
	var courses []Course
	err := json.Unmarshal(coursesJSON, &courses)
	if err != nil {
		return nil, fmt.Errorf("failed to parse courses: %v", err)
	}

	return courses, nil
}

func (api MoodleAPI) getUserID() (string, error) {
	if api.userid != "" {
		return api.userid, nil
	}

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
	info, err := api.MoodleAPIRequest("core_webservice_get_site_info", nil)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (api MoodleAPI) MoodleAPIRequest(
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
		Timeout: time.Second * 10,
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
