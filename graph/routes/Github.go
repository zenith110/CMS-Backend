package routes

import (
	"context"
	"fmt"
	"os"

	"github.com/shurcooL/githubv4"
	"github.com/zenith110/portfilo/graph/model"
	"golang.org/x/oauth2"
)

type Language struct {
	Language string
}
type Repository struct {
	Name        string
	Description string
	Url         string
	HomepageUrl string
}

/*
Given a language url, loop through the json contents and append it into a tag struct
*/
func ParseLanguages(languages []struct{ Name string }) []model.Tag {
	var tags []model.Tag
	for language := range languages {
		var tag model.Tag
		tag.Language = languages[language].Name
		tags = append(tags, tag)
	}
	return tags
}

func ParseTopics(topicsList []struct{ Topic struct{ Name string } }) []string {
	var topics []string
	for topicIndex := 0; topicIndex < len(topicsList); topicIndex++ {

		topics = append(topics, topicsList[topicIndex].Topic.Name)
	}

	return topics
}

/*
@Params - nil
@Description - Fetches github projects and returns the data in an array for the graphql endpoint to grab
*/
func FetchProjects(githubUser string) (*model.GithubProjects, error) {
	ctx := context.Background()
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUBACCESSTOKEN")},
	)
	httpClient := oauth2.NewClient(ctx, src)

	client := githubv4.NewClient(httpClient)
	var query struct {
		RepositoryOwner struct {
			Login        string
			Repositories struct {
				Nodes []struct {
					Name        string
					Description string
					Url         string
					HomepageUrl string
					Languages   struct {
						Nodes []struct {
							Name string
						}
					} `graphql:"languages(first: 100)"`
					RepositoryTopic struct {
						Nodes []struct {
							Topic struct {
								Name string
							}
						}
					} `graphql:"repositoryTopics(first: 100)"`
				}
			} `graphql:"repositories(privacy: PUBLIC, first: 100, ownerAffiliations: [OWNER])"`
		} `graphql:"repositoryOwner(login: $user)"`
	}
	variables := map[string]interface{}{
		"user": githubv4.String(githubUser),
	}
	queryErr := client.Query(ctx, &query, variables)
	if queryErr != nil {
		fmt.Printf("Error is %v", queryErr)
	}

	repos := query.RepositoryOwner.Repositories.Nodes
	var projects []*model.Project
	for repo := 0; repo < len(repos); repo++ {
		projects = append(projects, &model.Project{
			Name:           repos[repo].Name,
			Githublink:     repos[repo].Url,
			Description:    repos[repo].Description,
			Languages:      ParseLanguages(repos[repo].Languages.Nodes),
			Topics:         ParseTopics(repos[repo].RepositoryTopic.Nodes),
			Deploymentlink: repos[repo].HomepageUrl,
		})
	}

	var gitProjects = model.GithubProjects{Projects: projects}
	return &gitProjects, queryErr
}
