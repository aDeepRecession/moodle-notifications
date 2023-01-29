package gradeshistory

import (
	"testing"

	moodlegrades "github.com/aDeepRecession/moodle-scrapper/pkg/moodleGrades"
	"github.com/stretchr/testify/assert"
)

func TestGradesTablesComparison(t *testing.T) {
	t.Run("grade row update", func(t *testing.T) {
		gradeFrom := moodlegrades.GradeReport{ID: 5, Title: "Final exam", Grade: "-"}
		gradeTo := moodlegrades.GradeReport{ID: 5, Title: "FINAL EXAM", Grade: "60"}

		from := []moodlegrades.GradeReport{
			gradeFrom,
		}

		to := []moodlegrades.GradeReport{
			gradeTo,
		}

		gc := gradesComparator{}
		changelog := gc.compareGradeReports(from, to)

		expect := []GradeRowChange{
			{
				ID:     5,
				Type:   "update",
				Fields: []string{"Title", "Grade"},
				From:   gradeFrom,
				To:     gradeTo,
			},
		}
		assert.Equal(t, expect, changelog)
	})

	t.Run("several grade rows update", func(t *testing.T) {
		finalGradeFrom := moodlegrades.GradeReport{ID: 5, Title: "Final exam", Grade: "-"}
		finalGradeTo := moodlegrades.GradeReport{ID: 5, Title: "FINAL EXAM", Grade: "60"}

		midGradeFrom := moodlegrades.GradeReport{ID: 3, Title: "Mid exam", Grade: "-"}
		midGradeTo := moodlegrades.GradeReport{ID: 3, Title: "MID EXAM", Grade: "50"}

		from := []moodlegrades.GradeReport{
			midGradeFrom,
			finalGradeFrom,
		}

		to := []moodlegrades.GradeReport{
			finalGradeTo,
			midGradeTo,
		}

		gc := gradesComparator{}
		changelog := gc.compareGradeReports(from, to)

		expect := []GradeRowChange{
			{
				ID:     3,
				Type:   "update",
				Fields: []string{"Title", "Grade"},
				From:   midGradeFrom,
				To:     midGradeTo,
			},
			{
				ID:     5,
				Type:   "update",
				Fields: []string{"Title", "Grade"},
				From:   finalGradeFrom,
				To:     finalGradeTo,
			},
		}
		assert.Equal(t, expect, changelog)
	})

	t.Run("new grade row", func(t *testing.T) {
		gradeFrom := moodlegrades.GradeReport{}
		gradeTo := moodlegrades.GradeReport{ID: 1, Title: "FINAL EXAM", Grade: "60"}

		from := []moodlegrades.GradeReport{}
		to := []moodlegrades.GradeReport{gradeTo}

		gc := gradesComparator{}
		changelog := gc.compareGradeReports(from, to)

		expect := []GradeRowChange{
			{
				ID:     gradeTo.ID,
				Type:   "create",
				Fields: []string{},
				From:   gradeFrom,
				To:     gradeTo,
			},
		}
		assert.Equal(t, expect, changelog)
	})

	t.Run("row deleted", func(t *testing.T) {
		gradeFrom := moodlegrades.GradeReport{ID: 1, Title: "FINAL EXAM", Grade: "60"}
		gradeTo := moodlegrades.GradeReport{}

		from := []moodlegrades.GradeReport{gradeFrom}
		to := []moodlegrades.GradeReport{}

		gc := gradesComparator{}
		changelog := gc.compareGradeReports(from, to)

		expect := []GradeRowChange{
			{
				ID:     gradeFrom.ID,
				Type:   "remove",
				Fields: []string{},
				From:   gradeFrom,
				To:     gradeTo,
			},
		}
		assert.Equal(t, expect, changelog)
	})

	t.Run("course deleted", func(t *testing.T) {
	})

	t.Run("new course", func(t *testing.T) {
	})
}
