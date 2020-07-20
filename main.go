package main

import (
	"context"
	"log"
	"time"

	"github.com/knishioka/github-pr-stats/conf"
	"github.com/knishioka/github-pr-stats/engine"
	"github.com/knishioka/github-pr-stats/exporter"
	"github.com/knishioka/github-pr-stats/gitutil"
	"github.com/knishioka/github-pr-stats/token"
)

func main() {
	conf.InitConfigs()
	dateFormat := "2006-01-02"
	start, err := time.Parse(dateFormat, conf.Configs.StartDate)
	if err != nil {
		log.Fatal(err)
	}

	var end time.Time
	if conf.Configs.EndDate == "" {
		end = time.Now()
	} else {
		end, err = time.Parse(dateFormat, conf.Configs.EndDate)
		if err != nil {
			log.Fatal(err)
		}
	}

	end = end.AddDate(0, 0, 1)

	ctx := context.Background()
	gitClient := gitutil.NewGithubClient(ctx)
	exporter := exporter.NewExcelExporter()
	tokenAgent := token.NewInsTokenAgent(ctx, conf.Configs.InstallationID, conf.Configs.AccountName)

	engine := engine.Engine{
		Getter:     gitClient,
		Exporter:   exporter,
		TokenAgent: tokenAgent,
		Start:      start,
		End:        end,
		Base:       conf.Configs.Base,
	}

	if err := engine.Run(); err != nil {
		log.Fatal(err)
	}
}
