package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"reflect"
	"rwa/internal/handlers/middleware"
	"rwa/internal/models"
	"rwa/internal/repositories"
)

type ArticleHandler struct {
	articleRepo *repositories.ArticleRepository
}

func NewArticleHandler(articleRepo *repositories.ArticleRepository) *ArticleHandler {
	return &ArticleHandler{
		articleRepo: articleRepo,
	}
}

func (ah *ArticleHandler) Add(w http.ResponseWriter, r *http.Request) {
	session, err := middleware.GetSessionFromContext(r.Context())
	log.Printf("ArticleHandler: Add: session: %#v\n, err: %s", session, err)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	mapArticle := make(map[string]*models.Article)
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &mapArticle)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	article := mapArticle["article"]

	articleDB, err := ah.articleRepo.Add(session.UserID, article)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := make(map[string]*models.ArticleDto)
	resp["Article"] = &models.ArticleDto{
		Author: &models.UserDto{
			Username: articleDB.Author.Username,
			BIO:      articleDB.Author.BIO,
		},
		Body:        articleDB.Body,
		CreatedAt:   articleDB.CreatedAt,
		UpdatedAt:   articleDB.UpdatedAt,
		Description: articleDB.Description,
		TagList:     articleDB.TagList,
		Title:       articleDB.Title,
	}

	respJson, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(respJson)
}

func (ah *ArticleHandler) Get(w http.ResponseWriter, r *http.Request) {
	var articles []*models.Article
	var err error
	resp := make(map[string]interface{})
	filterArticle := &models.FilterArticle{}
	filterArticle.Author = r.FormValue("author")
	filterArticle.Tag = r.FormValue("tag")
	emptyFilterArticle := &models.FilterArticle{}
	if reflect.DeepEqual(filterArticle, emptyFilterArticle) {
		log.Printf("ArticleHandler: Get: branch filterOFF: %#v\n", filterArticle)
		articles, err = ah.articleRepo.Get()
		log.Printf("ArticleHandler: Get: branch filterOFF articles: %#v\n", articles)
	} else {
		log.Printf("ArticleHandler: Get: branch filterON: %#v\n", filterArticle)
		articles, err = ah.articleRepo.GetByFilter(filterArticle)
		log.Printf("ArticleHandler: Get: branch filterON articles: %#v\n", articles)
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	articlesDto := make([]*models.ArticleDto, len(articles))
	i := 0
	for _, article := range articles {
		articlesDto[i] = &models.ArticleDto{
			Author: &models.UserDto{
				Username: article.Author.Username,
				BIO:      article.Author.BIO,
			},
			Body:        article.Body,
			CreatedAt:   article.CreatedAt,
			UpdatedAt:   article.UpdatedAt,
			Description: article.Description,
			TagList:     article.TagList,
			Title:       article.Title,
		}
		i++
	}
	resp["Articles"] = articlesDto
	resp["ArticlesCount"] = len(articlesDto)
	respJson, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(respJson)
}

func (ah *ArticleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		ah.Add(w, r)
	case http.MethodGet:
		ah.Get(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
