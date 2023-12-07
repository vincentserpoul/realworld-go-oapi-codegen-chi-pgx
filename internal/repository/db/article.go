package db

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"

	"realworld/internal/domain"
)

// implement the interface ArticleRepository
func (r *Repository) GetArticles(
	ctx context.Context,
	userID uuid.UUID,
	author, tag, favorited *string,
	limit, offset *int,
) (
	[]*domain.Article,
	error,
) {
	query, args := getArticleQuery(userID, nil, author, tag, favorited, false, limit, offset)

	rows, errR := r.pool.Query(ctx, query, args)
	if errR != nil {
		return nil, fmt.Errorf("could not get articles: %w", errR)
	}

	articles, errA := pgx.CollectRows(rows, JSONRowToAddrOfStruct[domain.Article])
	if errA != nil {
		return nil, fmt.Errorf("could not collect rows: %w", errA)
	}

	return articles, nil
}

func (r *Repository) GetArticle(
	ctx context.Context,
	userID uuid.UUID,
	artSlug string,
) (*domain.Article, error) {
	query, args := getArticleQuery(userID, &artSlug, nil, nil, nil, false, nil, nil)

	rows, errR := r.pool.Query(ctx, query, args)
	if errR != nil {
		return nil, fmt.Errorf("could not get article: %w", errR)
	}

	article, errA := pgx.CollectExactlyOneRow(rows, JSONRowToAddrOfStruct[domain.Article])
	if errA != nil {
		return nil, fmt.Errorf("could not collect rows: %w", errA)
	}

	return article, nil
}

func (r Repository) GetFeedArticles(
	ctx context.Context,
	userID uuid.UUID,
	author, tag, favorited *string,
	limit, offset *int,
) ([]*domain.Article, error) {
	query, args := getArticleQuery(userID, nil, author, tag, favorited, true, limit, offset)

	rows, errR := r.pool.Query(ctx, query, args)
	if errR != nil {
		return nil, fmt.Errorf("could not get articles: %w", errR)
	}

	articles, errA := pgx.CollectRows(rows, JSONRowToAddrOfStruct[domain.Article])
	if errA != nil {
		return nil, fmt.Errorf("could not collect rows: %w", errA)
	}

	return articles, nil
}

func (r *Repository) CreateArticle(
	ctx context.Context,
	userID uuid.UUID,
	title,
	description,
	body string,
	tagList []string,
) (*domain.Article, error) {
	// create the article
	if err := r.createArticle(ctx, userID, title, description, body, tagList); err != nil {
		return nil, fmt.Errorf("could not create article: %w", err)
	}

	// get the article
	article, err := r.GetArticle(ctx, userID, domain.GetSlugFromTitle(title))
	if err != nil {
		return nil, fmt.Errorf("could not get article after creation: %w", err)
	}

	return article, nil
}

func (r *Repository) createArticle(
	ctx context.Context,
	userID uuid.UUID,
	title,
	description,
	body string,
	tagList []string,
) error {
	articleID, errU := uuid.NewV7()
	if errU != nil {
		return fmt.Errorf("could not generate uuid: %w", errU)
	}

	articleSlug := domain.GetSlugFromTitle(title)

	tagList = removeDuplicates(tagList)

	tagIDs, tagNames, errT := getTagValues(tagList)
	if errT != nil {
		return fmt.Errorf("could not generate uuid: %w", errT)
	}

	var errTx error

	trx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("could not begin tx: %w", err)
	}

	defer func() {
		if errTx != nil {
			if err := trx.Rollback(ctx); err != nil {
				errTx = errors.Join(errTx, fmt.Errorf("could not rollback tx: %w", err))
			}

			return
		}

		if err := trx.Commit(ctx); err != nil {
			errTx = errors.Join(errTx, fmt.Errorf("could not commit tx: %w", err))
		}
	}()

	// Create a batch
	batch := &pgx.Batch{}

	// add article insert to the batch
	batch.Queue(`
		INSERT INTO article (id, slug, title, description, body, author_id)
		VALUES (@articleID, @slug, @title, @description, @body, @userID);`,
		pgx.NamedArgs{
			"articleID":   articleID,
			"slug":        articleSlug,
			"title":       title,
			"description": description,
			"body":        body,
			"userID":      userID,
		},
	)

	// add tag insert to the batch
	batch.Queue(`
		INSERT INTO tag (id, name)
		SELECT * FROM UNNEST(@tagIDs::uuid[], @tagNames::text[])
		ON CONFLICT (name) DO NOTHING`,
		pgx.NamedArgs{
			"tagIDs":   tagIDs,
			"tagNames": tagNames,
		},
	)

	// associate article with tags
	batch.Queue(`
		INSERT INTO article_tag (article_id, tag_id)
		SELECT @articleID, id FROM tag WHERE name = ANY(@tagNames)`,
		pgx.NamedArgs{
			"articleID": articleID,
			"tagNames":  tagNames,
		},
	)

	// execute the batch
	batchRes := trx.SendBatch(ctx, batch)

	// check for errors
	if _, err := batchRes.Exec(); err != nil {
		errTx = errors.Join(errTx, fmt.Errorf("could not exec batch: %w", err))

		return errTx
	}

	// close the batch
	if err := batchRes.Close(); err != nil {
		errTx = errors.Join(errTx, fmt.Errorf("could not close batch: %w", err))

		return errTx
	}

	tags := make([]domain.Tag, len(tagList))
	for idx, tag := range tagList {
		tags[idx] = domain.Tag(tag)
	}

	return errTx
}

func getTagValues(tags []string) ([]uuid.UUID, []string, error) {
	ids := make([]uuid.UUID, len(tags))
	names := make([]string, len(tags))

	for idx, tag := range tags {
		tagID, errU := uuid.NewV7()
		if errU != nil {
			return nil, nil, fmt.Errorf("could not generate uuid: %w", errU)
		}

		ids[idx] = tagID
		names[idx] = tag
	}

	return ids, names, nil
}

func removeDuplicates(strs []string) []string {
	seen := make(map[string]struct{}, len(strs))
	jIdx := 0

	for _, str := range strs {
		if _, ok := seen[str]; ok {
			continue
		}

		seen[str] = struct{}{}
		strs[jIdx] = str
		jIdx++
	}

	return strs[:jIdx]
}

func (r *Repository) UpdateArticle(
	ctx context.Context,
	userID uuid.UUID,
	artSlug string,
	title, description, body *string,
) (*domain.Article, error) {
	updateParams := pgx.NamedArgs{
		"slug":   artSlug,
		"userID": userID,
	}

	updateFields := []string{}

	if title != nil {
		updateFields = append(updateFields, `title = @title`)
		updateParams["title"] = title
	}

	if description != nil {
		updateFields = append(updateFields, `description = @description`)
		updateParams["description"] = description
	}

	if body != nil {
		updateFields = append(updateFields, `body = @body`)
		updateParams["body"] = body
	}

	sql := fmt.Sprintf(`
		UPDATE article
		SET %s
		WHERE slug = @slug
		AND author_id = @userID`,
		strings.Join(updateFields, ", "),
	)

	_, err := r.pool.Exec(ctx, sql, updateParams)
	if err != nil {
		return nil, fmt.Errorf("could not update article: %w", err)
	}

	return r.GetArticle(ctx, userID, artSlug)
}

func (r *Repository) DeleteArticle(ctx context.Context, userID uuid.UUID, artSlug string) error {
	sql := `DELETE FROM article WHERE slug = @slug AND author_id = @userID`

	_, err := r.pool.Exec(ctx, sql, pgx.NamedArgs{"slug": artSlug, "userID": userID})
	if err != nil {
		return fmt.Errorf("could not delete article: %w", err)
	}

	return nil
}

func (r *Repository) FavoriteArticle(
	ctx context.Context,
	userID uuid.UUID,
	artSlug string,
) (*domain.Article, error) {
	sql := `
		INSERT INTO article_favorite (article_id, appuser_id) VALUES
		((SELECT id FROM article WHERE slug = @slug), @userID)
	`

	_, err := r.pool.Exec(ctx, sql, pgx.NamedArgs{"slug": artSlug, "userID": userID})
	if err != nil {
		return nil, fmt.Errorf("could not favorite article: %w", err)
	}

	return r.GetArticle(ctx, userID, artSlug)
}

func (r *Repository) UnfavoriteArticle(
	ctx context.Context,
	userID uuid.UUID,
	artSlug string,
) (*domain.Article, error) {
	sql := `
		DELETE FROM article_favorite
		WHERE article_id = (SELECT id FROM article WHERE slug = @slug)
		AND appuser_id = @userID
	`

	_, err := r.pool.Exec(ctx, sql, pgx.NamedArgs{"slug": artSlug, "userID": userID})
	if err != nil {
		return nil, fmt.Errorf("could not unfavorite article: %w", err)
	}

	return r.GetArticle(ctx, userID, artSlug)
}

func getArticleQuery(
	userID uuid.UUID,
	artSlug, author, tag, favorited *string,
	feed bool,
	limit, offset *int,
) (string, pgx.NamedArgs) {
	query := `
		SELECT
			JSON_BUILD_OBJECT(
				'id', a.id,
				'slug', a.slug,
				'title', a.title,
				'description', a.description,
				'body', a.body,
				'tag_list', (
					SELECT JSON_AGG(name ORDER BY name)
					FROM article_tag
					JOIN tag ON article_tag.tag_id = tag.id
					WHERE article_id = a.id
				),
				'created_at', a.created_at,
				'updated_at', a.updated_at,
				'favorited', EXISTS(
					SELECT 1
					FROM article_favorite
					WHERE article_id = a.id
					AND appuser_id = @userID
				),
				'favorites_count', (
						SELECT COUNT(*)
						FROM article_favorite
						WHERE article_id = a.id
				),
				'author', (
					SELECT JSON_BUILD_OBJECT(
						'username', u.username,
						'bio', u.bio,
						'img', u.img,
						'following', EXISTS(
							SELECT 1
							FROM appuser_follows
							WHERE follower_id = @userID
							AND followee_id = u.id
						)
					)
					FROM appuser u
					WHERE u.id = a.author_id
				)
			)
		FROM article a
	`

	queryFilter := []string{}

	queryArgs := pgx.NamedArgs{
		"userID": userID,
	}

	// filters
	if artSlug != nil {
		queryFilter = append(queryFilter, "a.slug = @slug")
		queryArgs["slug"] = artSlug
	}

	if tag != nil {
		queryFilter = append(
			queryFilter,
			`EXISTS(
				SELECT 1
				FROM article_tag
				JOIN tag ON article_tag.tag_id = tag.id
				WHERE article_id = a.id AND tag.name = @tag
			)`,
		)
		queryArgs["tag"] = tag
	}

	if author != nil {
		queryFilter = append(
			queryFilter,
			`EXISTS(
				SELECT 1
				FROM appuser
				WHERE username = @author
				AND id = a.author_id
			)`,
		)
		queryArgs["author"] = author
	}

	if favorited != nil {
		queryFilter = append(
			queryFilter,
			`EXISTS(
				SELECT 1
				FROM article_favorite af
				JOIN appuser au ON af.appuser_id = au.id
				WHERE au.username = @favoritedUsername
			)`,
		)

		queryArgs["favoritedUsername"] = favorited
	}

	if feed {
		queryFilter = append(
			queryFilter,
			`EXISTS(
				SELECT 1
				FROM appuser_follows
				WHERE follower_id = @userID
				AND followee_id = a.author_id
			)`,
		)
	}

	// filter
	if len(queryFilter) > 0 {
		query += "\nWHERE " + strings.Join(queryFilter, "\n AND ")
	}

	// pagination
	if limit != nil {
		query += " LIMIT @limit"
		queryArgs["limit"] = limit

		if offset != nil {
			query += " OFFSET @offset"
			queryArgs["offset"] = offset
		}
	}

	return query, queryArgs
}
