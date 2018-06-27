package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/eve-qunliu/articles/config"
	"github.com/eve-qunliu/articles/models"
	"github.com/eve-qunliu/articles/providers"
)

type response struct {
	Status  int
	Payload []byte
	err     error
}

func sendResponse(w http.ResponseWriter, resp *response) {
	if resp.err != nil {
		logger.Errorf("failed to handle request: %s", resp.err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.Status)
	w.Write(resp.Payload)
}

type ArticleHandler struct {
	Config   *config.Config
	Provider providers.DataProvider
}

func (ah *ArticleHandler) CreateArticles() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := &response{
			Status: http.StatusOK,
		}

		defer sendResponse(w, resp)

		body, err := ioutil.ReadAll(r.Body)

		if err != nil {
			resp.Status = http.StatusInternalServerError
			resp.err = err
			return
		}

		article := &models.Article{}
		err = json.Unmarshal(body, article)

		if err != nil {
			resp.Status = http.StatusBadRequest
			resp.err = err
			return
		}

		err = article.Normalize(ah.Config.TagLimit)

		if err != nil {
			resp.Status = http.StatusUnprocessableEntity
			resp.err = err
			return
		}
		err = ah.Provider.CreateArticle(article)

		if err != nil {
			resp.Status = http.StatusInternalServerError
			resp.err = err
			return
		}

		payload, err := json.Marshal(article)
		if err != nil {
			resp.Status = http.StatusInternalServerError
			resp.err = err
			return
		}

		resp.Status = http.StatusCreated
		resp.Payload = payload
	}
}

func (ah *ArticleHandler) FindArticle() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := &response{
			Status: http.StatusOK,
		}

		defer sendResponse(w, resp)

		vars := mux.Vars(r)
		article, err := ah.Provider.FindArticle(vars["id"])
		if err != nil {
			resp.Status = http.StatusInternalServerError
			resp.err = err
			return
		}

		if article == nil {
			resp.Status = http.StatusNotFound
			return
		}

		payload, err := json.Marshal(article)
		if err != nil {
			resp.Status = http.StatusInternalServerError
			resp.err = err
			return
		}

		resp.Status = http.StatusOK
		resp.Payload = payload
	}
}

func (ah *ArticleHandler) FindTag() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := &response{
			Status: http.StatusOK,
		}

		defer sendResponse(w, resp)

		vars := mux.Vars(r)
		article, err := ah.Provider.FindTag(vars["tagName"], formatDate(vars["date"]))

		if err != nil {
			resp.Status = http.StatusInternalServerError
			resp.err = err
			return
		}

		payload, err := json.Marshal(article)
		if err != nil {
			resp.Status = http.StatusInternalServerError
			resp.err = err
			return
		}

		resp.Status = http.StatusOK
		resp.Payload = payload
	}
}

func formatDate(date string) string {
	if len(date) != len("YYYYMMDD") {
		return date
	}

	return date[:4] + "-" + date[4:6] + "-" + date[6:]
}
