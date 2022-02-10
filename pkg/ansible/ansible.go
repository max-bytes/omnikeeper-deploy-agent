package ansible

import (
	"context"

	"github.com/apenella/go-ansible/pkg/playbook"
	"github.com/max-bytes/omnikeeper-deploy-agent/pkg/config"
	"github.com/sirupsen/logrus"
)

func Callout(ctx context.Context, config config.AnsibleCalloutConfig, id string, variableFile string, log *logrus.Logger) error {

	playbook := &playbook.AnsiblePlaybookCmd{
		Playbooks:         config.Playbooks,
		ConnectionOptions: config.ConnectionOptions,
		Options:           config.Options,
		Binary:            config.AnsibleBinary,
	}

	// overwrite/force set ansible variable host_id and host_variable_file
	playbook.Options.ExtraVars["host_id"] = id
	playbook.Options.ExtraVars["host_variable_file"] = variableFile

	finalCommand, err := playbook.Command()
	if err != nil {
		return err
	}
	log.Tracef("Calling playbook for item %s: %s", id, finalCommand)

	err = playbook.Run(ctx)
	if err != nil {
		return err
	}

	return nil
}
