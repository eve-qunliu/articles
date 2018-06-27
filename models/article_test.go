package models_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eve-qunliu/articles/models"
)

const tagLimit = 3

func TestInvalid(t *testing.T) {
	data := []struct {
		Name        string
		Article     *models.Article
		VerifyError func(t *testing.T, err error)
	}{
		{
			Name:    "Success",
			Article: &models.Article{Body: "z3", Date: "2018-06-12", ID: 123, Tags: []models.Tag{"music", "sports"}, Title: "z1"},
		},
		{
			Name:    "Failure - empty body",
			Article: &models.Article{Date: "2018-06-12", ID: 123, Tags: []models.Tag{"music", "sports"}, Title: "z1"},
			VerifyError: func(t *testing.T, err error) {
				assert.EqualError(t, err, "body cannot be empty", "Error")
			},
		},
		{
			Name:    "Failure - empty title",
			Article: &models.Article{Body: "z3", Date: "2018-06-12", ID: 123, Tags: []models.Tag{"music", "sports"}},
			VerifyError: func(t *testing.T, err error) {
				assert.EqualError(t, err, "title cannot be empty", "Error")
			},
		},
		{
			Name:    "Failure - too many tags",
			Article: &models.Article{Body: "z3", Date: "2018-06-12", ID: 123, Tags: []models.Tag{"music", "sports", "a", "b"}, Title: "z1"},
			VerifyError: func(t *testing.T, err error) {
				assert.EqualError(t, err, "too many tags", "Error")
			},
		},
		{
			Name:    "Failure - invalid date format",
			Article: &models.Article{Body: "z3", Date: "2018", ID: 123, Tags: []models.Tag{"music", "sports"}, Title: "z1"},
			VerifyError: func(t *testing.T, err error) {
				assert.EqualError(t, err, "date must have format YYYY-MM-DD", "error")
			},
		},
	}

	for _, d := range data {
		t.Run(d.Name, func(t *testing.T) {
			err := d.Article.Invalid(tagLimit)
			if d.VerifyError != nil {
				d.VerifyError(t, err)
				return
			}
			assert.NoError(t, err, "unexpected error")
		})
	}
}

func TestNormalize(t *testing.T) {
	data := []struct {
		Name        string
		Article     *models.Article
		Tags        []models.Tag
		VerifyError func(t *testing.T, err error)
	}{
		{
			Name:    "Success",
			Article: &models.Article{Body: "z3", Date: "2018-06-12", ID: 123, Tags: []models.Tag{"music", "sports"}, Title: "z1"},
		},
		{
			Name:    "Failure - with empty body",
			Article: &models.Article{Date: "2018", ID: 123, Tags: []models.Tag{"music", "sports"}, Title: "z1"},
			VerifyError: func(t *testing.T, err error) {
				assert.EqualError(t, err, "body cannot be empty", "error")
			},
		},
		{
			Name:    "Success - with duplicated tags removed",
			Article: &models.Article{Body: "z3", Date: "2018-06-12", ID: 123, Tags: []models.Tag{"music", "music", "sports"}, Title: "z1"},
			Tags:    []models.Tag{"music", "sports"},
		},
		{
			Name:    "Success - with empty tag removed",
			Article: &models.Article{Body: "z3", Date: "2018-06-12", ID: 123, Tags: []models.Tag{"", "music", "sports"}, Title: "z1"},
			Tags:    []models.Tag{"music", "sports"},
		},
	}

	for _, d := range data {
		t.Run(d.Name, func(t *testing.T) {
			err := d.Article.Normalize(tagLimit)
			if d.VerifyError != nil {
				d.VerifyError(t, err)
				return
			}

			if d.Tags != nil {
				assert.Equal(t, d.Tags, d.Article.Tags, "abc")
			}

			assert.NoError(t, err, "unexpected error")
		})
	}
}
