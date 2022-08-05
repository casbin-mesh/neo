package plan

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
