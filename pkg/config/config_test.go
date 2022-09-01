package config

import (
	"testing"

	"github.com/apenella/go-ansible/pkg/options"
	"github.com/apenella/go-ansible/pkg/playbook"
	"github.com/stretchr/testify/assert"
)

var inputConfig = `
log_level: Trace
username: omnikeeper-client-library-test
password: omnikeeper-client-library-test
omnikeeper_backend_url: "https://10.0.0.43:45455"
keycloak_client_id: landscape-omnikeeper
collect_interval_seconds: 60
healthcheck_threshold_seconds: 120
output_directory: ./output
ansible:
  ansiblebinary: ansible-playbook
  playbooks:
    - playbook.yml
  connectionoptions:
    connection: local
    privatekey: id_rsa
    user:       user
  options:
    inventory: target-host-a,
    extravars: 
      ansible_port: 2222
      host_id: H12312312
`

func TestConfigLoading(t *testing.T) {
	cfg := Configuration{}

	err := ReadConfigFromBytes([]byte(inputConfig), &cfg)

	if err != nil {
		t.Error(err)
	}

	ansibleConnectionOptions := &options.AnsibleConnectionOptions{
		Connection: "local",
		PrivateKey: "id_rsa",
		User:       "user",
	}
	ansiblePlaybookOptions := &playbook.AnsiblePlaybookOptions{
		Inventory: "target-host-a,",
		ExtraVars: map[string]interface{}{
			"ansible_port": 2222,
			"host_id":      "H12312312",
		},
	}
	playbook := AnsibleCalloutConfig{
		Playbooks:         []string{"playbook.yml"},
		ConnectionOptions: ansibleConnectionOptions,
		Options:           ansiblePlaybookOptions,
		AnsibleBinary:     "ansible-playbook",
	}

	assert.Equal(t, playbook, cfg.Ansible)
}
