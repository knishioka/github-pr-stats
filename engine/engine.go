package engine

import (
	"fmt"
	"log"
	"time"

	"github.com/knishioka/github-pr-stats/exporter"
	"github.com/knishioka/github-pr-stats/gitutil"
	"github.com/knishioka/github-pr-stats/models"
	"github.com/knishioka/github-pr-stats/token"
)

//Engine gets token from token agent, gets the required data
//From github through git helper and exports using exporter
type Engine struct {
	Getter     gitutil.GitHelper
	Exporter   exporter.ExportInterface
	TokenAgent token.InsTokenInterface
	Start      time.Time
	End        time.Time
	//Base defines #days before the startDate
	//Before which the system should ignore All the PRs
	Base int
}

//Run starts the engine
func (e *Engine) Run() error {
	base := e.Start.AddDate(0, 0, e.Base)
	e.Getter.SetBase(base)
	err := e.TokenAgent.GenerateNew()
	if err != nil {
		log.Fatalf("error getting installation token: %v", err.Error())
	}

	log.Println("gettingr org members")
	users, err := e.Getter.GetOrgMembers(e.TokenAgent)
	if err != nil {
		log.Fatalf("error getting org members: %v", err.Error())
	}

	log.Println("gettingr org repos")
	repos, err := e.Getter.GetOrgRepos(e.TokenAgent)
	if err != nil {
		log.Fatalf("error getting org repos: %v", err.Error())
	}

	log.Printf("repos found: %v\n", len(repos))
	log.Println("getting pull requests")
	prs, err := e.Getter.GetPullRequests(repos, e.TokenAgent)
	if err != nil {
		log.Fatalf("error getting repos pull requests: %v", err.Error())
	}

	log.Println("generating stats")
	stats := e.getStats(prs, users)

	log.Println("exporting stats")
	dateformat := "2006-01-02"
	filename := fmt.Sprintf("results_%v_to_%v.csv", e.Start.Format(dateformat), e.End.Format(dateformat))
	if err := e.Exporter.Export(stats, filename); err != nil {
		log.Fatalf("error exporting results: %v", err.Error())
	}

	log.Printf("stats exported to %v", filename)

	return nil
}

func (e *Engine) getStats(prs []*models.PullRequest, users []*models.User) map[string]*models.User {
	stats := make(map[string]*models.User)
	for i := 0; i < len(prs); i++ {
		for j := 0; j < len(prs[i].Reviews); j++ {
			if !e.Start.Before(prs[i].Reviews[j].SubmittedAt) {
				continue
			}

			if e.End.Before(prs[i].Reviews[j].SubmittedAt) {
				continue
			}

			if stats[prs[i].Reviews[j].Username] == nil {
				stats[prs[i].Reviews[j].Username] = &models.User{}
			}

			stats[prs[i].Reviews[j].Username].Username = prs[i].Reviews[j].Username
			stats[prs[i].Reviews[j].Username].ID = prs[i].Reviews[j].ID
			stats[prs[i].Reviews[j].Username].PullReqsReviewed++
		}

		if !e.Start.Before(prs[i].CreatedAt) {
			continue
		}

		if e.End.Before(prs[i].CreatedAt) {
			continue
		}

		if stats[prs[i].Username] == nil {
			stats[prs[i].Username] = &models.User{}
		}

		stats[prs[i].Username].PullReqsCreated++
		stats[prs[i].Username].TotalAdditions += prs[i].Additions
		stats[prs[i].Username].TotalDeletions += prs[i].Deletions
		stats[prs[i].Username].TotalChangedFiles += prs[i].ChangedFiles
		stats[prs[i].Username].TotalCommits += prs[i].Commits
		stats[prs[i].Username].ReviewsOnPullReqs += len(prs[i].Reviews)
		stats[prs[i].Username].Username = prs[i].Username
		stats[prs[i].Username].ID = prs[i].UserID
	}

	// Add the users which didn't create any PR
	// Nor did they gave any review
	for j := 0; j < len(users); j++ {
		if stats[users[j].Username] == nil {
			stats[users[j].Username] = &models.User{
				ID:       users[j].ID,
				Username: users[j].Username,
			}
		}
	}

	return stats
}
