package azure

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

//Create user struct
type user struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type AzureSubscription struct {
	CloudName        string   `json:"cloudName"`
	HomeTenantId     string   `json:"homeTenantId"`
	Id               string   `json:"id"`
	IsDefault        bool     `json:"isDefault"`
	ManagedByTenants []string `json:"managedByTenants"`
	Name             string   `json:"name"`
	State            string   `json:"state"`
	TenantId         string   `json:"tenantId"`
	User             user     `json:"user"`
}

//Azure tenant struct
type azureTenant struct {
	Id       string `json:"id"`
	TenantId string `json:"tenantId"`
}

//Azure application struct
type azureApplication struct {
	Id    string `json:"id"`
	AppId string `json:"appId"`
}

type FederatedIdentityCredentials struct {
	Name      string   `json:"name"`
	Issuer    string   `json:"issuer"`
	Subject   string   `json:"subject"`
	Audiences []string `json:"audiences"`
}

//Resource Group Struct
type resourceGroup struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
}

//Service principal struct
type servicePrincipal struct {
	Id string `json:"id"`
}

//Azure role definition struct
type RoleDefinition struct {
	Id       string `json:"id"`
	RoleName string `json:"roleName"`
	Name     string `json:"name"`
}

func Login() AzureSubscription {
	//invoke az login command
	response, error := executeCommand("az", "login")
	if error != nil {
		fmt.Println("Error logging in to Azure.")
		os.Exit(1)
	}
	//convert to json
	var subscriptions []AzureSubscription
	json.Unmarshal([]byte(response), &subscriptions)
	for _, subscription := range subscriptions {
		if subscription.IsDefault {
			return subscription
		}
	}
	panic("Failed to get user details")
}

func GetUserDetails() AzureSubscription {
	response, error := executeCommand("az", "account", "show")
	if error != nil || response == "" {
		//Call azure login if error
		fmt.Println("No logged in user found. Please login to Azure.")
		subscription := Login()
		return subscription
	}
	//convert to json
	var azureAccount AzureSubscription
	json.Unmarshal([]byte(response), &azureAccount)
	return azureAccount
}

func GetAzureSubscriptions() []AzureSubscription {
	fmt.Printf("Getting Azure subscriptions\n")
	response, error := executeCommand("az", "account", "list")
	if error != nil || response == "" {
		panic("No subscriptions found")
	}
	var subscriptions []AzureSubscription
	json.Unmarshal([]byte(response), &subscriptions)
	return subscriptions
}

func GetAzureTenants() []azureTenant {
	response, _ := executeCommand("az", "account", "tenant", "list")
	var azureTenantList []azureTenant
	json.Unmarshal([]byte(response), &azureTenantList)
	return azureTenantList
}

func SetActiveSubscription(subscriptionId string) {
	response, _ := executeCommand("az", "account", "set", "--subscription", subscriptionId)
	fmt.Println(response)
	fmt.Println("Active subscription set to " + subscriptionId)
}

// Change active tenant
func SetActiveTenant(tenantId string) {
	response, _ := executeCommand("az", "login", "--tenant", tenantId)
	fmt.Println(response)
	fmt.Println("Active tenant set to " + tenantId)
}

//Create Resource Group
func CreateResourceGroup(resourceGroupName string, location string) resourceGroup {
	response, _ := executeCommand("az", "group", "create", "--name", resourceGroupName, "--location", location)
	var resourceGrp resourceGroup
	json.Unmarshal([]byte(response), &resourceGrp)
	return resourceGrp
}

//Get Resource Groups
func GetResourceGroups() []resourceGroup {
	response, _ := executeCommand("az", "group", "list")
	var resourceGrpList []resourceGroup
	json.Unmarshal([]byte(response), &resourceGrpList)
	return resourceGrpList
}

//Create role assignment
func CreateRoleAssignment(subscriptionId string, resourceGrpName string, role string, servicePrincipalId string) {
	var response string
	if resourceGrpName == "" {
		scope := "subscriptions/" + subscriptionId
		response, _ = executeCommand("az", "role", "assignment", "create", "--role", role, "--scope", scope, "--assignee", servicePrincipalId)

	} else {
		scope := "subscriptions/" + subscriptionId + "/resourceGroups/" + resourceGrpName
		response, _ = executeCommand("az", "role", "assignment", "create", "--role", role, "--scope", scope, "--resource-group", resourceGrpName, "--assignee", servicePrincipalId)
	}
	if response == "" {
		fmt.Println("Role assignment failed")
	}
}

func executeCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	responsePipe, _ := cmd.StdoutPipe()
	err := cmd.Start()
	if err != nil {
		fmt.Errorf("Error executing command %s %s", name, args)
		return "", err
	}
	response, err := ioutil.ReadAll(responsePipe)
	if err != nil {
		fmt.Errorf("error reading response from command %s %s", name, args)
		return "", err
	}
	cmd.Wait()
	return string(response), nil
}
func CreateAzureAADApp(appName string) azureApplication {
	// body := `{"displayName:"` + appName + `"}`
	response, _ := executeCommand("az", "ad", "app", "create", "--display-name", appName)
	var azureApplication azureApplication
	json.Unmarshal([]byte(response), &azureApplication)
	if azureApplication.Id == "" || azureApplication.AppId == "" {
		panic("Failed to create Azure AAD app")
	}
	return azureApplication
}

//Create Service principal
func CreateServicePrincipal(appId string) servicePrincipal {
	if appId == "" {
		panic("AppId is empty")
	}
	response, error := executeCommand("az", "ad", "sp", "create", "--id", appId)
	if error != nil {
		fmt.Println(response)
		panic("Failed to create service principal")
	}
	var servicePrincipal servicePrincipal
	json.Unmarshal([]byte(response), &servicePrincipal)
	if servicePrincipal.Id == "" {
		panic("Failed to create service principal")
	}
	return servicePrincipal
}

func CreateFIC(id string, federatedIdentityCredentials *FederatedIdentityCredentials) {

	body, _ := json.Marshal(federatedIdentityCredentials)
	executeCommand("az", "rest",
		"--method", "post",
		"--url", "https://graph.microsoft.com/beta/applications/"+id+"/federatedIdentityCredentials/",
		"--headers", "Content-Type=application/json",
		"--body", string(body))
}

func GetRoleDefinitions() []RoleDefinition {
	response, _ := executeCommand("az", "role", "definition", "list")
	var roleDefinitions []RoleDefinition
	json.Unmarshal([]byte(response), &roleDefinitions)
	return roleDefinitions
}
