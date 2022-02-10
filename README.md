# omnikeeper-deploy-agent

Base Go module for an application that does the following:

Fetches host data from omnikeeper, transforms it and stores it in a JSON config file (1 file per host). Then it triggers an ansible playbook (again per host) that can load the JSON config file and perform all kinds of operations. The main usecase is to make ansible render templated configuration files and deploy them to target hosts.

## Run the sample app
Prerequisites for running the sample app:
- working ansible_playbook executable
- properly configured config file
- playbook that 
```bash
go run cmd/main.go --config config/sample-config.yml
```