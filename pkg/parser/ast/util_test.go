package ast

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRemoveStringQuote(t *testing.T) {
	target := "\"should be string\""
	target2 := "'should be string'"
	assert.Equal(t, "should be string", RemoveStringQuote(target))
	assert.Equal(t, "should be string", RemoveStringQuote(target2))
}
