package main

import (
	"context"
	"flag"

	"github.com/hasura/go-graphql-client"
	"github.com/max-bytes/omnikeeper-deploy-agent/pkg/runner"
	"github.com/sirupsen/logrus"
)

var (
	log        logrus.Logger
	version    = "0.0.0-src"
	configFile = flag.String("config", "config.yml", "Config file location")
)

func init() {
	log = *logrus.StandardLogger()
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetLevel(logrus.TraceLevel) // is overwritten by configuration below
}

func main() {
	flag.Parse()

	log.Infof("omnikeeper-deploy-agent-sample (Version: %s)", version)

	runner.Run(SampleAppProcessor{}, *configFile, &log)

	log.Infof("Stopping omnikeeper-deploy-agent-sample (Version: %s)", version)
}

type SampleAppProcessor struct {
}

func (p SampleAppProcessor) Process(configFile string, ctx context.Context, okClient *graphql.Client, log *logrus.Logger) (map[string]interface{}, error) {
	return map[string]interface{}{"test": ItemOutput{Name: "foo"}, "test2": ItemOutput{Name: "bar"}}, nil
}

// func (p SampleAppProcessor) Process(configFile string, ctx context.Context, okClient *graphql.Client, log *logrus.Logger) (map[string]interface{}, error) {
// 	variables := map[string]interface{}{}
// 	var query = SampleAppQuery{}
// 	err := okClient.Query(ctx, &query, variables)
// 	if err != nil {
// 		return nil, fmt.Errorf("Error running GraphQL query: %w", err)
// 	}
// 	namedCIs := query.TraitEntities.Named.All
// 	ret := make(map[string]interface{}, len(namedCIs))
// 	for _, nci := range namedCIs {
// 		inputHostCI := nci.Entity
// 		ciid := nci.Ciid

// 		ret[ciid] = ItemOutput{
// 			Name: inputHostCI.Name,
// 		}
// 	}

// 	return ret, nil
// }

func (p SampleAppProcessor) PostProcess(configFile string, ctx context.Context, okClient *graphql.Client, results map[string]runner.ProcessResultItem) {
	println(results)
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
