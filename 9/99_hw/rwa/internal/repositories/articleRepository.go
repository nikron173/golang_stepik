package repositories

import (
	"fmt"
	"log"
	"rwa/internal/common"
	"rwa/internal/models"
	"slices"
	"sync"
	"time"
)

type ArticleRepository struct {
	mu       sync.RWMutex
	userRepo *UserRepository
	articles map[string][]*models.Article
}

func NewArticleRepository() *ArticleRepository {
	return &ArticleRepository{
		mu:       sync.RWMutex{},
		articles: make(map[string][]*models.Article),
		userRepo: NewUserRepository(),
	}
}

func (ar *ArticleRepository) Add(userID string, article *models.Article) (*models.Article, error) {
	log.Printf("ArticleRepository: Add: ")
	user, err := ar.userRepo.Get(userID)
	if err != nil {
		return nil, fmt.Errorf("User not found")
	}
	article.CreatedAt = time.Now()
	article.UpdatedAt = article.CreatedAt
	article.Author = user
	article.Slug = common.RandStringRunes(20)
	ar.mu.Lock()
	if _, ok := ar.articles[userID]; !ok {
		ar.articles[userID] = make([]*models.Article, 0, 1)
	}
	ar.articles[userID] = append(ar.articles[userID], article)
	ar.mu.Unlock()
	log.Printf("ArticleRepository: Add: articlesUser: %#v\n", ar.articles[userID])
	return article, nil
}

func (ar *ArticleRepository) GetByFilter(filter *models.FilterArticle) ([]*models.Article, error) {
	articles := make([]*models.Article, 0)
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	if filter.Author != "" && filter.Tag != "" {
		for _, articleAuthor := range ar.articles {
			for _, article := range articleAuthor {
				if filter.Author == article.Author.Username &&
					slices.Contains(article.TagList, filter.Tag) {
					articles = append(articles, article)
				}
			}
		}
	} else if filter.Author != "" {
		for _, articleAuthor := range ar.articles {
			for _, article := range articleAuthor {
				if filter.Author == article.Author.Username {
					articles = append(articles, article)
				}
			}
		}
	} else {
		for _, articleAuthor := range ar.articles {
			for _, article := range articleAuthor {
				if slices.Contains(article.TagList, filter.Tag) {
					articles = append(articles, article)
				}
			}
		}
	}
	return articles, nil
}

func (ar *ArticleRepository) Get() ([]*models.Article, error) {
	ar.mu.RLock()
	defer ar.mu.RUnlock()
	log.Printf("ArticleRepository: Get: articles: %#v\n", ar.articles)
	articles := make([]*models.Article, 0)
	for _, articleAuthor := range ar.articles {
		for _, article := range articleAuthor {
			articles = append(articles, article)
		}
	}
	return articles, nil
}
