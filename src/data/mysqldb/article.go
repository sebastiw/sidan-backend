package mysqldb

import (
	"github.com/sebastiw/sidan-backend/src/models"
)

func (d *MySQLDatabase) CreateArticle(article *models.Article) (*models.Article, error) {
	return d.CommonDB.CreateArticle(article)
}

func (d *MySQLDatabase) ReadArticle(id int64) (*models.Article, error) {
	return d.CommonDB.ReadArticle(id)
}

func (d *MySQLDatabase) ReadArticles(take int, skip int) ([]models.Article, error) {
	return d.CommonDB.ReadArticles(take, skip)
}

func (d *MySQLDatabase) UpdateArticle(article *models.Article) (*models.Article, error) {
	return d.CommonDB.UpdateArticle(article)
}

func (d *MySQLDatabase) DeleteArticle(article *models.Article) (*models.Article, error) {
	return d.CommonDB.DeleteArticle(article)
}
