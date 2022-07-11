/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/kchandra548/azure-oidc/azure"
	"github.com/kchandra548/azure-oidc/github"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var tenant string
var subscription string
var resourceGroup string
var role string
var org string
var repo string
var environment string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "azure-oidc",
	Short: "Connect Github to Azure for Workflow automation",
	Long:  `Connect Github to Azure for Workflow automation`,

	Run: func(cmd *cobra.Command, args []string) {
		yesOrNo := []string{"Yes", "No"}

		// Check Github authentication status if not logged in login to Github
		fmt.Println("Checking Github authentication status")
		github.CheckStatus()

		//Check Azure authentication status if not logged in login to Azure
		fmt.Println("Checking Azure authentication status")
		subscriptionDetails := azure.GetUserDetails()
		fmt.Printf("You are loggedin as: %s \n", subscriptionDetails.User.Name)

		// Read values from flags
		tenant = cmd.Flag("tenant").Value.String()
		subscription = cmd.Flag("subscription").Value.String()
		resourceGroup = cmd.Flag("resource-group").Value.String()
		role = cmd.Flag("role").Value.String()
		org = cmd.Flag("org").Value.String()
		repo = cmd.Flag("repo").Value.String()
		environment = cmd.Flag("environment").Value.String()
		//read userdefaults from flag and check equal to yes ignore case
		useDefaults := strings.EqualFold(cmd.Flag("useDefaults").Value.String(), "yes") || strings.EqualFold(cmd.Flag("useDefaults").Value.String(), "y")

		// Check with user if he wants to proceed with all default values
		if !useDefaults {
			useDefaults = isUseDefaultsOpted(propmptOptions("Do you want to use default Azure Subscription details? Note: This will use availble information from existing Azure user session", yesOrNo))
		}

		if !useDefaults {
			result := propmptOptions("Do you want to continue as "+subscriptionDetails.User.Name+"?", yesOrNo)
			if strings.EqualFold(result, "no") || strings.EqualFold(result, "n") {
				azure.Login()
			}
		}
		if org == "" {
			org = promptForInput("Please enter the GitHub organization name")
		}

		if repo == "" {
			repo = promptForInput("Please enter the GitHub repository name")
		}
		resolveEnvironment(&environment)

		if useDefaults {
			resolveDefaultValues(subscriptionDetails)
		}

		resolveAzureTenant(&subscriptionDetails)

		resolveAzureSubscription(&subscriptionDetails)

		resolveAzureResourceGroup()

		resolveAzureRole()

		createAndConfigureAzureResources()

		//Create secrets for Github
		fmt.Println("Creating environment variable for github repo")
		github.CreateSecrets(org, repo, environment, "AZURE_TENANT_ID", tenant)
		github.CreateSecrets(org, repo, environment, "AZURE_SUBSCRIPTION_ID", subscription)
		github.CreateSecrets(org, repo, environment, "AZURE_RESOURCE_GROUP", resourceGroup)

		fmt.Println("🎉🎉 FINISHED!! Gthub repo is now connected to Azure")

		fmt.Printf("Please visit Github Repo: https://github.com/%s/%s/settings", org, repo)

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

	rootCmd.PersistentFlags().StringVarP(&tenant, "tenant", "t", "", "Enter the Azure Tenant id")
	rootCmd.PersistentFlags().StringVarP(&subscription, "subscription", "s", "", "Enter the Azure Subscription id")
	rootCmd.PersistentFlags().StringVarP(&resourceGroup, "resource-group", "g", "", "Enter the Azure Resource Group")
	rootCmd.PersistentFlags().StringVarP(&role, "role", "r", "", "Enter the Azure Role Name")
	rootCmd.PersistentFlags().StringVarP(&org, "org", "o", "", "Enter the Github Organization Name")
	rootCmd.PersistentFlags().StringVarP(&repo, "repo", "R", "", "Enter the Github Repository Name")
	rootCmd.PersistentFlags().StringVarP(&environment, "environment", "e", "", "Enter the Github Environment Name")
	rootCmd.PersistentFlags().String("useDefaults", "", "Use Defaults to create a connection quickly")
}

func propmptOptions(label string, options []string) string {
	prompt := promptui.Select{
		Label: label,
		Items: options,
	}

	_, result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return ""
	}
	return result
}

func promptForInput(label string) string {
	validate := func(input string) error {
		//Check if string lenth is less than 3
		if len(input) < 3 {
			return errors.New("invalid number")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    label,
		Validate: validate,
	}

	result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return ""
	}

	return result
}

func isUseDefaultsOpted(defaultString string) bool {
	return strings.EqualFold(defaultString, "yes") || strings.EqualFold(defaultString, "y")
}
func resolveDefaultValues(subscriptionDetails azure.AzureSubscription) {
	fmt.Println("Continuing with default values")
	if tenant == "" {
		tenant = subscriptionDetails.TenantId
		fmt.Println("Using tenant: " + subscriptionDetails.TenantId)
	}
	if subscription == "" {
		subscription = subscriptionDetails.Id
		fmt.Println("Using subscription: " + subscriptionDetails.Id)
	}
	if resourceGroup == "" {
		resourceGroup = org + "-" + repo
		if environment != "" {
			resourceGroup = resourceGroup + "-" + environment
		}
		fmt.Println("Creating a new resource group: " + resourceGroup)
	}
	if role == "" {
		role = "Contributor"
		fmt.Println("Using Contributor as default role")
	}
}
func resolveEnvironment(environment *string) {
	if *environment == "" {
		environments := github.GetRepositoryEnvironments(org, repo)
		if len(environments.Environmnets) == 0 {
			fmt.Println("No environments found for this repository, Secret will be created at repo level")
		} else {
			environmentOptions := make([]string, len(environments.Environmnets))
			for i, env := range environments.Environmnets {
				environmentOptions[i] = env.Name
			}
			*environment = propmptOptions("Please select the environment you want to connect to", environmentOptions)
		}
	} else {
		fmt.Println("No environments found for this repository, Secret will be created at repo level")
	}
}

func resolveAzureTenant(subscriptionDetails *azure.AzureSubscription) {
	if tenant == "" {
		fmt.Println("Please select your Azure tenant")
		tenants := azure.GetAzureTenants()
		tenantOptions := make([]string, len(tenants))
		for i, tenant := range tenants {
			tenantOptions[i] = tenant.TenantId
		}

		tenant = propmptOptions("Please select the tenant you want to connect to", tenantOptions)
		if tenant != subscriptionDetails.TenantId {
			azure.SetActiveTenant(tenant)
			*subscriptionDetails = azure.GetUserDetails()
		}
	} else {
		if tenant != subscriptionDetails.TenantId {
			fmt.Printf("Given tenant %s is not same as loggged in User default tenant %s\n", tenant, subscriptionDetails.TenantId)
			fmt.Printf("Switching to Tenant %s\n", tenant)
			fmt.Println("Please login to Azure with the given tenant")
			azure.SetActiveTenant(tenant)
			*subscriptionDetails = azure.GetUserDetails()
		}
	}
}

func resolveAzureSubscription(subscriptionDetails *azure.AzureSubscription) {
	if subscription == "" {
		subscriptions := azure.GetAzureSubscriptions()
		subscriptionOptions := make([]string, len(subscriptions))
		for i, sub := range subscriptions {
			subscriptionOptions[i] = sub.Name + "(" + sub.Id + ")"
		}

		subscription = propmptOptions("Please select the subscription you want to connect to", subscriptionOptions)
		subscription = strings.Split(subscription, "(")[1]
		if subscription != subscriptionDetails.Id {
			azure.SetActiveSubscription(subscription)
			*subscriptionDetails = azure.GetUserDetails()
		}
	} else {
		if subscription != subscriptionDetails.Id {
			fmt.Printf("Given subscription %s is not same as logged in user default subscription %s\n", subscription, subscriptionDetails.Id)
			fmt.Printf("Switching to subscription: %s\n", subscription)
			azure.SetActiveSubscription(subscription)
		}
	}
}

func resolveAzureResourceGroup() {
	if resourceGroup == "" {
		newOrExisting := propmptOptions("Do you want to create a new resource group or use an existing one ?", []string{"New", "Existing"})
		if strings.EqualFold(newOrExisting, "new") {
			resourceGroup = promptForInput("Please enter a name for the new resource group")
			fmt.Printf("Creating new resource group: %s\n", resourceGroup)
			azure.CreateResourceGroup(resourceGroup, "eastus")
		} else {
			fmt.Println("Fetching resource groups list")
			resourceGrpList := azure.GetResourceGroups()
			resourceGrpOptions := make([]string, len(resourceGrpList))
			for i, resourceGrp := range resourceGrpList {
				resourceGrpOptions[i] = resourceGrp.Name
			}
			resourceGroup = propmptOptions("Please select the resource group you want to connect to", resourceGrpOptions)
		}
	}
}

func resolveAzureRole() {
	if role == "" {
		roleList := azure.GetRoleDefinitions()
		roleOptions := make([]string, len(roleList))
		for i, role := range roleList {
			roleOptions[i] = role.RoleName
		}
		role = propmptOptions("Please select the role you want to assign to the Service Principal", roleOptions)
	}
}

func createAndConfigureAzureResources() {
	fmt.Println("Creating azure AAD Application")
	azureApp := azure.CreateAzureAADApp((org + "-" + repo))

	fmt.Println("Creating service principal for azure AAD Application")
	servicePrincipal := azure.CreateServicePrincipal(azureApp.AppId)

	subject := ("repo:" + org + "/" + repo)
	if environment == "" {
		subject += ":ref:refs/heads/main"
	} else {
		subject += ":environment:" + environment
	}

	federatedIdentityCredentials := azure.FederatedIdentityCredentials{
		Name:      org + "-" + repo + "-" + "fic",
		Issuer:    "https://token.actions.githubusercontent.com",
		Audiences: []string{"api://AzureADTokenExchange"},
		Subject:   subject,
	}

	fmt.Printf("Creating Federated Identity credentials %s . \n", federatedIdentityCredentials)
	azure.CreateFIC(azureApp.Id, &federatedIdentityCredentials)

	//Assign role
	fmt.Printf("Assigning role %s to the service principal at resourceGroup, %s level. \n", role, resourceGroup)
	azure.CreateRoleAssignment(subscription, resourceGroup, role, servicePrincipal.Id)
}
