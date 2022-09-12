package plan

type EffectType uint

const (
	AllowOverride EffectType = iota + 1
	DenyOverride
	AllowAndDeny
	Priority
	PriorityBaseOnRole
)

type MatcherPlan interface {
	AbstractPlan
	EffectType() EffectType
}

type matcherPlan struct {
	AbstractPlan
	effectType EffectType
}

func (m matcherPlan) EffectType() EffectType {
	return m.effectType
}

func NewMatcherPlan(children []AbstractPlan, effectType EffectType) MatcherPlan {
	return &matcherPlan{
		AbstractPlan: NewAbstractPlan(nil, children),
		effectType:   effectType,
	}
}
