package processors

import (
	"context"
	"fmt"

	"github.com/shurcooL/graphql"
	"github.com/sirupsen/logrus"
)

type TelegrafProcessor struct {
}

func (p TelegrafProcessor) Process(ctx context.Context, okClient *graphql.Client, log *logrus.Logger) (map[string]interface{}, error) {
	variables := map[string]interface{}{}
	var query = TelegrafAgentsQuery{}
	err := okClient.Query(ctx, &query, variables)
	if err != nil {
		return nil, fmt.Errorf("Error running GraphQL query: %w", err)
	}

	categoryMembers := query.TraitEntities.Tsa_cmdb__category.ByDataID.Entity.Category_members
	ret := make(map[string]interface{}, len(categoryMembers))
	for _, cm := range categoryMembers {
		hostCI := cm.RelatedCI.TraitEntity.Tsa_cmdb__host.Entity
		hostID := hostCI.Cmdb_id

		if hostID != "" {
			ret[hostID] = hostCI
		} else {
			log.Warningf("Category member (%s) of root CI does not appear to fulfill trait entity \"tsa_cmdb.host\", skipping", cm.RelatedCIID)
		}
	}

	return ret, nil
}

type TelegrafAgentsQuery struct {
	TraitEntities struct {
		Tsa_cmdb__category struct {
			ByDataID struct {
				Entity struct {
					Category_members []struct {
						RelatedCIID string `graphql:"relatedCIID"` // library does not support native GUIDs
						RelatedCI   struct {
							TraitEntity struct {
								Tsa_cmdb__host struct {
									Entity struct {
										Cmdb_id string `json:"cmdb_id"`
									}
								}
							}
						} `graphql:"relatedCI"`
					}
				}
			} `graphql:"byDataID(id:{cmdb_id:\"CAT12189590\"})"`
		}
	} `graphql:"traitEntities(layers: [\"tsa_cmdb\"])"`
}

// traitEntities(layers: ["tsa_cmdb"]) {
//     tsa_cmdb__category
//     {
//       byDataID(id: {cmdb_id: "CAT12189590"}){
//         entity{
//           category_members{
//             relatedCI{
//               traitEntity{
//                 tsa_cmdb__host{
//                   entity{
//                     cmdb_id
//                   }
//                 }
//               }
//             }
//           }
//         }
//       }
//     }
//   }
