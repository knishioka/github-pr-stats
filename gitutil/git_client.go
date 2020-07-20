package gitutil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/google/go-github/github"
	"github.com/knishioka/github-pr-stats/models"
	"github.com/knishioka/github-pr-stats/token"
)

// GitHelper represents Github API Helper
type GitHelper interface {
	GetOrgRepos(token.InsTokenInterface) ([]*models.Repo, error)
	GetOrgMembers(token.InsTokenInterface) ([]*models.User, error)
	GetPullRequests([]*models.Repo, token.InsTokenInterface) ([]*models.PullRequest, error)
	SetBase(time.Time)
}

// GithubClient implements GitHelper
type GithubClient struct {
	c    *http.Client
	base time.Time
}

// NewGithubClient returns a GitHelper
func NewGithubClient(ctx context.Context) GitHelper {
	return &GithubClient{
		c: &http.Client{
			Timeout: time.Second * 23,
		},
	}
}

//Github API Docs: https://developer.github.com/v3/repos/#list-organization-repositories
func (h *GithubClient) getOrgReposURL(orgName string) string {
	return fmt.Sprintf("https://api.github.com/orgs/%v/repos?per_page=100", orgName)
}

//Github API Docs: https://developer.github.com/v3/orgs/members/
func (h *GithubClient) getOrgMembersURL(orgName string) string {
	return fmt.Sprintf("https://api.github.com/orgs/%v/members?per_page=100", orgName)
}

//Github API Docs: https://developer.github.com/v3/pulls/#list-pull-requests
func (h *GithubClient) getRepoPrsURL(orgName, repoName string) string {
	return fmt.Sprintf("https://api.github.com/repos/%v/%v/pulls?state=all&per_page=20", orgName, repoName)
}

//Github API Docs:https://developer.github.com/v3/pulls/reviews/#list-reviews-for-a-pull-request
func (h *GithubClient) getPrReviewsURL(orgName string, repoName string, prNo int) string {
	return fmt.Sprintf("https://api.github.com/repos/%v/%v/pulls/%v/reviews?per_page=100", orgName, repoName, prNo)
}

//Github API Docs:https://developer.github.com/v3/pulls/#get-a-pull-request
func (h *GithubClient) getPrDetailURL(orgName string, repoName string, prNo int) string {
	return fmt.Sprintf("https://api.github.com/repos/%v/%v/pulls/%v", orgName, repoName, prNo)
}

//SetBase sets the base date
func (h *GithubClient) SetBase(base time.Time) {
	h.base = base
}

func paginate(uri string, i int) string {
	return fmt.Sprintf("%v&page=%v", uri, i)
}

//GetAllUsers traverse through the API pagination & returns all the users
func (h *GithubClient) GetAllUsers(uri string, ita token.InsTokenInterface) (users []*github.User, err error) {
	i := 1
	for {
		var members []*github.User
		body, err := h.Get(paginate(uri, i), ita)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(body, &members); err != nil {
			return nil, err
		}

		users = append(users, members...)

		if len(members) < 99 {
			break
		}

		i++
	}

	return users, nil
}

//GetAllPullRequests traverse through the API pagination & returns all the Pull Requests
func (h *GithubClient) GetAllPullRequests(uri string, ita token.InsTokenInterface) (pullReqs []*github.PullRequest, err error) {
	i := 1
	for {
		var prs []*github.PullRequest
		body, err := h.Get(paginate(uri, i), ita)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(body, &prs); err != nil {
			return nil, err
		}

		pullReqs = append(pullReqs, prs...)

		if len(prs) < 19 {
			break
		}

		createdAt := prs[len(prs)-1].CreatedAt
		if createdAt != nil {
			if !h.base.Before(*createdAt) {
				break
			}
		}

		i++
	}

	return pullReqs, nil
}

//GetAllReviews traverse through the API pagination & returns all the Reviews on a Pull Request
func (h *GithubClient) GetAllReviews(uri string, ita token.InsTokenInterface) (reviews []*github.PullRequestReview, err error) {
	i := 1
	for {
		var revs []*github.PullRequestReview
		body, err := h.Get(paginate(uri, i), ita)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(body, &revs); err != nil {
			return nil, err
		}

		reviews = append(reviews, revs...)

		if len(revs) < 99 {
			break
		}

		i++
	}

	return reviews, nil
}

//GetAllRepos traverse through the API pagination & returns all the repos
func (h *GithubClient) GetAllRepos(uri string, ita token.InsTokenInterface) (repos []*github.Repository, err error) {
	i := 1
	for {
		var rep []*github.Repository
		body, err := h.Get(paginate(uri, i), ita)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(body, &rep); err != nil {
			return nil, err
		}

		repos = append(repos, rep...)
		if len(rep) < 99 {
			break
		}

		i++
	}

	return repos, nil
}

// Get returns bytes given a URL
func (h *GithubClient) Get(uri string, ita token.InsTokenInterface) (body []byte, err error) {
	req, err := http.NewRequest("GET", uri, &bytes.Buffer{})
	if err != nil {
		return body, fmt.Errorf("create new HTTP request: %v: %v", uri, err.Error())
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", ita.Bearer()))
	req.Header.Add("Accept", "application/vnd.github.machine-man-preview+json")

	data, err := h.c.Do(req)
	if err != nil {
		return body, fmt.Errorf("make request error:%v: %v", uri, err.Error())
	}
	defer data.Body.Close()

	if data.StatusCode != 200 {
		if data.StatusCode == 401 {
			err := ita.GenerateNew()
			if err != nil {
				return body, fmt.Errorf("get installation token: %v", err.Error())
			}

			req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", ita.Bearer()))
			data, err = h.c.Do(req)
			if err != nil {
				return body, fmt.Errorf("retry with fresh token make request error:%v: %v", uri, err.Error())
			}
			if data.StatusCode != 200 {
				return body, fmt.Errorf("retry with fresh token make request error: unexpected response status %v", data.StatusCode)
			}
		} else {
			return body, fmt.Errorf("make request : %v error: unexpected response status %v", uri, data.StatusCode)
		}
	}

	body, err = ioutil.ReadAll(data.Body)
	if err != nil {
		return body, fmt.Errorf("read body error: %v", err.Error())
	}

	return body, nil
}
