package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/google/go-github/v27/github"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var deploymentStatus = &cobra.Command{
	Use:   "deployment_status",
	Short: "Deployment Status",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(0)
	},
}

var deploymentStatusListCmd = &cobra.Command{
	Use:     "list ID",
	Short:   "List deployment statuses",
	Aliases: []string{"ls"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		deploymentID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			log.Fatal(err)
		}
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

		deploymentStatuses, _, err := client.Repositories.ListDeploymentStatuses(ctx, owner, repository, deploymentID, &listOptions)
		if err != nil {
			log.Fatal(err)
		}

		json, err := json.MarshalIndent(deploymentStatuses, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", json)
	},
}

var deploymentStatusGetCmd = &cobra.Command{
	Use:     "create DEPLOYMENT_ID DEPLOYMENT_STATUS_ID",
	Short:   "Get a single deployment status",
	Aliases: []string{"g"},
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		deploymentID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			log.Fatal(err)
		}

		deploymentStatusID, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			log.Fatal(err)
		}

		deploymentStatus, _, err := client.Repositories.GetDeploymentStatus(ctx, owner, repository, deploymentID, deploymentStatusID)
		if err != nil {
			log.Fatal(err)
		}

		json, err := json.MarshalIndent(deploymentStatus, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", json)
	},
}

func isValidState(state string) error {
	switch state {
	case
		"error",
		"failure",
		"inactive",
		"in_progress",
		"queued",
		"pending":
		return nil
	}

	return errors.New("invalid state: should be one of error, failure, inactive, in_progress, queued or pending")
}

var deploymentStatusCreateCmd = &cobra.Command{
	Use:     "create DEPLOYMENT_ID STATE",
	Short:   "Create a deployment status",
	Aliases: []string{"c"},
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		deploymentID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			log.Fatal(err)
		}

		err = isValidState(args[1])
		if err != nil {
			log.Fatal(err)
		}
		state := args[1]

		viper.BindPFlags(cmd.PersistentFlags())
		var deploymentStatusRequest github.DeploymentStatusRequest
		err = viper.Unmarshal(&deploymentStatusRequest)
		deploymentStatusRequest.State = &state
		if err != nil {
			log.Fatal(err)
		}

		deploymentStatus, _, err := client.Repositories.CreateDeploymentStatus(ctx, owner, repository, deploymentID, &deploymentStatusRequest)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%v\n", deploymentStatus.GetID())
	},
}

func init() {
	deploymentStatus.AddCommand(deploymentStatusGetCmd)

	deploymentStatus.AddCommand(deploymentStatusListCmd)
	deploymentStatusListCmd.PersistentFlags().IntP("page", "p", 1, "for paginated result sets, page of results to retrieve.")
	deploymentStatusListCmd.PersistentFlags().IntP("perPage", "l", 10, "for paginated result sets, the number of results to include per page.")

	deploymentStatus.AddCommand(deploymentStatusCreateCmd)
	deploymentStatusCreateCmd.PersistentFlags().String("logURL", "", "The full URL of the deployment's output.")
	deploymentStatusCreateCmd.PersistentFlags().StringP("description", "d", "", "A short description of the status. The maximum description length is 140 characters.")
	deploymentStatusCreateCmd.PersistentFlags().StringP("environment", "e", "", "Name for the target deployment environment, which can be changed when setting a deploy status. For example, production, staging, or qa.")
	deploymentStatusCreateCmd.PersistentFlags().StringP("environmentURL", "u", "", "for paginated result sets, page of results to retrieve.")
	deploymentStatusCreateCmd.PersistentFlags().BoolP("autoInactive", "a", true, "Adds a new inactive status to all prior non-transient, non-production environment deployments with the same repository and environment name as the created status's deployment. An inactive status is only added to deployments that had a success state.")

	deploymentStatus.Aliases = []string{"ds"}

	rootCmd.AddCommand(deploymentStatus)
}
