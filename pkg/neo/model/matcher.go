package model

type MatcherInfo struct {
	ID           uint64
	Name         CIStr
	Raw          string
	EffectPolicy byte
}

func (m *MatcherInfo) Clone() *MatcherInfo {
	return &*m
}
