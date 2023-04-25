package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.27

import (
	"context"
	"fmt"

	"github.com/zenith110/CMS-Backend/graph/model"
	"github.com/zenith110/CMS-Backend/graph/routes"
)

// CreateArticle is the resolver for the createArticle field.
func (r *mutationResolver) CreateArticle(ctx context.Context, input *model.CreateArticleInfo) (*model.Article, error) {
	article, err := routes.CreateArticle(input)
	return article, err
}

// UpdateArticle is the resolver for the updateArticle field.
func (r *mutationResolver) UpdateArticle(ctx context.Context, input *model.UpdatedArticleInfo) (*model.Article, error) {
	article, err := routes.UpdateArticle(input)
	return article, err
}

// DeleteArticle is the resolver for the deleteArticle field.
func (r *mutationResolver) DeleteArticle(ctx context.Context, input *model.DeleteBucketInfo) (string, error) {
	_, err := routes.DeleteArticle(input)
	return "", err
}

// DeleteAllArticles is the resolver for the deleteAllArticles field.
func (r *mutationResolver) DeleteAllArticles(ctx context.Context, input *model.DeleteAllArticlesInput) (string, error) {
	_, err := routes.DeleteArticles(input)
	return "", err
}

// CreateProject is the resolver for the createProject field.
func (r *mutationResolver) CreateProject(ctx context.Context, input *model.CreateProjectInput) (*model.Project, error) {
	project, err := routes.CreateProject(input)
	return project, err
}

// CreateUser is the resolver for the createUser field.
func (r *mutationResolver) CreateUser(ctx context.Context, input *model.UserCreation) (*model.User, error) {
	user, err := routes.CreateUser(input)
	return user, err
}

// LoginUser is the resolver for the loginUser field.
func (r *mutationResolver) LoginUser(ctx context.Context, username string, password string) (*model.LoginData, error) {
	jwt, err := routes.Login(username, password)
	return jwt, err
}

// DeleteProject is the resolver for the deleteProject field.
func (r *mutationResolver) DeleteProject(ctx context.Context, input *model.DeleteProjectType) (string, error) {
	message, err := routes.DeleteProject(input)
	return message, err
}

// DeleteProjects is the resolver for the deleteProjects field.
func (r *mutationResolver) DeleteProjects(ctx context.Context, input *model.DeleteAllProjects) (string, error) {
	result, err := routes.DeleteProjects(input)
	return result, err
}

// Logout is the resolver for the logout field.
func (r *mutationResolver) Logout(ctx context.Context, jwt string) (string, error) {
	result, err := routes.Logout(jwt)
	return result, err
}

// DeleteUser is the resolver for the deleteUser field.
func (r *mutationResolver) DeleteUser(ctx context.Context, input *model.DeleteUser) (string, error) {
	message, err := routes.DeleteUser(input)
	return message, err
}

// DeleteAllUsers is the resolver for the deleteAllUsers field.
func (r *mutationResolver) DeleteAllUsers(ctx context.Context, jwt string) (string, error) {
	results, err := routes.DeleteAllUsers(jwt)
	return results, err
}

// EditUser is the resolver for the editUser field.
func (r *mutationResolver) EditUser(ctx context.Context, input *model.EditUser) (string, error) {
	panic(fmt.Errorf("not implemented: EditUser - editUser"))
}

// UploadArticleImage is the resolver for the uploadArticleImage field.
func (r *mutationResolver) UploadArticleImage(ctx context.Context, input *model.UploadArticleImageInput) (string, error) {
	articleImageURL, err := routes.UploadArticleImages(input)
	return articleImageURL, err
}

// ArticlePrivate is the resolver for the articlePrivate field.
func (r *queryResolver) ArticlePrivate(ctx context.Context, input *model.FindArticlePrivateType) (*model.Article, error) {
	article, err := routes.FindArticle(input)
	return article, err
}

// ArticlesPrivate is the resolver for the articlesPrivate field.
func (r *queryResolver) ArticlesPrivate(ctx context.Context, input *model.ArticlesPrivate) (*model.Articles, error) {
	articles, err := routes.FetchArticles(input)
	return articles, err
}

// ArticlesPublic is the resolver for the articlesPublic field.
func (r *queryResolver) ArticlesPublic(ctx context.Context, input *model.GetZincArticleInput) (*model.Articles, error) {
	results, err := routes.FetchArticlesZinc(input)
	return results, err
}

// GetGalleryImages is the resolver for the getGalleryImages field.
func (r *queryResolver) GetGalleryImages(ctx context.Context, jwt string) (*model.GalleryImages, error) {
	images, err := routes.GalleryFindImages(jwt)
	return images, err
}

// GetProjects is the resolver for the getProjects field.
func (r *queryResolver) GetProjects(ctx context.Context, input *model.GetProjectType) (*model.Projects, error) {
	projects, err := routes.GetProjects(input)
	return projects, err
}

// ArticlePublic is the resolver for the articlePublic field.
func (r *queryResolver) ArticlePublic(ctx context.Context, input *model.FindArticlePublicType) (*model.Article, error) {
	article, err := routes.FindArticlePublic(input)
	return article, err
}

// GetUsers is the resolver for the getUsers field.
func (r *queryResolver) GetUsers(ctx context.Context, jwt string) (*model.Users, error) {
	users, err := routes.FetchUsers(jwt)
	return users, err
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//   - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//     it when you're done.
//   - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *mutationResolver) EditUserCheck(ctx context.Context, input *model.EditUser) (string, error) {
	panic(fmt.Errorf("not implemented: EditUserCheck - editUserCheck"))
}
