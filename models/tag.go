package models

import (
	"strings"

	"github.com/pkg/errors"
)

type Tag string

func (tag *Tag) Invalid() error {
	if len(string(*tag)) == 0 {
		return errors.New("tag is empty")
	}

	return nil
}

func uniqTags(tags []Tag) []Tag {
	set := make(map[Tag]bool)
	uniqTags := make([]Tag, 0, len(tags))

	for _, tag := range tags {
		if tag.Invalid() == nil {
			lowTag := Tag(strings.ToLower(string(tag)))
			set[lowTag] = true
		}
	}

	for tag := range set {
		uniqTags = append(uniqTags, tag)
	}

	return uniqTags
}
