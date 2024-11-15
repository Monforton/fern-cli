package fern

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	projectName      string
	reportsDirectory string
	fernApiUrl       string
)

var rootCmd = &cobra.Command{
	Use:   "fern",
	Short: "Fern reporter",
	Long:  `Fern reporter cli tool`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print("Please choose a test format; see fern --help for more\n\n")

		fmt.Printf("Project Name: %s\n", projectName)
		fmt.Printf("Test Reports Directory: %s\n", reportsDirectory)
		fmt.Printf("Fern API Url: %s\n", fernApiUrl)
	},
}

func init() {
	// Define flags for the command
	rootCmd.PersistentFlags().StringVarP(&projectName, "projectName", "n", "", "Name of the project (required)")
	rootCmd.PersistentFlags().StringVarP(&reportsDirectory, "reportDirectory", "d", "", "Path to the test reports directory (required)")
	rootCmd.PersistentFlags().StringVarP(&fernApiUrl, "fernApiUrl", "u", "", "Fern API url to send reports (required)")

	// Mark flags as required
	rootCmd.MarkFlagRequired("projectName")
	rootCmd.MarkFlagRequired("reportDirectory")
	rootCmd.MarkFlagRequired("fernApiUrl")

}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing the CLI '%s'", err)
		os.Exit(1)
	}
}
