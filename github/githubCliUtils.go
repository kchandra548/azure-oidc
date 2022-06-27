package github

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
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
	fmt.Println(response)
	if strings.Contains(response, "You are not logged into any GitHub hosts. Run gh auth login to authenticate.") {
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
	var resources []resource
	err := getRestClient().Get("user/orgs", &resources)
	if err != nil {
		fmt.Println("failed to get orgs")
	}
	return resources
}

//Create Repository Secrets
func CreateSecrets(orgName string, repoName string, environment string, secretName string, secretValue string) {
	args := []string{"secret", "set", secretName, "-R", orgName + "/" + repoName, "--body", secretValue}
	if environment != "" {
		args = append(args, "--env")
		args = append(args, environment)
	}
	response, err := exec(args)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(response)
}

//Get Repository Environments
func GetRepositoryEnvironments(orgName string, repoName string) EnvironmnetsResponse {

	var environments EnvironmnetsResponse

	er := getRestClient().Get("repos/"+orgName+"/"+repoName+"/environments", &environments)
	if er != nil {
		fmt.Println("failed to get environments")
	}

	return environments
}

//Invoke API
func getRestClient() api.RESTClient {
	client, err := gh.RESTClient(nil)
	if err != nil {
		fmt.Println(err)
		panic("failed to get GH rest client")
	}
	return client
}

type resource struct {
	Name  string `json:"name"`
	Id    int64  `json:"id"`
	Login string `json:"login"`
}

type EnvironmnetsResponse struct {
	Environmnets []Environment `json:"environments"`
}
type Environment struct {
	Name string `json:"name"`
	Id   int64  `json:"id"`
}
