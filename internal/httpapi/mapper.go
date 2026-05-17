package httpapi

import (
	"time"

	"realworld/internal/domain"
)

func fromDomainTags(tags []domain.Tag) []string {
	tagsStr := make([]string, len(tags))

	for i, tag := range tags {
		tagsStr[i] = string(tag)
	}

	return tagsStr
}

func fromDomainArticle(art *domain.Article) Article {
	return Article{
		Slug:           art.Slug,
		Title:          art.Title,
		Description:    art.Description,
		Body:           art.Body,
		TagList:        fromDomainTags(art.TagList),
		CreatedAt:      art.CreatedAt,
		UpdatedAt:      art.UpdatedAt,
		Favorited:      art.Favorited,
		FavoritesCount: art.FavoritesCount,
		Author:         FromDomainProfile(&art.Author),
	}
}

func fromDomainArticleToArticleListItem(art *domain.Article) ArticleListItem {
	return ArticleListItem{
		Author:         FromDomainProfile(&art.Author),
		CreatedAt:      art.CreatedAt,
		Description:    art.Description,
		Favorited:      art.Favorited,
		FavoritesCount: art.FavoritesCount,
		Slug:           art.Slug,
		TagList:        fromDomainTags(art.TagList),
		Title:          art.Title,
		UpdatedAt:      art.UpdatedAt,
	}
}

// ArticleListItem matches the anonymous struct type in MultipleArticlesResponseJSONResponse.Articles
type ArticleListItem = struct {
	Author         Profile   `json:"author"`
	CreatedAt      time.Time `json:"createdAt"`
	Description    string    `json:"description"`
	Favorited      bool      `json:"favorited"`
	FavoritesCount int       `json:"favoritesCount"`
	Slug           string    `json:"slug"`
	TagList        []string  `json:"tagList"`
	Title          string    `json:"title"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func fromDomainArticles(articles []*domain.Article) []ArticleListItem {
	articlesAPI := make([]ArticleListItem, len(articles))

	for i, a := range articles {
		articlesAPI[i] = fromDomainArticleToArticleListItem(a)
	}

	return articlesAPI
}

func FromDomainProfile(p *domain.Profile) Profile {
	return Profile{
		Username:  p.Username,
		Bio:       p.Bio,
		Image:     p.Image,
		Following: p.Following,
	}
}

func fromDomainComment(cmt *domain.Comment) Comment {
	return Comment{
		Id:        cmt.ID,
		CreatedAt: cmt.CreatedAt,
		UpdatedAt: cmt.UpdatedAt,
		Body:      cmt.Body,
		Author:    FromDomainProfile(&cmt.Author),
	}
}

func fromDomainComments(comments []*domain.Comment) []Comment {
	commentsAPI := make([]Comment, len(comments))

	for i, c := range comments {
		commentsAPI[i] = fromDomainComment(c)
	}

	return commentsAPI
}

func fromDomainUser(user *domain.User, token string) User {
	return User{
		Email:    user.Email,
		Token:    token,
		Username: user.Username,
		Bio:      user.Bio,
		Image:    user.Image,
	}
}
