package ansible

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/apenella/go-ansible/pkg/options"
	"github.com/apenella/go-ansible/pkg/playbook"
	"github.com/max-bytes/omnikeeper-deploy-agent/pkg/config"
	"github.com/sirupsen/logrus"
)

func TestCallout(t *testing.T) {
	ctx := context.Background()

	ansibleConnectionOptions := &options.AnsibleConnectionOptions{
		Connection: "local",
		PrivateKey: "/home/max/omnikeeper-deploy-agent-stack/keys/id_rsa",
		User:       "user",
	}

	ansiblePlaybookOptions := &playbook.AnsiblePlaybookOptions{
		Inventory: "target-host-a,",
		ExtraVars: map[string]interface{}{
			"ansible_port": "2222",
			"host_id":      "H12312312",
		},
	}

	cfg := config.AnsibleCalloutConfig{
		Playbooks:         []string{"playbook.yml"},
		Options:           ansiblePlaybookOptions,
		AnsibleBinary:     "/home/max/omnikeeper-deploy-agent-stack/ansible-playbook-wrapper",
		ConnectionOptions: ansibleConnectionOptions,
	}

	log := logrus.New()
	log.Out = ioutil.Discard
	err := Callout(ctx, cfg, "H12312312", "~/H12312312.json", log)
	t.Error(err)
}
