package oapi

import (
	"realworld/internal/domain"
	"realworld/internal/oapi/oapigen"
)

func fromDomainTags(tags []domain.Tag) []string {
	tagsStr := make([]string, len(tags))

	for i, tag := range tags {
		tagsStr[i] = string(tag)
	}

	return tagsStr
}

func fromDomainArticle(art *domain.Article) oapigen.Article {
	return oapigen.Article{
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

func fromDomainArticles(articles []*domain.Article) []oapigen.Article {
	articlesAPI := make([]oapigen.Article, len(articles))

	for i, a := range articles {
		articlesAPI[i] = fromDomainArticle(a)
	}

	return articlesAPI
}

func FromDomainProfile(p *domain.Profile) oapigen.Profile {
	return oapigen.Profile{
		Username:  p.Username,
		Bio:       p.Bio,
		Image:     p.Image,
		Following: p.Following,
	}
}

func fromDomainComment(cmt *domain.Comment) oapigen.Comment {
	return oapigen.Comment{
		Id:        cmt.ID,
		CreatedAt: cmt.CreatedAt,
		UpdatedAt: cmt.UpdatedAt,
		Body:      cmt.Body,
		Author:    FromDomainProfile(&cmt.Author),
	}
}

func fromDomainComments(comments []*domain.Comment) []oapigen.Comment {
	commentsAPI := make([]oapigen.Comment, len(comments))

	for i, c := range comments {
		commentsAPI[i] = fromDomainComment(c)
	}

	return commentsAPI
}

func fromDomainUser(user *domain.User, token string) oapigen.User {
	return oapigen.User{
		Email:    user.Email,
		Token:    token,
		Username: user.Username,
		Bio:      user.Bio,
		Image:    user.Image,
	}
}
