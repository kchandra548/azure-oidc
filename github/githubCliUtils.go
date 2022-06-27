package github

import (
	"encoding/json"
	"fmt"

	"github.com/cli/go-gh"
)

//execute command
func exec(args []string) (string, error) {
	stdOut, _, err := gh.Exec(args...)
	if err != nil {
		return "", err
	}
	return stdOut.String(), nil
}

//Check status
func CheckStatus() {
	args := []string{"auth", "status"}
	response, _ := exec(args)
	if response == "You are not logged into any GitHub hosts. Run gh auth login to authenticate." {
		fmt.Println("You are not logged into any GitHub hosts. Run gh auth login to authenticate. Please login to continue.")
		login()
	}
}

func login() {
	args := []string{"auth", "login"}
	response, _ := exec(args)
	fmt.Println(response)
}

//Get Repositories List
func GetRepositoriesList(orgName string) []resource {
	args := []string{"repo", "list", orgName}
	response, _ := exec(args)
	var resources []resource
	json.Unmarshal([]byte(response), &response)
	return resources
}

//Get Orgs list
func GetOrgsList() []resource {
	orgs := invokeAPI("organizations")
	return orgs
}

//Create Repository Secrets
func CreateSecrets(orgName string, repoName string, environment string, secretName string, secretValue string) {
	args := []string{"secret", "set", secretName, "-R", orgName + "/" + repoName, "--body", secretValue}
	if environment != "" {
		args = append(args, "--environment")
		args = append(args, environment)
	}
	response, err := exec(args)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(response)
}

//Get Repository Environments
func GetRepositoryEnvironments(orgName string, repoName string) {

	environment := invokeAPI("/repos/" + orgName + "/" + repoName + "/environments")
	fmt.Println(environment)
}

//Invoke API
func invokeAPI(uri string) []resource {
	client, err := gh.RESTClient(nil)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	var response []resource
	err = client.Get(uri, &response)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return response
}

type resource struct {
	Name  string `json:"name"`
	Id    int64  `json:"id"`
	Login string `json:"login"`
}
