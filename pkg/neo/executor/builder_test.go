package executor

import (
	"github.com/casbin-mesh/neo/pkg/neo/executor/plan"
	"github.com/stretchr/testify/assert"
	"testing"
)

type BuildSet struct {
	input plan.AbstractPlan
	err   error
	exec  Executor
}

func TestExecutorBuilder_InvalidInput(t *testing.T) {
	Sets := []BuildSet{
		{
			input: plan.NewLimitPlan(nil, 10),
			err:   ErrMissChildPlan,
		},
		{
			input: plan.NewDeletePlan(nil, 1, 1),
			err:   ErrMissChildPlan,
		},
		{
			input: plan.NewDeletePlan(nil, 1, 1),
			err:   ErrMissChildPlan,
		},
		{
			input: plan.NewUpdatePlan(nil, 1, 1, nil),
			err:   ErrMissChildPlan,
		},
	}
	builder := NewExecutorBuilder(nil)
	for _, set := range Sets {
		exec, err := builder.Build(set.input), builder.Error()
		assert.Equal(t, set.exec, exec)
		assert.Equal(t, set.err, err)
		builder.ResetError()
	}
}
