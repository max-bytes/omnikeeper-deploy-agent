package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/max-bytes/omnikeeper-deploy-agent/pkg/runner"
	"github.com/shurcooL/graphql"
	"github.com/sirupsen/logrus"
)

var (
	version    = "0.0.0-src"
	configFile = flag.String("config", "config.yml", "Config file location")
)

func main() {
	flag.Parse()

	runner.Run(SampleAppProcessor{}, version, *configFile)
}

type SampleAppProcessor struct {
}

func (p SampleAppProcessor) Process(configFile string, ctx context.Context, okClient *graphql.Client, log *logrus.Logger) (map[string]interface{}, error) {
	variables := map[string]interface{}{}
	var query = SampleAppQuery{}
	err := okClient.Query(ctx, &query, variables)
	if err != nil {
		return nil, fmt.Errorf("Error running GraphQL query: %w", err)
	}

	namedCIs := query.TraitEntities.Named.All
	ret := make(map[string]interface{}, len(namedCIs))
	for _, nci := range namedCIs {
		inputHostCI := nci.Entity
		ciid := nci.Ciid

		ret[ciid] = ItemOutput{
			Name: inputHostCI.Name,
		}
	}

	return ret, nil
}

type ItemOutput struct {
	Name string `json:"name"`
}

type SampleAppQuery struct {
	TraitEntities struct {
		Named struct {
			All []struct {
				Ciid   string
				Entity struct {
					Name string
				}
			}
		}
	} `graphql:"traitEntities(layers: [\"__okconfig\"])"`
}
