package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/v27/github"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var release = &cobra.Command{
	Use:   "release",
	Short: "Release",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(0)
	},
}

func getRelease(ctx context.Context, client *github.Client, owner string, repository string, ref string) (*github.RepositoryRelease, *github.Response, error) {
	if strings.ToLower(ref) == "latest" {
		return client.Repositories.GetLatestRelease(ctx, owner, repository)
	}

	releaseID, err := strconv.ParseInt(ref, 10, 64)
	if err == nil {
		return client.Repositories.GetRelease(ctx, owner, repository, releaseID)
	}

	return client.Repositories.GetReleaseByTag(ctx, owner, repository, ref)
}

var releaseGetCmd = &cobra.Command{
	Use:     "get ID|LATEST|TAG",
	Short:   "Get a release",
	Example: "github release get latest\ngithub release get 133742\ngithub release get 19.01.2",
	Aliases: []string{"g"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		release, _, err := getRelease(ctx, client, owner, repository, args[0])
		if err != nil {
			log.Fatal(err)
		}

		json, err := json.MarshalIndent(release, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", json)
	},
}

var releaseListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List releases",
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		pFlags := cmd.PersistentFlags()
		page, err := pFlags.GetInt("page")
		if err != nil {
			log.Fatal(err)
		}
		perPage, err := pFlags.GetInt("perPage")
		if err != nil {
			log.Fatal(err)
		}
		listOptions := github.ListOptions{Page: page, PerPage: perPage}

		releases, _, err := client.Repositories.ListReleases(ctx, owner, repository, &listOptions)
		if err != nil {
			log.Fatal(err)
		}

		json, err := json.MarshalIndent(releases, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", json)
	},
}

var releaseCreateCmd = &cobra.Command{
	Use:     "create TAG_NAME",
	Short:   "Create a release",
	Aliases: []string{"c"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		viper.BindPFlags(cmd.PersistentFlags())

		var repositoryRelease github.RepositoryRelease
		err := viper.Unmarshal(&repositoryRelease)
		repositoryRelease.TagName = &args[0]
		if err != nil {
			log.Fatal(err)
		}

		release, _, err := client.Repositories.CreateRelease(ctx, owner, repository, &repositoryRelease)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%v\n", release.GetID())
	},
}

var releaseEditCmd = &cobra.Command{
	Use:     "edit ID",
	Short:   "Edit a release",
	Aliases: []string{"e"},
	Args:    cobra.ExactArgs(1),
	Example: `github release edit 133742 --draft=false # publish a release
printf "%s\nsomething to add?" $(github release get 133742 | jq .body) | github release edit 133742 --body -
`,
	Run: func(cmd *cobra.Command, args []string) {
		releaseID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			log.Fatal(err)
		}

		pFlags := cmd.PersistentFlags()
		repositoryRelease := github.RepositoryRelease{}

		if pFlags.Changed("tagName") == true {
			value, err := pFlags.GetString("tagName")
			if err != nil {
				log.Fatal(err)
			}
			repositoryRelease.TagName = &value
		}

		if pFlags.Changed("targetCommitish") == true {
			value, err := pFlags.GetString("targetCommitish")
			if err != nil {
				log.Fatal(err)
			}
			repositoryRelease.TargetCommitish = &value
		}

		if pFlags.Changed("name") == true {
			value, err := pFlags.GetString("name")
			if err != nil {
				log.Fatal(err)
			}
			repositoryRelease.Name = &value
		}

		if pFlags.Changed("body") == true {
			value, err := pFlags.GetString("body")
			if err != nil {
				log.Fatal(err)
			}
			if value == "-" {
				data, err := ioutil.ReadAll(os.Stdin)
				if err != nil {
					log.Fatal(err)
				}
				value = string(data)
			}
			repositoryRelease.Body = &value
		}

		if pFlags.Changed("draft") == true {
			value, err := pFlags.GetBool("draft")
			if err != nil {
				log.Fatal(err)
			}
			repositoryRelease.Draft = &value
		}

		if pFlags.Changed("prerelease") == true {
			value, err := pFlags.GetBool("prerelease")
			if err != nil {
				log.Fatal(err)
			}
			repositoryRelease.Prerelease = &value
		}

		_, _, err = client.Repositories.EditRelease(ctx, owner, repository, releaseID, &repositoryRelease)
		if err != nil {
			log.Fatal(err)
		}
	},
}

var releaseDeleteCmd = &cobra.Command{
	Use:     "delete ID",
	Short:   "Delete a release",
	Aliases: []string{"d"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		releaseID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			log.Fatal(err)
		}

		_, err = client.Repositories.DeleteRelease(ctx, owner, repository, releaseID)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	release.Aliases = []string{"r"}

	release.AddCommand(releaseCreateCmd)
	releaseCreateCmd.PersistentFlags().StringP("targetCommitish", "c", "", "Specifies the commitish value that determines where the Git tag is created from. Can be any branch or commit SHA. Unused if the Git tag already exists. Default: the repository's default branch (usually `master`).")
	releaseCreateCmd.PersistentFlags().StringP("name", "n", "", "The name of the release.")
	releaseCreateCmd.PersistentFlags().StringP("body", "b", "", "Text describing the contents of the tag.")
	releaseCreateCmd.PersistentFlags().BoolP("draft", "d", false, "true to create a draft (unpublished) release, false to create a published one.")
	releaseCreateCmd.PersistentFlags().BoolP("prerelease", "p", false, "true to identify the release as a prerelease. false to identify the release as a full release.")

	release.AddCommand(releaseEditCmd)
	releaseEditCmd.PersistentFlags().StringP("tagName", "t", "", "Specifies the commitish value that determines where the Git tag is created from. Can be any branch or commit SHA. Unused if the Git tag already exists.")
	releaseEditCmd.PersistentFlags().StringP("targetCommitish", "c", "", "Specifies the commitish value that determines where the Git tag is created from. Can be any branch or commit SHA. Unused if the Git tag already exists. Default: the repository's default branch (usually `master`).")
	releaseEditCmd.PersistentFlags().StringP("name", "n", "", "The name of the release.")
	releaseEditCmd.PersistentFlags().StringP("body", "b", "", "Text describing the contents of the tag.")
	releaseEditCmd.PersistentFlags().BoolP("draft", "d", false, "true to create a draft (unpublished) release, false to create a published one.")
	releaseEditCmd.PersistentFlags().BoolP("prerelease", "p", false, "true to identify the release as a prerelease. false to identify the release as a full release.")

	release.AddCommand(releaseDeleteCmd)

	release.AddCommand(releaseGetCmd)

	release.AddCommand(releaseListCmd)

	rootCmd.AddCommand(release)
}
