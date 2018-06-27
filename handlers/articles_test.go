package handlers_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/eve-qunliu/articles/config"
	"github.com/eve-qunliu/articles/handlers"
	"github.com/eve-qunliu/articles/models"
)

type dataProviderMock struct {
	mock.Mock
}

func (m *dataProviderMock) CreateArticle(a *models.Article) error {
	rtn := m.Called(a)
	return rtn.Error(0)
}

func (m *dataProviderMock) OnCreateArticle(a *models.Article) *mock.Call {
	return m.On("CreateArticle", mock.MatchedBy(equalArticle(a)))
}

func (m *dataProviderMock) FindArticle(id string) (*models.Article, error) {
	rtn := m.Called(id)
	return rtn.Get(0).(*models.Article), rtn.Error(1)
}

func (m *dataProviderMock) OnFindArticle(id string) *mock.Call {
	return m.On("FindArticle", id)
}

func (m *dataProviderMock) FindTag(name, date string) (*models.TagArticles, error) {
	rtn := m.Called(name, date)
	return rtn.Get(0).(*models.TagArticles), rtn.Error(1)
}

func (m *dataProviderMock) OnFindTag(name, date string) *mock.Call {
	return m.On("FindTag", name, date)
}

func equalArticle(expected *models.Article) func(a *models.Article) bool {
	return func(a *models.Article) bool {
		if expected.Body != a.Body || expected.Title != a.Title || expected.Date != a.Date {
			return false
		}
		if !reflect.DeepEqual(expected.Tags, a.Tags) {
			return false
		}

		return true
	}
}

func TestCreateArticles(t *testing.T) {
	article := models.Article{Body: "z3", Date: "2018-06-12", ID: 0, Tags: []models.Tag{"sports"}, Title: "z1"}
	data := []struct {
		Name              string
		Article           *models.Article
		ExpectedStatus    int
		MockCreateArticle func(m *dataProviderMock, a *models.Article)
		Payload           io.Reader
	}{
		{
			Name:           "Failure - invalid JSON payload",
			ExpectedStatus: http.StatusBadRequest,
			Payload:        strings.NewReader("Invalid JSON here"),
		},
		{
			Name:           "Failure - invalid article payload (missing title)",
			ExpectedStatus: http.StatusUnprocessableEntity,
			Payload:        strings.NewReader(`{"body":"z3","date":"2018-06-12","tags":["sports"]}`),
		},
		{
			Name:           "Failure - invalid article payload (missing body)",
			ExpectedStatus: http.StatusUnprocessableEntity,
			Payload:        strings.NewReader(`{"title":"z1","date":"2018-06-12","tags":["sports"]}`),
		},
		{
			Name:           "Failure - invalid article payload (two many tags)",
			ExpectedStatus: http.StatusUnprocessableEntity,
			Payload:        strings.NewReader(`{"title":"z1","body":"z3","date":"2018-06-12","tags":["sports","music","news","opinion"]}`),
		},
		{
			Name:           "Failure - invalid article payload (invalid date)",
			ExpectedStatus: http.StatusUnprocessableEntity,
			Payload:        strings.NewReader(`{"title":"z1","body":"z3","date":"aa","tags":["sports"]}`),
		},
		{
			Name:           "Failure - error to create an article",
			Article:        &article,
			ExpectedStatus: http.StatusInternalServerError,
			MockCreateArticle: func(m *dataProviderMock, a *models.Article) {
				m.OnCreateArticle(a).Return(errors.New("unknown error"))
			},
			Payload: strings.NewReader(`{"title":"z1","body":"z3","date":"2018-06-12","tags":["sports"]}`),
		},
		{
			Name:           "Success - create an article",
			Article:        &article,
			ExpectedStatus: http.StatusCreated,
			MockCreateArticle: func(m *dataProviderMock, a *models.Article) {
				m.OnCreateArticle(a).Return(nil)
			},
			Payload: strings.NewReader(`{"title":"z1","body":"z3","date":"2018-06-12","tags":["sports"]}`),
		},
		{
			Name:           "Success - create an article by removing duplicated tag names",
			Article:        &article,
			ExpectedStatus: http.StatusCreated,
			MockCreateArticle: func(m *dataProviderMock, a *models.Article) {
				m.OnCreateArticle(a).Return(nil)
			},
			Payload: strings.NewReader(`{"title":"z1","body":"z3","date":"2018-06-12","tags":["sports","sports"]}`),
		},
		{
			Name:           "Success - create an article by removing empty tag name",
			Article:        &article,
			ExpectedStatus: http.StatusCreated,
			MockCreateArticle: func(m *dataProviderMock, a *models.Article) {
				m.OnCreateArticle(a).Return(nil)
			},
			Payload: strings.NewReader(`{"title":"z1","body":"z3","date":"2018-06-12","tags":["", "sports"]}`),
		},
	}

	for _, d := range data {
		t.Run(d.Name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, err := http.NewRequest("POST", "", d.Payload)
			assert.NoError(t, err, "failed to create request")

			provider := new(dataProviderMock)
			if d.MockCreateArticle != nil {
				d.MockCreateArticle(provider, d.Article)
			}

			config := config.Config{TagLimit: 3}
			ah := handlers.ArticleHandler{Config: &config, Provider: provider}
			handler := ah.CreateArticles()

			handler(w, r)
			provider.Mock.AssertExpectations(t)
			assert.Equal(t, d.ExpectedStatus, w.Code, "expectedStatus code")
		})
	}
}

func TestFindArticles(t *testing.T) {
	article := models.Article{Body: "z3", Date: "2018-06-12", ID: 123, Tags: []models.Tag{"music", "sports"}, Title: "z1"}

	data := []struct {
		Name            string
		Article         *models.Article
		ExpectedStatus  int
		MockFindArticle func(m *dataProviderMock, id string, rtn *models.Article)
	}{
		{
			Name:           "Failure - query error",
			Article:        (*models.Article)(nil),
			ExpectedStatus: http.StatusInternalServerError,
			MockFindArticle: func(m *dataProviderMock, id string, a *models.Article) {
				m.OnFindArticle(id).Return(a, errors.New("unknown error"))
			},
		},
		{
			Name:           "Failure - article not exist",
			Article:        (*models.Article)(nil),
			ExpectedStatus: http.StatusNotFound,
			MockFindArticle: func(m *dataProviderMock, id string, a *models.Article) {
				m.OnFindArticle(id).Return(a, nil)
			},
		},
		{
			Name:           "Success - find article",
			Article:        (*models.Article)(&article),
			ExpectedStatus: http.StatusOK,
			MockFindArticle: func(m *dataProviderMock, id string, a *models.Article) {
				m.OnFindArticle(id).Return(a, nil)
			},
		},
	}

	for _, d := range data {
		t.Run(d.Name, func(t *testing.T) {
			id := "123"
			w := httptest.NewRecorder()
			r, err := http.NewRequest("GET", "", nil)
			assert.NoError(t, err, "failed to create request")
			r = mux.SetURLVars(r, map[string]string{"id": id})

			provider := new(dataProviderMock)
			if d.MockFindArticle != nil {
				d.MockFindArticle(provider, id, d.Article)
			}

			config := config.Config{TagLimit: 3}
			ah := handlers.ArticleHandler{Config: &config, Provider: provider}
			handler := ah.FindArticle()

			handler(w, r)
			provider.Mock.AssertExpectations(t)
			assert.Equal(t, d.ExpectedStatus, w.Code, "expectedStatus code")
		})
	}
}

func TestFindTag(t *testing.T) {
	tagArticles := models.TagArticles{Articles: pq.StringArray{"1", "2"}, Count: 2, RelatedTags: pq.StringArray{"music", "sports"}, Tag: "sports"}
	data := []struct {
		Name           string
		TagArticles    *models.TagArticles
		ExpectedStatus int
		MockFindTag    func(m *dataProviderMock, tag, date string, rtn *models.TagArticles)
	}{
		{
			Name:           "Failure - query error",
			TagArticles:    (*models.TagArticles)(nil),
			ExpectedStatus: http.StatusInternalServerError,
			MockFindTag: func(m *dataProviderMock, tag, date string, rtn *models.TagArticles) {
				m.OnFindTag(tag, date).Return(rtn, errors.New("unknown error"))
			},
		},
		{
			Name:           "Success - find tag articles",
			TagArticles:    &tagArticles,
			ExpectedStatus: http.StatusOK,
			MockFindTag: func(m *dataProviderMock, tag, date string, rtn *models.TagArticles) {
				m.OnFindTag(tag, date).Return(rtn, nil)
			},
		},
	}

	for _, d := range data {
		t.Run(d.Name, func(t *testing.T) {
			tag := "sports"
			date := "2018-06-12"
			w := httptest.NewRecorder()
			r, err := http.NewRequest("GET", "", nil)
			assert.NoError(t, err, "failed to create request")
			r = mux.SetURLVars(r, map[string]string{"tagName": tag, "date": date})

			provider := new(dataProviderMock)
			if d.MockFindTag != nil {
				d.MockFindTag(provider, tag, date, d.TagArticles)
			}

			config := config.Config{TagLimit: 3}
			ah := handlers.ArticleHandler{Config: &config, Provider: provider}
			handler := ah.FindTag()

			handler(w, r)
			provider.Mock.AssertExpectations(t)
			assert.Equal(t, d.ExpectedStatus, w.Code, "expectedStatus code")
		})
	}
}
