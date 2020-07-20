package exporter

import (
	"fmt"
	"os"
	"strconv"

	"encoding/csv"

	"github.com/knishioka/github-pr-stats/models"
)

//ExportInterface defines framework for an exporter
type ExportInterface interface {
	Export(map[string]*models.User, string) error
}

type excelExporter struct{}

//NewExcelExporter returns excelExporter instance as ExportInterface
func NewExcelExporter() ExportInterface {
	return &excelExporter{}
}
func (exp *excelExporter) Export(stats map[string]*models.User, filename string) error {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err.Error())
	}

	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// write header row
	record := []string{"username", "Pull Requests Created", "Pull Requests Reviewed",
		"Reviews on Pull Requests", "Additions", "Deletions", "Files Changed", "Total Commits"}

	if err = writer.Write(record); err != nil {
		return fmt.Errorf("error writing to file: %v", err.Error())
	}

	// write stats
	for _, user := range stats {
		record := []string{user.Username, strconv.Itoa(user.PullReqsCreated), strconv.Itoa(user.PullReqsReviewed),
			strconv.Itoa(user.ReviewsOnPullReqs), strconv.Itoa(user.TotalAdditions),
			strconv.Itoa(user.TotalDeletions), strconv.Itoa(user.TotalChangedFiles),
			strconv.Itoa(user.TotalCommits),
		}

		err := writer.Write(record)
		if err != nil {
			return fmt.Errorf("error writing to file: %v", err.Error())
		}
	}

	return nil
}
