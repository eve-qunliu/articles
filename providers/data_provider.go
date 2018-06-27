package providers

import "github.com/eve-qunliu/articles/models"

type DataProvider interface {
	CreateArticle(*models.Article) error
	FindArticle(string) (*models.Article, error)
	FindTag(string, string) (*models.TagArticles, error)
}
