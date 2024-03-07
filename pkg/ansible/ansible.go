package ansible

import (
	"context"
	"io"

	"github.com/apenella/go-ansible/pkg/execute"
	"github.com/apenella/go-ansible/pkg/playbook"
	"github.com/max-bytes/omnikeeper-deploy-agent/v2/pkg/config"
	"github.com/sirupsen/logrus"
)

func buildPlaybookCommand(config config.AnsibleCalloutConfig, id string, variableFile string, logWriter io.Writer) *playbook.AnsiblePlaybookCmd {
	execute := execute.NewDefaultExecute(
		execute.WithWrite(logWriter),
	)

	// clone/copy options to be able to change them independently in parallel setup
	myAnsiblePlaybookOptions := *config.Options
	if myAnsiblePlaybookOptions.ExtraVars == nil {
		// if the extra vars map is not initialized through config, we do it here
		myAnsiblePlaybookOptions.ExtraVars = make(map[string]interface{})
	} else {
		// copy existing extra vars, to be able to modify them per playbook run
		tmp := make(map[string]interface{})
		for k, v := range myAnsiblePlaybookOptions.ExtraVars {
			tmp[k] = v
		}
		myAnsiblePlaybookOptions.ExtraVars = tmp
	}
	// overwrite/force set ansible variable host_id and host_variable_file
	myAnsiblePlaybookOptions.ExtraVars["host_id"] = id
	myAnsiblePlaybookOptions.ExtraVars["host_variable_file"] = variableFile

	playbook := &playbook.AnsiblePlaybookCmd{
		Playbooks:         config.Playbooks,
		ConnectionOptions: config.ConnectionOptions,
		Options:           &myAnsiblePlaybookOptions,
		Binary:            config.AnsibleBinary,
		Exec:              execute,
	}

	return playbook
}

func Callout(ctx context.Context, config config.AnsibleCalloutConfig, id string, variableFile string, simulateOnly bool, log *logrus.Entry) error {

	logWriter := log.Writer()
	defer logWriter.Close()

	playbook := buildPlaybookCommand(config, id, variableFile, logWriter)

	finalCommand, err := playbook.Command()
	if err != nil {
		return err
	}
	if simulateOnly {
		log.Tracef("[SIMULATING] Calling playbook for item %s: %s", id, finalCommand)

		// not actually calling playbook
	} else {
		log.Tracef("Calling playbook for item %s: %s", id, finalCommand)

		err = playbook.Run(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
