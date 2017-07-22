package cmd

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/thesoenke/news-crawler/scraper"
)

var itemsInputFile string
var scrapeVerbose bool

var cmdScrape = &cobra.Command{
	Use:   "scrape",
	Short: "Scrape all provided articles",
	RunE: func(cmd *cobra.Command, args []string) error {

		location, err := time.LoadLocation(timezone)
		if err != nil {
			return err
		}

		path, err := getFeedsFilePath(location)
		if err != nil {
			return err
		}

		contentScraper, err := scraper.New(path)
		if err != nil {
			return err
		}

		elasticClient, err := scraper.NewElasticClient()
		if err != nil {
			return err
		}

		start := time.Now()
		err = contentScraper.Scrape(elasticClient, scrapeVerbose)
		if err != nil {
			return err
		}

		log.Printf("Articles: %d successful, %d failures in %s from %s", contentScraper.Articles-contentScraper.Failures, contentScraper.Failures, time.Since(start), path)

		return nil
	},
}

func init() {
	cmdScrape.PersistentFlags().StringVarP(&itemsInputFile, "file", "f", "out/feeds", "Path to a JSON file with feed items")
	cmdScrape.PersistentFlags().StringVarP(&timezone, "timezone", "t", "Europe/Berlin", "Timezone for storing the feeds")
	cmdScrape.PersistentFlags().BoolVarP(&scrapeVerbose, "verbose", "v", false, "Verbose logging of scraper")
	RootCmd.AddCommand(cmdScrape)
}

func getFeedsFilePath(location *time.Location) (string, error) {
	if itemsInputFile == "" {
		return "", errors.New("Please provide a file with articles")
	}

	stat, err := os.Stat(itemsInputFile)
	if err != nil {
		return "", err
	}

	// Append current day to path when only received directory as input location
	if stat.IsDir() {
		day := time.Now().In(location)
		dayStr := day.Format("2-1-2006")
		path := filepath.Join(itemsInputFile, dayStr+".json")
		_, err := os.Stat(path)
		if err == nil {
			return path, nil
		}

		// most recent file could be from 1 day before
		day = day.AddDate(0, 0, -1)
		dayStr = day.Format("2-1-2006")
		path = filepath.Join(itemsInputFile, dayStr+".json")

		_, err = os.Stat(path)
		if err != nil {
			return "", nil
		}

		return path, nil
	}

	return itemsInputFile, nil
}
