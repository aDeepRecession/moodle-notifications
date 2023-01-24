package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"time"

	moodleapi "github.com/aDeepRecession/moodle-scrapper/pkg/moodleAPI"
	moodletokensmanager "github.com/aDeepRecession/moodle-scrapper/pkg/moodleTokens"
)

func main() {
	tokens, err := moodletokensmanager.GetTokens()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	log.Println("initializing moodleAPI...")
	moodleAPI, err := moodleapi.NewMoodleAPI(tokens.Token)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	log.Println("getting moodle courses...")
	courses, err := moodleAPI.GetCourses()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	// for _, course := range courses {
	course := courses[3]

	log.Printf("getting moodle course grades for \"%v\"", course.Fullname)
	grades, err := moodleAPI.GetCourseGrades(course)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	for i, grade := range grades {
		fmt.Printf("%02v: %v\n", i, grade)
	}
	// }
}

func printCoursesWithGrades() {
}

func printNonHidedCourses(courses []map[string]interface{}) {
	fmt.Println("Courses:")
	for _, course := range courses {
		if course["hidden"] == true {
			continue
		}

		unixTimeFloat, err := strconv.ParseFloat(fmt.Sprint(course["timemodified"]), 0)
		if err != nil {
			panic(err)
		}
		lastTimeModified := convertFloatUnixTimeToTime(unixTimeFloat)

		fmt.Printf(
			"%s\n progress=%v\n last_time_modified=%v\n students_count=%v\n\n",
			course["fullname"],
			course["progress"],
			lastTimeModified,
			course["enrolledusercount"],
		)
	}
}

func convertFloatUnixTimeToTime(floatUnixTime float64) time.Time {
	sec, dec := math.Modf(floatUnixTime)
	time := time.Unix(int64(sec), int64(dec*(1e9)))
	return time
}
