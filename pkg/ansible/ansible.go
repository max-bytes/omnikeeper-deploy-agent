package ansible

import (
	"context"

	"github.com/apenella/go-ansible/pkg/playbook"
	"github.com/max-bytes/omnikeeper-deploy-agent/pkg/config"
	"github.com/sirupsen/logrus"
)

func buildPlaybookCommand(config config.AnsibleCalloutConfig, id string, variableFile string) *playbook.AnsiblePlaybookCmd {
	playbook := &playbook.AnsiblePlaybookCmd{
		Playbooks:         config.Playbooks,
		ConnectionOptions: config.ConnectionOptions,
		Options:           config.Options,
		Binary:            config.AnsibleBinary,
	}

	// if the extra vars map is not initialized through config, we do it here
	if playbook.Options.ExtraVars == nil {
		playbook.Options.ExtraVars = make(map[string]interface{})
	}

	// overwrite/force set ansible variable host_id and host_variable_file
	playbook.Options.ExtraVars["host_id"] = id
	playbook.Options.ExtraVars["host_variable_file"] = variableFile

	return playbook
}

func Callout(ctx context.Context, config config.AnsibleCalloutConfig, id string, variableFile string, simulateOnly bool, log *logrus.Logger) error {

	playbook := buildPlaybookCommand(config, id, variableFile)

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
