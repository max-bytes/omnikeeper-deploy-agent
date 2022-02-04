package runner

import (
	"context"

	"github.com/shurcooL/graphql"
	"github.com/sirupsen/logrus"
)

type Processor interface {
	Process(ctx context.Context, okClient *graphql.Client, log *logrus.Logger) (map[string]interface{}, error)
}
