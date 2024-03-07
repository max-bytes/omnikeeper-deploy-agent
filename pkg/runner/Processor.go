package runner

import (
	"context"

	"github.com/hasura/go-graphql-client"
	"github.com/sirupsen/logrus"
)

type ProcessResultItem struct {
	Success bool
	Logs    []string
}

type Processor interface {
	Process(configFile string, ctx context.Context, okClient *graphql.Client, log *logrus.Logger) (map[string]interface{}, error)
	PostProcess(configFile string, ctx context.Context, okClient *graphql.Client, results map[string]ProcessResultItem)
}
