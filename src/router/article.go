package router

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/sebastiw/sidan-backend/src/data"
	"github.com/sebastiw/sidan-backend/src/models"
	ru "github.com/sebastiw/sidan-backend/src/router_util"
)

func NewArticleHandler(db data.Database) ArticleHandler {
	return ArticleHandler{db}
}

type ArticleHandler struct {
	db data.Database
}

func (ah ArticleHandler) createArticleHandler(w http.ResponseWriter, r *http.Request) {
	var a models.Article
	_ = json.NewDecoder(r.Body).Decode(&a)

	slog.Info(ru.GetRequestId(r), "article", a.Fmt())
	article, err := ah.db.CreateArticle(&a)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("unable to render the error page: %v", err.Error()), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(article)
}

func (ah ArticleHandler) readArticleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("unable to render the error page: %v", err.Error()), http.StatusInternalServerError)
	}

	article, err := ah.db.ReadArticle(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("unable to render the error page: %v", err.Error()), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(article)
}

func (ah ArticleHandler) updateArticleHandler(w http.ResponseWriter, r *http.Request) {
	var a models.Article
	_ = json.NewDecoder(r.Body).Decode(&a)

	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	slog.Debug(ru.GetRequestId(r), "article", a.Fmt())
	a.Id = int64(id)
	article, err := ah.db.UpdateArticle(&a)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("unable to render the error page: %v", err.Error()), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(article)
}

func (ah ArticleHandler) deleteArticleHandler(w http.ResponseWriter, r *http.Request) {
	var a models.Article
	_ = json.NewDecoder(r.Body).Decode(&a)

	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	slog.Debug(ru.GetRequestId(r), "article", a.Fmt())
	a.Id = int64(id)
	article, err := ah.db.DeleteArticle(&a)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("unable to render the error page: %v", err.Error()), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(article)
}

func (ah ArticleHandler) readAllArticleHandler(w http.ResponseWriter, r *http.Request) {
	take := MakeDefaultInt(r, "take", "20")
	skip := MakeDefaultInt(r, "skip", "0")
	articles, err := ah.db.ReadArticles(take, skip)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("unable to render the error page: %v", err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(articles)
}
