package runner

import (
	"context"

	"github.com/hasura/go-graphql-client"
	"github.com/sirupsen/logrus"
)

type Processor interface {
	Process(configFile string, ctx context.Context, okClient *graphql.Client, log *logrus.Logger) (map[string]interface{}, error)
}
