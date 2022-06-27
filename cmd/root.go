/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"

	"github.com/kchandra548/azure-oidc/azure"
	"github.com/kchandra548/azure-oidc/github"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "azure-oidc",
	Short: "This will help to quickly connect your Github repos to Azure",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your applicatccccccuueechcidccjfutbdbldcrjecetbecdfejhgtk
ion. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,

	Run: func(cmd *cobra.Command, args []string) {

		// Check Github authentication status if not logged in login to Github
		github.CheckStatus()

		//Check Azure authentication status if not logged in login to Azure
		fmt.Println("Fetching Azure logged in user details")
		subscriptionDetails := azure.GetUserDetails()

		//Read org flag if not get orgs list
		orgName := cmd.Flag("o").Value.String()
		if orgName == "" {
			// fmt.Println("Fetching GitHub orgs list")
			// orgs := github.GetOrgsList()
			// fmt.Println("Please select the org you want to connect to")
			// for i, org := range orgs {
			// 	fmt.Printf("%d. %s\n", i+1, org.Login)
			// }
			// fmt.Println("Enter the org number")
			//ask for Github org name
			fmt.Println("Enter the GitHub org name")
			fmt.Scanln(&orgName)
			// var orgNumber int
			// fmt.Scanln(&orgNumber)
		}

		//Read repo flag if not get repositories list
		repoName := cmd.Flag("p").Value.String()
		// repoName = "actions-playground"
		if repoName == "" {
			// fmt.Println("Fetching Github orgs repositories list")
			// repos := github.GetRepositoriesList(orgName)
			// fmt.Println("Please select the repo you want to connect to")
			// for i, repo := range repos {
			// 	fmt.Printf("%d. %s\n", i+1, repo.Name)
			// }
			// fmt.Println("Enter the repo number")
			// var repoNumber int
			// fmt.Scanln(&repoNumber)
			// repoName = repos[repoNumber-1].Name
			//ask for repo name
			fmt.Println("Enter the repo name")
			fmt.Scanln(&repoName)
		}
		// Read tenant flag
		tenant := cmd.Flag("t").Value.String()
		if tenant == "" {
			fmt.Printf("Tenant is not provided. Using default tenant %s\n", subscriptionDetails.TenantId)
			tenant = subscriptionDetails.TenantId
		}
		subscription := cmd.Flag("s").Value.String()
		if subscription == "" {
			fmt.Printf("Subscription is not provided. Using default subscription %s\n", subscriptionDetails.Id)
			subscription = subscriptionDetails.Id
		}
		resourceGroup := cmd.Flag("g").Value.String()
		if resourceGroup == "" {
			fmt.Println("Resource group is not provided. Omiting resource group")
		}
		roleName := cmd.Flag("r").Value.String()
		if roleName == "" {
			fmt.Println("Role name is empty, defaulting to 'Contributor'")
			roleName = "Contributor"
		}

		environment := cmd.Flag("e").Value.String()
		if environment == "" {
			fmt.Println("Environment is not provided. Secrets will be created at GitHub Repository level")
		}

		fmt.Println("Creating azure AAD Application")
		azureApp := azure.CreateAzureAADApp((repoName))

		fmt.Println("Creating service principaal for azure AAD Application")
		servicePrincipal := azure.CreateServicePrincipal(azureApp.AppId)

		subject := ("repo:" + orgName + "/" + repoName)
		if environment == "" {
			subject += ":ref:refs/heads/main"
		} else {
			subject += ":environment:" + environment
		}
		federatedIdentityCredentials := azure.FederatedIdentityCredentials{
			Name:      repoName + "FIC",
			Issuer:    "https://token.actions.githubusercontent.com",
			Audiences: []string{"api://AzureADTokenExchange"},
			Subject:   subject,
		}

		fmt.Println("Creating Federated Identity credentials")
		azure.CreateFIC(azureApp.Id, &federatedIdentityCredentials)

		//Assign role
		fmt.Println("Assigning role to the service principal at resource group / subscription level")
		azure.CreateRoleAssignment(subscription, resourceGroup, roleName, servicePrincipal.Id)

		//Create secrets for Github
		fmt.Println("Creating environment variable for github repo")
		github.CreateSecrets(orgName, repoName, environment, "AZURE_TENANT_ID", tenant)
		github.CreateSecrets(orgName, repoName, environment, "AZURE_SUBSCRIPTION_ID", subscription)
		github.CreateSecrets(orgName, repoName, environment, "AZURE_RESOURCE_GROUP", resourceGroup)

		fmt.Print("Gthub repo is now connected to Azure")

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {

	rootCmd.PersistentFlags().String("t", "", "Enter the tenant id")
	rootCmd.PersistentFlags().String("s", "", "Enter the subscription id")
	rootCmd.PersistentFlags().String("g", "", "Enter the resource group name")
	rootCmd.PersistentFlags().String("r", "", "Enter the role name")
	rootCmd.PersistentFlags().String("o", "", "Enter the organization name")
	rootCmd.PersistentFlags().String("p", "", "Enter the repository name")
	rootCmd.PersistentFlags().String("e", "", "Enter the environment")
}
