package gitutil

import (
	"github.com/knishioka/github-pr-stats/models"
	"github.com/knishioka/github-pr-stats/token"
)

// GetOrgRepos calls github API and returns list of repos that belong to org
func (h *GithubClient) GetOrgRepos(ita token.InsTokenInterface) (repos []*models.Repo, err error) {
	// Get org. memebrs
	repositories, err := h.GetAllRepos(h.getOrgReposURL(ita.AccountName()), ita)
	if err != nil {
		return nil, err
	}

	for j := 0; j < len(repositories); j++ {
		repo := &models.Repo{
			ID:   repositories[j].GetID(),
			Name: repositories[j].GetName(),
		}

		repos = append(repos, repo)
	}

	return repos, nil
}
