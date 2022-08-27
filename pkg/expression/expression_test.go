package expression

import (
	"github.com/casbin-mesh/neo/pkg/expression/ast"
	"github.com/casbin-mesh/neo/pkg/parser"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
	"sort"
	"strings"
	"testing"
)

func TestNewAbstractExpression(t *testing.T) {
	parser.ParseFormString("r.sub == p.sub && r.obj == p.obj && r.act == p.act")

}

func sortStrings(s []string) []string {
	sort.Slice(s, func(i, j int) bool {
		return strings.Compare(s[i], s[j]) < 0
	})
	return s
}

func TestAbstractExpression_AccessorMembers(t *testing.T) {
	t.Run("basic_model", func(t *testing.T) {
		// this test also covered:
		// basic_without_users_model
		// basic_without_resources_model
		evaluable := parser.ParseFormString("r.sub == p.sub && r.obj == p.obj && r.act == p.act").(ast.Evaluable)
		ae := NewAbstractExpression(evaluable)
		assert.Equal(t, []string{"act", "obj", "sub"}, sortStrings(ae.AccessorMembers()))
	})
	t.Run("basic_with_root_model", func(t *testing.T) {
		evaluable := parser.ParseFormString("r.sub == p.sub && r.obj == p.obj && r.act == p.act || r.sub == \"root\"").(ast.Evaluable)
		ae := NewAbstractExpression(evaluable)
		assert.Equal(t, []string{"act", "obj", "sub"}, sortStrings(ae.AccessorMembers()))
	})
	t.Run("rbac_model", func(t *testing.T) {
		evaluable := parser.ParseFormString("g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act").(ast.Evaluable)
		ae := NewAbstractExpression(evaluable)
		assert.Equal(t, []string{"act", "obj", "sub"}, sortStrings(ae.AccessorMembers()))
	})
	t.Run("rbac_with_resource_roles_model", func(t *testing.T) {
		// this test also covered:
		// keymatch_model
		// rbac_with_not_deny_model
		// rbac_with_deny_model
		// priority_model
		// priority_model_explicit
		// subject_priority_model
		evaluable := parser.ParseFormString("g(r.sub, p.sub) && g2(r.obj, p.obj) && r.act == p.act").(ast.Evaluable)
		ae := NewAbstractExpression(evaluable)
		assert.Equal(t, []string{"act", "obj", "sub"}, sortStrings(ae.AccessorMembers()))
	})
	t.Run("abac_model.conf", func(t *testing.T) {
		evaluable := parser.ParseFormString("r.sub == r.obj.Owner").(ast.Evaluable)
		ae := NewAbstractExpression(evaluable)
		assert.Equal(t, []string{"Owner", "obj", "sub"}, sortStrings(ae.AccessorMembers()))
	})
}

func TestAbstractExpression_Prune(t *testing.T) {
	t.Run("prunes the leftmost_subtree", func(t *testing.T) {
		// 			  			 AND
		// 					/	   	  \
		//  	  		 /        		  \
		// 			   AND	    	        EQ
		// 	   	  /  		 \     		  /    \
		//  	 EQ 	  	  EQ	   r.act   p.act
		//	   / 	 \   	/     \
		//	r.sub  p.sub  r.obj  p.obj
		evaluable := parser.ParseFormString("r.sub == p.sub && r.obj == p.obj && r.act == p.act").(ast.Evaluable)
		expectedPruned := parser.ParseFormString("r.sub == p.sub").(ast.Evaluable)
		expectedRemained := parser.ParseFormString("r.obj == p.obj && r.act == p.act").(ast.Evaluable)

		// prunes the leftmost subtree
		// remained expected:
		// 			  			 AND
		// 					/	   	   \
		//  	  		 /        		  \
		// 			   EQ	    	        EQ
		// 	   	    /  	  \     		  /    \
		//  	r.obj      p.obj       r.act   p.act
		pruned, remained := PruneSubtree(evaluable, func(evaluable ast.Evaluable) bool {
			members := GetAccessorMembers(evaluable)
			return len(members) == 1 && slices.Contains(members, "sub")
		})
		assert.Equal(t, expectedPruned, pruned)
		assert.Equal(t, expectedRemained, remained)
	})
	t.Run("prunes the leftmost subtree's sibling", func(t *testing.T) {
		// 			  			 AND
		// 					/	   	  \
		//  	  		 /        		  \
		// 			   AND	    	        EQ
		// 	   	  /  		 \     		  /    \
		//  	 EQ 	  	  EQ	   r.act   p.act
		//	   / 	 \   	/     \
		//	r.sub  p.sub  r.obj  p.obj
		evaluable := parser.ParseFormString("r.sub == p.sub && r.obj == p.obj && r.act == p.act").(ast.Evaluable)
		expectedPruned := parser.ParseFormString("r.obj == p.obj").(ast.Evaluable)
		expectedRemained := parser.ParseFormString("r.sub == p.sub && r.act == p.act").(ast.Evaluable)

		// prunes the leftmost subtree
		// remained expected:
		// 			  			 AND
		// 					/	   	   \
		//  	  		 /        		  \
		// 			   EQ	    	        EQ
		// 	   	    /  	  \     		  /    \
		//  	r.sub      p.sub       r.act   p.act
		pruned, remained := PruneSubtree(evaluable, func(evaluable ast.Evaluable) bool {
			members := GetAccessorMembers(evaluable)
			return len(members) == 1 && slices.Contains(members, "obj")
		})
		assert.Equal(t, expectedPruned, pruned)
		assert.Equal(t, expectedRemained, remained)
	})
	t.Run("prunes the rightmost subtree", func(t *testing.T) {
		// 			  			 AND
		// 					/	   	  \
		//  	  		 /        		  \
		// 			   AND	    	        EQ
		// 	   	  /  		 \     		  /    \
		//  	 EQ 	  	  EQ	   r.act   p.act
		//	   / 	 \   	/     \
		//	r.sub  p.sub  r.obj  p.obj
		evaluable := parser.ParseFormString("r.sub == p.sub && r.obj == p.obj && r.act == p.act").(ast.Evaluable)
		expectedPruned := parser.ParseFormString("r.act == p.act").(ast.Evaluable)
		expectedRemained := parser.ParseFormString("r.sub == p.sub && r.obj == p.obj").(ast.Evaluable)

		// prunes the leftmost subtree
		// remained expected:
		// 			  			 AND
		// 					/	   	   \
		//  	  		 /        		  \
		// 			   EQ	    	        EQ
		// 	   	    /  	  \     		  /    \
		//  	r.sub    p.sub 		   r.obj  p.obj
		pruned, remained := PruneSubtree(evaluable, func(evaluable ast.Evaluable) bool {
			members := GetAccessorMembers(evaluable)
			return len(members) == 1 && slices.Contains(members, "act")
		})
		assert.Equal(t, expectedPruned, pruned)
		assert.Equal(t, expectedRemained, remained)
	})
	t.Run("prunes the rightmost subtree's sibling", func(t *testing.T) {
		// 			  			 AND
		// 					/	   	  \
		//  	  		 /        		  \
		// 			   AND	    	        EQ
		// 	   	  /  		 \     		  /    \
		//  	 EQ 	  	  EQ	   r.act   p.act
		//	   / 	 \   	/     \
		//	r.sub  p.sub  r.obj  p.obj
		evaluable := parser.ParseFormString("r.sub == p.sub && r.obj == p.obj && r.act == p.act").(ast.Evaluable)
		expectedPruned := parser.ParseFormString("r.sub == p.sub && r.obj == p.obj").(ast.Evaluable)
		expectedRemained := parser.ParseFormString("r.act == p.act").(ast.Evaluable)

		// prunes the leftmost subtree
		// remained expected:
		// 			   EQ
		// 	   	    /  	  \
		//  	 r.act   p.act
		pruned, remained := PruneSubtree(evaluable, func(evaluable ast.Evaluable) bool {
			members := GetAccessorMembers(evaluable)
			return len(members) == 2 && slices.Contains(members, "sub") && slices.Contains(members, "obj")
		})
		assert.Equal(t, expectedPruned, pruned)
		assert.Equal(t, expectedRemained, remained)
	})
	t.Run("prunes a subtree contains a constant primitive", func(t *testing.T) {
		evaluable := parser.ParseFormString("r.sub == p.sub && r.obj == p.obj && r.act == p.act || r.sub == \"root\"").(ast.Evaluable)
		expectedPruned := parser.ParseFormString("r.sub == \"root\"").(ast.Evaluable)
		expectedRemained := parser.ParseFormString("r.sub == p.sub && r.obj == p.obj && r.act == p.act").(ast.Evaluable)

		pruned, remained := PruneSubtree(evaluable, func(evaluable ast.Evaluable) bool {
			be := evaluable.(*ast.BinaryOperationExpr)
			p, ok := be.R.(*ast.Primitive)
			if !ok {
				return false
			}
			members := GetAccessorMembers(evaluable)
			return len(members) == 1 && p.Typ == ast.STRING
		})
		assert.Equal(t, expectedPruned, pruned)
		assert.Equal(t, expectedRemained, remained)
	})
	t.Run("prunes the whole tree", func(t *testing.T) {
		evaluable := parser.ParseFormString("r.sub == p.sub && r.obj == p.obj && r.act == p.act").(ast.Evaluable)

		pruned, remained := PruneSubtree(evaluable, func(evaluable ast.Evaluable) bool {
			members := GetAccessorMembers(evaluable)
			return len(members) == 3 && slices.Contains(members, "obj") && slices.Contains(members, "act")
		})
		assert.Equal(t, evaluable, pruned)
		assert.Equal(t, nil, remained)
	})
	t.Run("prunes nothing", func(t *testing.T) {
		evaluable := parser.ParseFormString("r.sub == p.sub && r.obj == p.obj && r.act == p.act").(ast.Evaluable)

		pruned, remained := PruneSubtree(evaluable, func(evaluable ast.Evaluable) bool {
			return false
		})
		assert.Equal(t, nil, pruned)
		assert.Equal(t, evaluable, remained)
	})

}
