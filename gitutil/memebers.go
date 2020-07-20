package gitutil

import (
	"github.com/knishioka/github-pr-stats/models"
	"github.com/knishioka/github-pr-stats/token"
)

// GetOrgMembers calls github API and returns list of accounts that are members of an org
func (h *GithubClient) GetOrgMembers(ita token.InsTokenInterface) (accounts []*models.User, err error) {
	// Get org. memebrs
	users, err := h.GetAllUsers(h.getOrgMembersURL(ita.AccountName()), ita)
	if err != nil {
		return nil, err
	}

	for j := 0; j < len(users); j++ {
		user := &models.User{
			ID:       users[j].GetID(),
			Username: users[j].GetLogin(),
		}

		accounts = append(accounts, user)
	}

	return accounts, nil
}
