package database

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/eve-qunliu/articles/models"
)

func stringArrayToTags(strArray pq.StringArray) []models.Tag {
	tags := make([]models.Tag, 0, len(strArray))

	for _, tag := range strArray {
		tags = append(tags, models.Tag(tag))
	}

	return tags
}

func articleTagsPairs(article int64, tags []int64) string {
	values := make([]string, 0, len(tags))

	for _, tag := range tags {
		value := fmt.Sprintf("(%d, %d)", article, tag)
		values = append(values, value)
	}

	return strings.Join(values, ",")
}

func tagsValue(tags []models.Tag) (string, []interface{}) {
	valueIndexes := make([]string, 0, len(tags))
	values := make([]interface{}, 0, len(tags))

	for idx, tag := range tags {
		valueIndexes = append(valueIndexes, fmt.Sprintf("($%d)", idx+1))
		values = append(values, string(tag))
	}

	return strings.Join(valueIndexes, ","), values
}

func toIds(rows *sqlx.Rows) ([]int64, error) {
	if rows == nil {
		return nil, errors.New("rows is nil")
	}

	ids := make([]int64, 0)

	defer rows.Close()

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "Failed to retrieve tag ids")
	}

	return ids, nil
}
