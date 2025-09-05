package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "andrewwwillettedotcom",
	Short: "CLI for andrewwillette.com",
	Long:  `CLI for managing andrewwillette.com server and uploading media`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
