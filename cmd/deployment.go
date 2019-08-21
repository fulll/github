package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/google/go-github/v27/github"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var deployment = &cobra.Command{
	Use:   "deployment",
	Short: "Deployment",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(0)
	},
}

var deploymentGetCmd = &cobra.Command{
	Use:     "get ID",
	Short:   "Get a deployment",
	Aliases: []string{"g"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		deploymentID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			log.Fatal(err)
		}

		deployment, _, err := client.Repositories.GetDeployment(ctx, owner, repository, deploymentID)
		if err != nil {
			log.Fatal(err)
		}

		json, err := json.MarshalIndent(deployment, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", json)
	},
}

var deploymentListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List deployments",
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		viper.BindPFlags(cmd.PersistentFlags())
		var deploymentsListOptions github.DeploymentsListOptions
		err := viper.Unmarshal(&deploymentsListOptions)
		if err != nil {
			log.Fatal(err)
		}

		var listOptions github.ListOptions
		err = viper.Unmarshal(&listOptions)
		if err != nil {
			log.Fatal(err)
		}

		deploymentsListOptions.ListOptions = listOptions

		deployments, _, err := client.Repositories.ListDeployments(ctx, owner, repository, &deploymentsListOptions)
		if err != nil {
			log.Fatal(err)
		}

		json, err := json.MarshalIndent(deployments, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", json)
	},
}

var deploymentCreateCmd = &cobra.Command{
	Use:     "create REF",
	Short:   "Create a deployment",
	Aliases: []string{"c"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		viper.BindPFlags(cmd.PersistentFlags())

		var deploymentRequest github.DeploymentRequest
		err := viper.Unmarshal(&deploymentRequest)
		deploymentRequest.Ref = &args[0]
		if len(*deploymentRequest.RequiredContexts) == 0 {
			requiredContexts := make([]string, 0)
			deploymentRequest.RequiredContexts = &requiredContexts
		}

		if err != nil {
			log.Fatal(err)
		}

		deployment, _, err := client.Repositories.CreateDeployment(ctx, owner, repository, &deploymentRequest)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Successfully created deployment: %v\n", deployment.GetID())
	},
}

func init() {
	deployment.Aliases = []string{"d"}

	deployment.AddCommand(deploymentCreateCmd)
	deploymentCreateCmd.PersistentFlags().StringP("task", "t", "deploy", "Specifies a task to execute (e.g., deploy or deploy:migrations).")
	deploymentCreateCmd.PersistentFlags().BoolP("autoMerge", "a", true, "Attempts to automatically merge the default branch into the requested ref, if it's behind the default branch.")
	deploymentCreateCmd.PersistentFlags().StringSliceP("requiredContexts", "c", []string{}, "The status contexts to verify against commit status checks. If you omit this parameter, GitHub verifies all unique contexts before creating a deployment. To bypass checking entirely, pass an empty array. Defaults to all unique contexts.")
	deploymentCreateCmd.PersistentFlags().StringP("payload", "p", "", "JSON payload with extra information about the deployment.")
	deploymentCreateCmd.PersistentFlags().StringP("environment", "e", "production", "Name for the target deployment environment (e.g., production, staging, qa).")
	deploymentCreateCmd.PersistentFlags().StringP("description", "d", "", "Short description of the deployment.")
	deploymentCreateCmd.PersistentFlags().Bool("transientEnvironment", false, "Specifies if the given environment is specific to the deployment and will no longer exist at some point in the future.")
	deploymentCreateCmd.PersistentFlags().BoolP("productionEnvironment", "i", true, "Specifies if the given environment is one that end-users directly interact with. Default: true when environment is production and false otherwise.")

	deployment.AddCommand(deploymentGetCmd)

	deployment.AddCommand(deploymentListCmd)
	deploymentListCmd.PersistentFlags().StringP("sha", "s", "", "sha of the deployment")
	deploymentListCmd.PersistentFlags().StringP("ref", "", "", "list deployments for a given ref.")
	deploymentListCmd.PersistentFlags().StringP("task", "t", "", "list deployments for a given task.")
	deploymentListCmd.PersistentFlags().StringP("environment", "e", "", "list deployments for a given environment.")
	deploymentListCmd.PersistentFlags().IntP("page", "p", 1, "for paginated result sets, page of results to retrieve.")
	deploymentListCmd.PersistentFlags().IntP("perPage", "l", 10, "for paginated result sets, the number of results to include per page.")

	rootCmd.AddCommand(deployment)
}
