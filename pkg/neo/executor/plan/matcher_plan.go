package plan

import (
	"fmt"
	"github.com/casbin-mesh/neo/pkg/neo/utils"
)

type EffectType uint

const (
	AllowOverride EffectType = iota
	DenyOverride
	AllowAndDeny
	Priority
	PriorityBaseOnRole
)

var eftPolicy2Str = []string{
	"AllowOverride",
	"DenyOverride",
	"AllowAndDeny",
	"Priority",
	"PriorityBaseOnRole",
}

func (e EffectType) String() string {
	return eftPolicy2Str[e]
}

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

func (m matcherPlan) String() string {
	childStr := make([]string, 0, len(m.GetChildren()))
	for _, child := range m.GetChildren() {
		childStr = append(childStr, child.String())
	}
	return utils.TreeFormat(fmt.Sprintf("MatcherPlan | Type: %s", m.effectType), childStr...)
}
