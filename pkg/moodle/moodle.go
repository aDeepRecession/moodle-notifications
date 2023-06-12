package moodle

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Moodle struct {
	token  string
	userid string
	log    *log.Logger
}

type Course struct {
	ID                int    `json:"id"`
	Timemodified      int64  `json:"timemodified"`
	Fullname          string `json:"fullname"`
	Enrolledusercount int    `json:"enrolledusercount"`
	Startdate         int64  `json:"startdate"`
	Enddate           int64  `json:"enddate"`
	Hidden            bool   `json:"hidden"`
	Grades            []GradeReport
}

func NewMoodle(token string, log *log.Logger) (Moodle, error) {
	moodleAPI := Moodle{token: token, log: log}
	userid, err := moodleAPI.getUserID()
	if err != nil {
		return Moodle{}, err
	}
	moodleAPI.userid = userid

	return moodleAPI, nil
}

func (moodle Moodle) GetCourseGrades(course Course) ([]GradeReport, error) {
	moodleUser := NewMoodleUser(moodle, moodle.userid)

	courseGrades, err := moodleUser.GetCourseGrades(fmt.Sprint(course.ID))
	if err != nil {
		return nil, err
	}

	return courseGrades, nil
}

func (moodle Moodle) GetNonHiddenCourses() ([]Course, error) {
	courses, err := moodle.GetCourses()
	if err != nil {
		return nil, fmt.Errorf("failed to get course grades %v", err)
	}

	nonHiddenCourses := make([]Course, 0)
	for _, course := range courses {
		if course.Hidden {
			continue
		}

		nonHiddenCourses = append(nonHiddenCourses, course)
	}

	return nonHiddenCourses, nil
}

func (moodle Moodle) GetCourses() ([]Course, error) {
	data := map[string]string{
		"userid": moodle.userid,
	}

	coursesJSON, err := moodle.MoodleAPIRequest("core_enrol_get_users_courses", data)
	if err != nil {
		return nil, fmt.Errorf("failed to get courses: %v", err)
	}

	courses, err := moodle.parseCoursesJSON(coursesJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to get courses: %v", err)
	}

	for _, course := range courses {
		if course.Hidden {
			continue
		}

		courseGrades, err := moodle.GetCourseGrades(course)
		if err != nil {
			return nil, fmt.Errorf("failed to get course grades for %q: %v", course.Fullname, err)
		}

		course.Grades = courseGrades
	}

	return courses, nil
}

func (api Moodle) parseCoursesJSON(coursesJSON []byte) ([]Course, error) {
	var courses []Course
	err := json.Unmarshal(coursesJSON, &courses)
	if err != nil {
		return nil, fmt.Errorf("failed to parse courses: %v", err)
	}

	return courses, nil
}

func (api Moodle) getUserID() (string, error) {
	if api.userid != "" {
		return api.userid, nil
	}

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

func (api Moodle) getCoreWebsiteInfo() ([]byte, error) {
	info, err := api.MoodleAPIRequest("core_webservice_get_site_info", nil)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (api Moodle) MoodleAPIRequest(
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

	if strings.Contains(string(body), "Invalid token") {
		return nil, fmt.Errorf("invalid token")
	}

	return body, nil
}

func (api Moodle) isTokenGood() bool {
	_, err := api.getCoreWebsiteInfo()
	return err == nil
}
