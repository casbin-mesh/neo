package plan

import (
	"fmt"
	"github.com/casbin-mesh/neo/pkg/neo/utils"
)

type LimitPlan interface {
	AbstractPlan
	Limit() int
}

type limitPlan struct {
	AbstractPlan
	limit int
}

func (l limitPlan) Limit() int {
	return l.limit
}

func NewLimitPlan(children []AbstractPlan, limit int) LimitPlan {
	return &limitPlan{
		AbstractPlan: NewAbstractPlan(nil, children),
		limit:        limit,
	}
}

func (m limitPlan) String() string {
	childStr := make([]string, 0, len(m.GetChildren()))
	for _, child := range m.GetChildren() {
		childStr = append(childStr, child.String())
	}
	return utils.TreeFormat(fmt.Sprintf("LimitPlan | Limit:%d", m.limit), childStr...)
}
