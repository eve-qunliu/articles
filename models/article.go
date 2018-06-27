package models

import (
	"regexp"

	"github.com/pkg/errors"
)

type Article struct {
	Body  string `json:"body"`
	Date  string `json:"date"`
	ID    int64  `json:"id,string"`
	Tags  []Tag  `json:"tags"`
	Title string `json:"title"`
}

func (article *Article) invalidDate() error {
	matched, _ := regexp.MatchString("[0-9]{4}-[0-9]{2}-[0-9]{2}", article.Date)

	if !matched {
		return errors.New("date must have format YYYY-MM-DD")
	}

	return nil
}

func (article *Article) Invalid(tagLimit int) error {
	if len(article.Title) == 0 {
		return errors.New("title cannot be empty")
	}

	if len(article.Body) == 0 {
		return errors.New("body cannot be empty")
	}

	if len(article.Tags) > tagLimit {
		return errors.New("too many tags")
	}

	return article.invalidDate()
}

func (article *Article) Normalize(tagLimit int) error {
	if err := article.Invalid(tagLimit); err != nil {
		return err
	}

	article.Tags = uniqTags(article.Tags)

	return nil
}
