package handlers

import (
	"log"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/eve-qunliu/articles/config"
	"github.com/eve-qunliu/articles/providers"
)

var logger *zap.SugaredLogger

func init() {
	baseLogger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}

	logger = baseLogger.Sugar()
	defer logger.Sync()
}

func NewHandler(config *config.Config, provider providers.DataProvider) *mux.Router {
	router := mux.NewRouter()
	article := &ArticleHandler{Config: config, Provider: provider}

	router.HandleFunc("/articles", article.CreateArticles()).
		Methods("POST")
	router.HandleFunc("/articles/{id}", article.FindArticle()).
		Methods("GET")
	router.HandleFunc("/tag/{tagName}/{date}", article.FindTag()).
		Methods("GET")

	return router
}
