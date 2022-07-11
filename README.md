# azuer-oidc
## Prerequisites
* Azure CLI (Install from [here](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli))
* GH CLI (Install from [here](https://github.com/cli/cli#installation))

## Installation
* Run `go build`
* Run `azuer-oidc`

## With Defaults

`azure-oidc --useDefaults yes --org <gh-org> --repo <gh-repo>`

by setting useDefaults to yes, the following defaults will be used from the Azure CLI session:
* Subscription ID
* Tenant ID



And the following default values will be used:
* Role as `Contributor`
* Resource Group as `<org>-<repo>-<env>`

## Without Defaults

`azure-oidc --org <gh-org> --repo <gh-repo> --enviroment <gh-repo-env> --tenant <tenant-id> --subscription <subscription-id> --role <role-name> --resource-group <resource-group-name>  --role <role>`

Note: All the flags/arguments are optional. If you don't specify any value, it will prompt you for it.
