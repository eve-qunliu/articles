package models

import (
	"github.com/lib/pq"
)

type TagArticles struct {
	Articles    pq.StringArray `json:"articles,omitempty"`
	Count       int            `json:"count"`
	RelatedTags pq.StringArray `json:"related_tags,omitempty"`
	Tag         string         `json:"tag"`
}
