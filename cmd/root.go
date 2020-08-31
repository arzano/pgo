package cmd

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"os"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
	"strings"
)

var showBugs bool
var showPullRequests bool
var showChangelog bool
var showQAreports bool
var showDependencies bool
var showMetadata bool
var showVersions bool

var searchPackageResults bool

var rootCmd = &cobra.Command{
	Use:   "soko [searchTerm or subcommand]",
	Short: "Soko is a command line interface for packages.g.o",
	Long: `Still TODO`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
		showPackage(args[0], !searchPackageResults)
	},
}

var addCmd = &cobra.Command{
	Use:   "maintainer",
	Short: "Search for package maintainers",
	Long: `Still TODO`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
		fmt.Println("Sub-cmd: " + strings.Join(args, ", "))

	},
}

func Execute() {

	cobra.OnInitialize(initConfig)

	rootCmd.Flags().BoolVarP(&searchPackageResults, "search", "s", viper.GetBool("packages.search"), "Search for packages")
	rootCmd.Flags().BoolVarP(&showBugs, "bugs", "b", false, "Search bugs related to the packages")
	rootCmd.Flags().BoolVarP(&showPullRequests, "pull-requests", "p", false, "Show pull requests for packages")
	rootCmd.Flags().BoolVarP(&showChangelog, "changelog", "c", false, "Show changelog of the packages")
	rootCmd.Flags().BoolVarP(&showQAreports, "qa-reports", "q", false, "Show QA report for packages")
	rootCmd.Flags().BoolVarP(&showDependencies, "dependencies", "d", false, "Search dependencies of the packages")
	rootCmd.Flags().BoolVarP(&showMetadata, "metadata", "m", false, "Show metadata of the packages")
	rootCmd.Flags().BoolVarP(&showVersions, "versions", "v", false, "Show available versions of the packages")
	rootCmd.AddCommand(addCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}



func initConfig() {
	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
	}

	// Search config in home directory
	viper.SetConfigType("toml")
	viper.SetConfigFile(home + "/.soko")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("err reading config file")
		fmt.Println(err)
	}

	setViperDefaults()

	if !(showBugs || showPullRequests || showChangelog || showQAreports || showDependencies || showMetadata || showVersions) {
		if viper.GetString("packages.defaultView") == "full" {
			showBugs, showPullRequests, showChangelog, showQAreports, showDependencies, showMetadata, showVersions = true, true, true, true, true, true, true
		} else {
			showBugs, showPullRequests, showChangelog, showQAreports, showDependencies, showMetadata, showVersions = true, true, true, true, true, true, true
		}
	}
}

func setViperDefaults() {
	viper.SetDefault("packages.defaultView", "full")
	viper.SetDefault("packages.search", false)
}
