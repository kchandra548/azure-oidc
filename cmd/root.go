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
		yesOrNo := []string{"Yes", "No"}

		// Check Github authentication status if not logged in login to Github
		fmt.Println("Checking Github authentication status")
		github.CheckStatus()

		//Check Azure authentication status if not logged in login to Azure
		fmt.Println("Checking Azure authentication status")
		subscriptionDetails := azure.GetUserDetails()
		fmt.Printf("You are loggedin as: %s \n", subscriptionDetails.User.Name)

		// Check with user if he wants to proceed with all default values
		useDefaults := useDefaults(propmptOptions("Do you want to use default values? Note: This will use availble information from existing session", yesOrNo))

		if useDefaults {
			fmt.Println("Using default values")
		}
		if !useDefaults {
			result := propmptOptions("Do you want to continue as "+subscriptionDetails.User.Name+"?", yesOrNo)
			if result == "N" || result == "n" || result == "No" || result == "no" {
				azure.Login()
			}
		}
		//Read org flag if not get orgs list
		orgName := cmd.Flag("o").Value.String()
		if orgName == "" {
			fmt.Println("No organization name provided")
			orgName = promptForInput("Please enter the organization name")
		}

		//Read repo flag if not get repositories list
		repoName := cmd.Flag("p").Value.String()
		if repoName == "" {
			fmt.Println("No repository name provided")
			repoName = promptForInput("Please enter the repository name")
		}
		// Read tenant flag
		tenant := cmd.Flag("t").Value.String()
		if tenant == "" {
			//Check if user want to proceed with default tenant if not provide options
			fmt.Println("No tenant id provided")
			if !useDefaults {
				tenants := azure.GetAzureTenants()
				tenantOptions := make([]string, len(tenants))
				for i, tenant := range tenants {
					tenantOptions[i] = tenant.TenantId
				}

				fmt.Println("Fetching Azure tenants list")
				tenant = propmptOptions("Please select the tenant you want to connect to", tenantOptions)
				if tenant != subscriptionDetails.TenantId {
					azure.SetActiveTenant(tenant)
					subscriptionDetails = azure.GetUserDetails()
				}

			} else {
				fmt.Printf("Using default tenant: %s\n", subscriptionDetails.TenantId)
				tenant = subscriptionDetails.TenantId
			}
		} else {
			//Check if tenant is default tenant
			if tenant != subscriptionDetails.TenantId {
				fmt.Printf("Given tenant %s is not same as loggged in User default tenant %s\n", tenant, subscriptionDetails.TenantId)
				fmt.Printf("Switchin to Tenant %s\n", tenant)
				fmt.Println("Please login to Azure with the given tenant")
				azure.SetActiveTenant(tenant)
				subscriptionDetails = azure.GetUserDetails()
			}
		}

		subscription := cmd.Flag("s").Value.String()
		if subscription == "" {
			fmt.Println("No subscription id provided")
			if !useDefaults {
				subscriptions := azure.GetAzureSubscriptions()
				subscriptionOptions := make([]string, len(subscriptions))
				for i, sub := range subscriptions {
					subscriptionOptions[i] = sub.Name + "(" + sub.Id + ")"
				}

				subscription = propmptOptions("Please select the subscription you want to connect to", subscriptionOptions)
				subscription = strings.Split(subscription, "(")[1]
				if subscription != subscriptionDetails.Id {
					azure.SetActiveSubscription(subscription)
					subscriptionDetails = azure.GetUserDetails()
				}
			} else {
				fmt.Printf("Using default subscription: %s\n", subscriptionDetails.Id)
				subscription = subscriptionDetails.Id
			}
		} else {
			// Check if given subscription is same as logged in user subscription
			if subscription != subscriptionDetails.Id {
				fmt.Printf("Given subscription %s is not same as logged in user default subscription %s\n", subscription, subscriptionDetails.Id)
				fmt.Printf("Switching subscription to %s\n", subscription)
				azure.SetActiveSubscription(subscription)
			}
		}

		resourceGroup := cmd.Flag("g").Value.String()
		if resourceGroup == "" {
			fmt.Println("Resource group is not provided.")
			if !useDefaults {
				newOrExisting := propmptOptions("Do you want to create a new resource group or use an existing one ?", []string{"New", "Existing", "Proceed without resource group"})
				if newOrExisting == "New" || newOrExisting == "new" {
					resourceGroup = orgName + "-" + repoName
					fmt.Printf("Creating new resource group: %s\n", resourceGroup)
					azure.CreateResourceGroup(resourceGroup, "eastus")
				} else if newOrExisting == "Existing" || newOrExisting == "existing" {
					fmt.Println("Fetching resource groups list")
					resourceGrpList := azure.GetResourceGroups()
					resourceGrpOptions := make([]string, len(resourceGrpList))
					for i, resourceGrp := range resourceGrpList {
						resourceGrpOptions[i] = resourceGrp.Name
					}
					resourceGroup = propmptOptions("Please select the resource group you want to connect to", resourceGrpOptions)
				}
			}
			fmt.Println("Omitting resource group")
		}

		roleName := cmd.Flag("r").Value.String()
		if roleName == "" {
			fmt.Println("No Role name provided")
			if !useDefaults {
				fmt.Println("Fetching Role definitions")
				roleList := azure.GetRoleDefinitions()
				roleOptions := make([]string, len(roleList))
				for i, role := range roleList {
					roleOptions[i] = role.RoleName
				}
				roleName = propmptOptions("Please select the role you want to assign to the Service Principal", roleOptions)
			} else {
				fmt.Println("Using default role: Contributor")
				roleName = "Contributor"
			}
		}

		environment := cmd.Flag("e").Value.String()
		if environment == "" {
			fmt.Println("Environment is not provided.")
			if !useDefaults {
				fmt.Println("Fetching environments list")
				environments := github.GetRepositoryEnvironments(orgName, repoName)
				if len(environments.Environmnets) == 0 {
					fmt.Println("No environments found for this repository, Secret will be created at repo level")
				} else {
					environmentOptions := make([]string, len(environments.Environmnets))
					for i, env := range environments.Environmnets {
						environmentOptions[i] = env.Name
					}
					environment = propmptOptions("Please select the environment you want to connect to", environmentOptions)
				}
			} else {
				fmt.Println("No environments found for this repository, Secret will be created at repo level")
			}

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

		fmt.Printf("Please visit Github Repo: https://github.com/%s/%s/settings", orgName, repoName)

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

func confirmPrompt(label string) string {
	prompt := promptui.Prompt{
		Label:     label,
		IsConfirm: true,
	}

	result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Recieved invalid input, so proceeding with default value")
		fmt.Printf("Prompt failed %v\n", err)
		return "Y"
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

func isTrue(selectedString string) bool {
	if selectedString == "yes" || selectedString == "Yes" {
		return true
	}
	return false
}

func useDefaults(defaultString string) bool {
	if defaultString == "yes" || defaultString == "Yes" {
		return true
	}
	return false
}
