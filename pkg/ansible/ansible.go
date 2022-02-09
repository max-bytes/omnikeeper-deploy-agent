package ansible

import (
	"context"

	"github.com/apenella/go-ansible/pkg/playbook"
	"github.com/max-bytes/omnikeeper-deploy-agent/pkg/config"
)

func Callout(ctx context.Context, config config.AnsibleCalloutConfig, id string) error {

	playbook := &playbook.AnsiblePlaybookCmd{
		Playbooks:         config.Playbooks,
		ConnectionOptions: config.ConnectionOptions,
		Options:           config.Options,
		Binary:            config.AnsibleBinary,
	}

	// overwrite/force set ansible variable host_id
	playbook.Options.ExtraVars["host_id"] = id

	err := playbook.Run(ctx)
	if err != nil {
		return err
	}

	return nil
}
