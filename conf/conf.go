package conf

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/subosito/gotenv"
)

// Configuration specifies env variables
type Configuration struct {
	AppID          string
	GithubKey      string
	AccountName    string
	StartDate      string
	EndDate        string
	InstallationID int64
	// Base is #days before the StartDate before which we
	// want to ignore the PRs
	Base int
}

var (
	// Configs can be used gloablly to get env variables
	Configs *Configuration
)

// InitConfigs loads enviornment variables
func InitConfigs() {
	if err := gotenv.Load(); err != nil {
		log.Fatalf("gotenv: could not find .env file - Error: %v\n", err)
	}

	Configs = &Configuration{
		AppID:       os.Getenv("GITHUB_APP_ID"),
		GithubKey:   os.Getenv("GITHUB_APP_PRIVATE_KEY"),
		AccountName: os.Getenv("ACCOUNT_NAME"),
		StartDate:   os.Getenv("START_DATE"),
		EndDate:     os.Getenv("END_DATE"),
	}

	insID, err := strconv.Atoi(os.Getenv("INSTALLATION_ID"))
	if err != nil {
		log.Fatalf("invalid installation id : %v", os.Getenv("INSTALLATION_ID"))
	}
	Configs.InstallationID = int64(insID)

	baseStr := strings.TrimSpace(os.Getenv("BASE"))
	if baseStr != "" {
		base, err := strconv.Atoi(baseStr)
		if err != nil {
			log.Fatalf("invalid variable, Base : %v", baseStr)
		}

		if base > 0 {
			base = -base
		}
		Configs.Base = base
	} else {
		// assign default value: 30 days.
		Configs.Base = -30
	}
}
