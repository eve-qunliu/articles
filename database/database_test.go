package database_test

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/eve-qunliu/articles/config"
	"github.com/eve-qunliu/articles/database"
	"github.com/eve-qunliu/articles/models"
)

func TestFindArticle(t *testing.T) {
	article := models.Article{Body: "z3", Date: "2018-06-12", ID: 123, Tags: []models.Tag{"sports", "music"}, Title: "z1"}
	testTable := []struct {
		Name           string
		Article        models.Article
		ExpectedError  error
		MockOperations func(m sqlmock.Sqlmock, err error, id string, row models.Article)
		ID             string
		VerifyError    func(t *testing.T, err error)
	}{
		{
			Name:          "Failure - db error",
			Article:       article,
			ExpectedError: errors.New("database error"),
			MockOperations: func(m sqlmock.Sqlmock, err error, id string, row models.Article) {
				expectArticleQuery(m).WillReturnError(err)
			},
			VerifyError: func(t *testing.T, err error) {
				assert.EqualError(t, err, "failed to retrieve article: database error", "Error")
			},
		},
		{
			Name:          "Success - no article found",
			Article:       article,
			ExpectedError: sql.ErrNoRows,
			MockOperations: func(m sqlmock.Sqlmock, err error, id string, row models.Article) {
				expectArticleQuery(m).WillReturnError(err)
			},
		},
		{
			Name:    "Success - article found",
			Article: article,
			MockOperations: func(m sqlmock.Sqlmock, err error, id string, row models.Article) {
				selectArticleWithID(m, id, row)
			},

			ID: "123",
		},
	}
	for _, d := range testTable {
		t.Run(d.Name, func(t *testing.T) {
			sqlDB, mock, err := sqlmock.New()
			require.NoError(t, err, "Unable to create SqlMock DB")
			db := sqlx.NewDb(sqlDB, "postgres")
			defer db.Close()

			d.MockOperations(mock, d.ExpectedError, d.ID, d.Article)
			config := config.Config{TagLimit: 3}
			provider := database.DBProvider{&config, db}

			article, err := provider.FindArticle(d.ID)

			assert.NoError(t, mock.ExpectationsWereMet(), "%s: DB Expectations", d.Name)
			if d.VerifyError != nil {
				d.VerifyError(t, err)
				return
			}
			assert.NoError(t, err, "Error: %s", d.Name)
			if article != nil {
				assert.Equal(t, d.Article, *article, "%s: article", d.Name)
			}
		})
	}
}

func TestCreateArticle(t *testing.T) {
	article := models.Article{Body: "z3", Date: "2018-06-12", ID: 123, Tags: []models.Tag{"sports", "music"}, Title: "z1"}
	testTable := []struct {
		Name           string
		Article        models.Article
		ExpectedError  error
		MockOperations func(m sqlmock.Sqlmock, err error, article models.Article)
		VerifyError    func(t *testing.T, err error)
	}{
		{
			Name:          "Failure - db error",
			Article:       article,
			ExpectedError: errors.New("database error"),
			MockOperations: func(m sqlmock.Sqlmock, err error, article models.Article) {
				expectCreateArticle(m).WillReturnError(err)
			},
			VerifyError: func(t *testing.T, err error) {
				assert.EqualError(t, err, "failed to insert article: database error", "Error")
			},
		},
		{
			Name:          "Failure - failed to create tags",
			Article:       article,
			ExpectedError: errors.New("database error"),
			MockOperations: func(m sqlmock.Sqlmock, err error, article models.Article) {
				createArticle(m, article)
				expectCreateTags(m).WillReturnError(err)
			},
			VerifyError: func(t *testing.T, err error) {
				assert.EqualError(t, err, "failed to create tags: database error", "Error")
			},
		},
		{
			Name:    "Success - create article with tags",
			Article: article,
			MockOperations: func(m sqlmock.Sqlmock, err error, article models.Article) {
				createArticle(m, article)
				createTags(m, article)
				createArticleTagMap(m, article)
			},
		},
	}
	for _, d := range testTable {
		t.Run(d.Name, func(t *testing.T) {
			sqlDB, mock, err := sqlmock.New()
			require.NoError(t, err, "Unable to create SqlMock DB")
			db := sqlx.NewDb(sqlDB, "postgres")
			defer db.Close()

			d.MockOperations(mock, d.ExpectedError, d.Article)
			config := config.Config{TagLimit: 3}
			provider := database.DBProvider{&config, db}

			err = provider.CreateArticle(&d.Article)

			assert.NoError(t, mock.ExpectationsWereMet(), "%s: DB Expectations", d.Name)
			if d.VerifyError != nil {
				d.VerifyError(t, err)
				return
			}
			assert.NoError(t, err, "Error: %s", d.Name)
		})
	}
}

func TestFindTag(t *testing.T) {
	testTable := []struct {
		Name           string
		Rows           map[string][]interface{}
		ExpectedError  error
		MockOperations func(m sqlmock.Sqlmock, result map[string][]interface{}, err error)
		VerifyError    func(t *testing.T, err error)
	}{
		{
			Name: "No data from database",
			MockOperations: func(m sqlmock.Sqlmock, result map[string][]interface{}, err error) {
				expectLatestArticleswithTag(m, emptyRows()).WillReturnError(err)
				expectArticlesCount(m, emptyRows()).WillReturnError(err)
				expectRelatedTags(m, emptyRows()).WillReturnError(err)
			},
		},
		{
			Name:          "Database error",
			ExpectedError: errors.New("database error"),
			MockOperations: func(m sqlmock.Sqlmock, result map[string][]interface{}, err error) {
				expectLatestArticleswithTag(m, emptyRows()).WillReturnError(err)
			},
			VerifyError: func(t *testing.T, err error) {
				assert.EqualError(t, err, "Failed to retrieve tag: database error", "Error")
			},
		},
		{
			Name: "With data from database",
			Rows: map[string][]interface{}{
				"articles": {"{'1','2'}"},
				"count":    {2},
				"related":  {"{'music','drama'}"},
			},
			MockOperations: func(m sqlmock.Sqlmock, result map[string][]interface{}, err error) {
				expectLatestArticleswithTag(m, mockedRows([]string{"ids"}, result["articles"]))
				expectArticlesCount(m, mockedRows([]string{"count"}, result["count"]))
				expectRelatedTags(m, mockedRows([]string{"related"}, result["related"]))
			},
		},
	}

	for _, d := range testTable {
		t.Run(d.Name, func(t *testing.T) {
			sqlDB, mock, err := sqlmock.New()
			require.NoError(t, err, "Unable to create SqlMock DB")
			db := sqlx.NewDb(sqlDB, "postgres")
			defer db.Close()

			d.MockOperations(mock, d.Rows, d.ExpectedError)
			provider := database.DBProvider{&config.Config{}, db}

			_, err = provider.FindTag("sports", "20180101")

			assert.NoError(t, mock.ExpectationsWereMet(), "%s: DB Expectations", d.Name)
			if d.VerifyError != nil {
				d.VerifyError(t, err)
				return
			}
			assert.NoError(t, err, "Error: %s", d.Name)
		})
	}
}

func expectArticleQuery(m sqlmock.Sqlmock) *sqlmock.ExpectedQuery {
	return m.ExpectQuery(`SELECT articles.id, articles.title, articles.body, articles.date, array_agg\(tags.name\) FROM articles, tags, tags_articles WHERE articles.id = tags_articles.article_id AND tags.id = tags_articles.tag_id AND articles.id = \$1 GROUP BY articles.id`)
}

func selectArticleWithID(m sqlmock.Sqlmock, id string, row models.Article) *sqlmock.ExpectedQuery {
	return expectArticleQuery(m).WithArgs(id).WillReturnRows(asMockArticleRow(row))
}

func asMockArticleRow(article models.Article) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{"id", "title", "body", "date", "tags"})
	var tags []string
	for _, tag := range article.Tags {
		tags = append(tags, string(tag))
	}
	rows.AddRow(article.ID, article.Title, article.Body, article.Date, fmt.Sprintf("{%s}", strings.Join(tags, ",")))
	return rows
}

func expectCreateArticle(m sqlmock.Sqlmock) *sqlmock.ExpectedQuery {
	return m.ExpectQuery(`INSERT INTO articles \(title, body, date\) VALUES \(\$1, \$2, \$3\) RETURNING id`)
}

func createArticle(m sqlmock.Sqlmock, row models.Article) *sqlmock.ExpectedQuery {
	return expectCreateArticle(m).WithArgs(row.Title, row.Body, row.Date).WillReturnRows(asMockIDRows([]int64{row.ID}))
}

func asMockIDRows(ids []int64) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{"id"})
	for _, i := range ids {
		rows.AddRow(i)
	}
	return rows
}

func expectCreateTags(m sqlmock.Sqlmock) *sqlmock.ExpectedQuery {
	return m.ExpectQuery(`INSERT INTO tags \(name\) VALUES \(\$1\),\(\$2\) ON CONFLICT \(name\) DO UPDATE SET name = EXCLUDED.name RETURNING id`)
}

func createTags(m sqlmock.Sqlmock, row models.Article) *sqlmock.ExpectedQuery {
	return expectCreateTags(m).WithArgs(row.Tags[0], row.Tags[1]).WillReturnRows(asMockIDRows([]int64{1, 2}))
}

func expectCreateArticleTagMap(m sqlmock.Sqlmock) *sqlmock.ExpectedQuery {
	return m.ExpectQuery(`INSERT INTO tags_articles \(article_id, tag_id\) VALUES \(123, 1\),\(123, 2\)`)
}

func createArticleTagMap(m sqlmock.Sqlmock, row models.Article) *sqlmock.ExpectedQuery {
	return expectCreateArticleTagMap(m).WillReturnRows(sqlmock.NewRows([]string{}))
}

func expectLatestArticleswithTag(m sqlmock.Sqlmock, rows *sqlmock.Rows) *sqlmock.ExpectedQuery {
	return m.ExpectQuery(`SELECT array_agg\(article_ids.id::text\) FROM
				\(SELECT articles.id AS id FROM tags, articles, tags_articles
				 WHERE tags.id = tags_articles.tag_id AND articles.id = tags_articles.article_id
			     	 AND tags.name = \$1 AND articles.date = \$2 ORDER BY articles.created_at DESC LIMIT 10\)
			      	 AS article_ids`).WillReturnRows(rows)
}

func expectArticlesCount(m sqlmock.Sqlmock, rows *sqlmock.Rows) *sqlmock.ExpectedQuery {
	return m.ExpectQuery(`SELECT COUNT\(articles.id\)
				 FROM tags, articles, tags_articles
				 WHERE tags.id = tags_articles.tag_id AND articles.id = tags_articles.article_id
			     	 AND tags.name = \$1 AND articles.date = \$2`).WillReturnRows(rows)
}

func expectRelatedTags(m sqlmock.Sqlmock, rows *sqlmock.Rows) *sqlmock.ExpectedQuery {
	return m.ExpectQuery(`SELECT array_agg\(DISTINCT\(tags.name\)\) FROM tags, tags_articles
			      WHERE tags.name != \$1 AND tags.id = tags_articles.tag_id
			      AND tags_articles.article_id IN \(SELECT articles.id
				 FROM tags, articles, tags_articles
				 WHERE tags.id = tags_articles.tag_id AND articles.id = tags_articles.article_id
			     	 AND tags.name = \$1 AND articles.date = \$2\)`).WillReturnRows(rows)
}

func mockedRows(fields []string, values []interface{}) *sqlmock.Rows {
	rows := sqlmock.NewRows(fields)

	for _, value := range values {
		rows.AddRow(value)
	}

	return rows
}

func emptyRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{})
}
