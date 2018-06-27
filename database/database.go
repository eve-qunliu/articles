package database

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/eve-qunliu/articles/config"
	"github.com/eve-qunliu/articles/models"
)

type DBProvider struct {
	Config     *config.Config
	Connection *sqlx.DB
}

func (db *DBProvider) dbString() string {
	env := db.Config
	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", env.DBUser, env.DBPassword, env.DBHost, env.DBName)
}

func NewProvider(config *config.Config) (*DBProvider, error) {
	provider := &DBProvider{Config: config}
	return provider, provider.connect()
}

func (db *DBProvider) connect() error {
	connection, err := sqlx.Connect("postgres", db.dbString())
	if err != nil {
		return errors.Wrap(err, "failed to connect DB")
	}
	db.Connection = connection
	return nil
}

func (db *DBProvider) CreateArticle(article *models.Article) error {
	err := db.Connection.QueryRowx(
		`INSERT INTO articles (title, body, date) VALUES ($1, $2, $3) RETURNING id`,
		article.Title,
		article.Body,
		article.Date,
	).Scan(&article.ID)

	if err != nil {
		return errors.Wrap(err, "failed to insert article")
	}

	var tags *sqlx.Rows
	if tags, err = db.createTags(article.Tags); err != nil {
		return errors.Wrap(err, "failed to create tags")
	}

	var ids []int64
	if ids, err = toIds(tags); err != nil {
		return err
	}

	return db.createArticleTagMap(article.ID, ids)
}

func (db *DBProvider) createTags(tags []models.Tag) (*sqlx.Rows, error) {
	valueIndexes, values := tagsValue(tags)
	statement := fmt.Sprintf("INSERT INTO tags (name) VALUES %s ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name RETURNING id", valueIndexes)
	return db.Connection.Queryx(statement, values...)
}

func (db *DBProvider) createArticleTagMap(article int64, tags []int64) error {
	statement := fmt.Sprintf("INSERT INTO tags_articles (article_id, tag_id) VALUES %s", articleTagsPairs(article, tags))
	return db.Connection.QueryRowx(statement).Err()
}

func (db *DBProvider) FindArticle(id string) (*models.Article, error) {
	article := &models.Article{}
	var tags pq.StringArray
	statement := fmt.Sprintf(`SELECT articles.id, articles.title, articles.body, articles.date, array_agg(tags.name)
				  FROM articles, tags, tags_articles WHERE articles.id = tags_articles.article_id AND
				  tags.id = tags_articles.tag_id AND articles.id = $1 GROUP BY articles.id`)
	err := db.Connection.QueryRowx(statement, id).Scan(&article.ID, &article.Title, &article.Body, &article.Date, &tags)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, errors.Wrap(err, "failed to retrieve article")
	}

	article.Tags = stringArrayToTags(tags)

	return article, nil
}

func (db *DBProvider) FindTag(tag string, date string) (*models.TagArticles, error) {
	tag = strings.ToLower(tag)
	tagArticle := &models.TagArticles{Tag: tag}

	fromSubStatement := `FROM tags, articles, tags_articles
			     WHERE tags.id = tags_articles.tag_id AND articles.id = tags_articles.article_id
			     AND tags.name = $1 AND articles.date = $2`

	articlesStatement := fmt.Sprintf(`SELECT array_agg(article_ids.id::text)
					  FROM (SELECT articles.id AS id %s ORDER BY articles.created_at DESC LIMIT 10)
					  AS article_ids`, fromSubStatement)

	articlesCountStatement := fmt.Sprintf(`SELECT COUNT(articles.id) %s`, fromSubStatement)

	relatedTagsStatement := fmt.Sprintf(`SELECT array_agg(DISTINCT(tags.name)) FROM tags, tags_articles
					     WHERE tags.name != $1 AND tags.id = tags_articles.tag_id
					     AND tags_articles.article_id IN (SELECT articles.id %s)`, fromSubStatement)

	statements := []string{articlesStatement, articlesCountStatement, relatedTagsStatement}
	targets := []interface{}{&tagArticle.Articles, &tagArticle.Count, &tagArticle.RelatedTags}

	for idx := range statements {
		err := db.Connection.QueryRowx(statements[idx], tag, date).Scan(targets[idx])

		if err != nil && err != sql.ErrNoRows {
			return nil, errors.Wrap(err, "Failed to retrieve tag")
		}
	}

	return tagArticle, nil
}
