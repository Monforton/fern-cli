package fern

import (
	"github.com/monforton/fern-cli/pkg/fern"
	"github.com/spf13/cobra"
)

var junitCmd = &cobra.Command{
	Use:     "junit",
	Aliases: []string{"ju"},
	Short:   "Sends the reporter data from JUnit files",
	Args:    cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		fern.ReportJunit(projectName, reportsDirectory, fernApiUrl)
	},
}

func init() {
	rootCmd.AddCommand(junitCmd)
}
