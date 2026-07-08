package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/andrewwillette/andrewwillettedotcom/aws"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	showTitleFlag       string
	showDateFlag        string
	showTimeFlag        string
	showDescriptionFlag string
)

var uploadShowCmd = &cobra.Command{
	Use:   "upload-show",
	Short: "Upload a show (title + description) to S3",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runUploadShow(); err != nil {
			log.Fatal().Err(err).Msg("upload-show failed")
		}
	},
}

func init() {
	uploadShowCmd.Flags().StringVarP(&showTitleFlag, "title", "t", "", "Show title")
	uploadShowCmd.Flags().StringVarP(&showDateFlag, "date", "D", "", "Show date (YYYY-MM-DD)")
	uploadShowCmd.Flags().StringVarP(&showTimeFlag, "time", "T", "", "Show time (e.g. 8:00pm-10:00pm)")
	uploadShowCmd.Flags().StringVarP(&showDescriptionFlag, "description", "d", "", "Show description")
	rootCmd.AddCommand(uploadShowCmd)
}

func runUploadShow() error {
	reader := bufio.NewReader(os.Stdin)

	title := strings.TrimSpace(showTitleFlag)
	if title == "" {
		var err error
		title, err = prompt(reader, "Show title: ")
		if err != nil {
			return err
		}
		title = strings.TrimSpace(title)
	}
	if title == "" {
		return fmt.Errorf("title is required")
	}

	date := strings.TrimSpace(showDateFlag)
	if date == "" {
		var err error
		date, err = prompt(reader, "Show date (YYYY-MM-DD): ")
		if err != nil {
			return err
		}
		date = strings.TrimSpace(date)
	}

	showTime := strings.TrimSpace(showTimeFlag)
	if showTime == "" {
		var err error
		showTime, err = prompt(reader, "Show time (e.g. 8:00pm-10:00pm): ")
		if err != nil {
			return err
		}
		showTime = strings.TrimSpace(showTime)
	}

	description := strings.TrimSpace(showDescriptionFlag)
	if description == "" {
		var err error
		description, err = prompt(reader, "Show description: ")
		if err != nil {
			return err
		}
		description = strings.TrimSpace(description)
	}

	log.Info().Msgf("Uploading show: title=%q date=%q time=%q", title, date, showTime)
	if err := aws.PutShowJSON(title, date, showTime, description); err != nil {
		return err
	}
	log.Info().Msg("Upload complete")
	return nil
}
