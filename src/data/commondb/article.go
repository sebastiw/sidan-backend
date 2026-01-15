package commondb

import (
	"github.com/sebastiw/sidan-backend/src/models"
)

func (d *CommonDatabase) CreateArticle(article *models.Article) (*models.Article, error) {
	result := d.DB.Create(article)
	if result.Error != nil {
		return nil, result.Error
	}
	return article, nil
}

func (d *CommonDatabase) ReadArticle(id int64) (*models.Article, error) {
	var article models.Article
	result := d.DB.First(&article, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &article, nil
}

func (d *CommonDatabase) ReadArticles(take int, skip int) ([]models.Article, error) {
	var articles []models.Article
	result := d.DB.Order("Id DESC").Limit(take).Offset(skip).Find(&articles)
	if result.Error != nil {
		return nil, result.Error
	}
	return articles, nil
}

func (d *CommonDatabase) UpdateArticle(article *models.Article) (*models.Article, error) {
	result := d.DB.Save(article)
	if result.Error != nil {
		return nil, result.Error
	}
	return article, nil
}

func (d *CommonDatabase) DeleteArticle(article *models.Article) (*models.Article, error) {
	result := d.DB.Delete(article)
	if result.Error != nil {
		return nil, result.Error
	}
	return article, nil
}
