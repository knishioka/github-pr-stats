package gitutil

import (
	"encoding/json"
	"fmt"

	"github.com/google/go-github/github"
	"github.com/knishioka/github-pr-stats/models"
	"github.com/knishioka/github-pr-stats/token"
)

// GetPullRequests calls github API and returns pul reqs for each repo
func (h *GithubClient) GetPullRequests(repos []*models.Repo, ita token.InsTokenInterface) (pullReqs []*models.PullRequest, err error) {
	for i := 0; i < len(repos); i++ {
		// Get all pull requests for the repo
		prs, err := h.GetAllPullRequests(h.getRepoPrsURL(ita.AccountName(), repos[i].Name), ita)
		if err != nil {
			return nil, err
		}

		// for each PR, get all of its reviews
		for j := 0; j < len(prs); j++ {
			detailData, err := h.Get(h.getPrDetailURL(ita.AccountName(), repos[i].Name, prs[j].GetNumber()), ita)
			if err != nil {
				return nil, err
			}

			pullReqDetail := &github.PullRequest{}
			if err := json.Unmarshal(detailData, &pullReqDetail); err != nil {
				return nil, fmt.Errorf("pull request detail unmarshal error: %v ", err)
			}

			pr := &models.PullRequest{
				ID:           prs[j].GetID(),
				RepoID:       repos[i].ID,
				RepoName:     repos[i].Name,
				UserID:       prs[j].User.GetID(),
				Username:     prs[j].User.GetLogin(),
				PrNo:         prs[j].GetNumber(),
				Additions:    pullReqDetail.GetAdditions(),
				Deletions:    pullReqDetail.GetDeletions(),
				ChangedFiles: pullReqDetail.GetChangedFiles(),
				CreatedAt:    pullReqDetail.GetCreatedAt(),
				UpdatedAt:    pullReqDetail.GetUpdatedAt(),
				Commits:      pullReqDetail.GetCommits(),
				Reviews:      []*models.Review{},
			}

			// get all reviews of the PR
			revs, err := h.GetAllReviews(h.getPrReviewsURL(ita.AccountName(), repos[i].Name, pr.PrNo), ita)
			if err != nil {
				return nil, err
			}

			// get the needed values and store
			for k := 0; k < len(revs); k++ {
				pr.Reviews = append(pr.Reviews, &models.Review{
					ID:          revs[k].GetID(),
					State:       revs[k].GetState(),
					SubmittedAt: revs[k].GetSubmittedAt(),
					UserID:      revs[k].User.GetID(),
					Username:    revs[k].User.GetLogin(),
				})
			}

			// append PR to the final list to be returned
			pullReqs = append(pullReqs, pr)
		}
	}

	return pullReqs, nil
}
