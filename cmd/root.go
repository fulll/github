package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/go-github/v27/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

// Github owner (user or organization)
var owner string

// Github repository name
var repository string

// Github full repository path: <owner>/<repository>
var githubRepository string

// Context
var ctx context.Context

// Http client
var client *github.Client

var rootCmd = &cobra.Command{
	Use:   "github",
	Short: "An (unofficial) Github command line client (based on Api V3)",
	Long:  ``,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		token := os.Getenv("GITHUB_TOKEN")
		if cmd.Name() != "help" && cmd.Name() != "deployment" && cmd.Name() != "deployment_status" && token == "" {
			log.Fatal("Please define GITHUB_TOKEN. See documentation to obtain one if needed: https://help.github.com/en/articles/creating-a-personal-access-token-for-the-command-line")
		}
		ctx = context.Background()
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)

		if githubRepository == "" {
			githubRepository = os.Getenv("GITHUB_REPOSITORY")
		}
		owner, repository = splitGithubRepository(githubRepository)
		if cmd.Name() != "help" && cmd.Name() != "deployment" && cmd.Name() != "deployment_status" && owner == "" && repository == "" {
			log.Fatal("Github repository is required.")
		}
	},
}

// Execute main cmd function
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func splitGithubRepository(repository string) (string, string) {
	if githubRepository == "" {
		return "", ""
	}

	values := strings.SplitN(repository, "/", 2)
	if len(values) != 2 {
		log.Fatal("Github repository should respect following format: <owner>/<repository>")
	}

	return values[0], values[1]
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&githubRepository, "repository", "r", "", "the owner and repository name. For example, octocat/Hello-World. Environment variable GITHUB_REPOSITORY will be used as a fallback.")

	rootCmd.Version = "1.0.0"
}
