package models

import "time"

//User defines a github user
type User struct {
	ID                int64
	Username          string
	PullReqsCreated   int
	PullReqsReviewed  int
	ReviewsOnPullReqs int
	TotalAdditions    int
	TotalDeletions    int
	TotalChangedFiles int
	TotalCommits      int
}

//Repo defines a github repo
type Repo struct {
	ID       int64
	Name     string
	FullName string
}

//PullRequest defines a github pr
type PullRequest struct {
	ID           int64
	RepoID       int64
	RepoName     string
	UserID       int64
	Username     string
	PrNo         int
	Additions    int
	Deletions    int
	ChangedFiles int
	Commits      int
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Reviews      []*Review
}

//Review defines a review on github pr
type Review struct {
	ID          int64
	State       string
	UserID      int64
	Username    string
	SubmittedAt time.Time
}
