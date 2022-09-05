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
	parser.ParseFromString("r.sub == p.sub && r.obj == p.obj && r.act == p.act")

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
		evaluable := parser.ParseFromString("r.sub == p.sub && r.obj == p.obj && r.act == p.act").(ast.Evaluable)
		ae := NewAbstractExpression(evaluable)
		assert.Equal(t, []string{"act", "obj", "sub"}, sortStrings(ae.AccessorMembers()))
	})
	t.Run("basic_with_root_model", func(t *testing.T) {
		evaluable := parser.ParseFromString("r.sub == p.sub && r.obj == p.obj && r.act == p.act || r.sub == \"root\"").(ast.Evaluable)
		ae := NewAbstractExpression(evaluable)
		assert.Equal(t, []string{"act", "obj", "sub"}, sortStrings(ae.AccessorMembers()))
	})
	t.Run("rbac_model", func(t *testing.T) {
		evaluable := parser.ParseFromString("g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act").(ast.Evaluable)
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
		evaluable := parser.ParseFromString("g(r.sub, p.sub) && g2(r.obj, p.obj) && r.act == p.act").(ast.Evaluable)
		ae := NewAbstractExpression(evaluable)
		assert.Equal(t, []string{"act", "obj", "sub"}, sortStrings(ae.AccessorMembers()))
	})
	t.Run("abac_model.conf", func(t *testing.T) {
		evaluable := parser.ParseFromString("r.sub == r.obj.Owner").(ast.Evaluable)
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
		evaluable := parser.ParseFromString("r.sub == p.sub && r.obj == p.obj && r.act == p.act").(ast.Evaluable)
		expectedPruned := parser.ParseFromString("r.sub == p.sub").(ast.Evaluable)
		expectedRemained := parser.ParseFromString("r.obj == p.obj && r.act == p.act").(ast.Evaluable)

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
		evaluable := parser.ParseFromString("r.sub == p.sub && r.obj == p.obj && r.act == p.act").(ast.Evaluable)
		expectedPruned := parser.ParseFromString("r.obj == p.obj").(ast.Evaluable)
		expectedRemained := parser.ParseFromString("r.sub == p.sub && r.act == p.act").(ast.Evaluable)

		// prunes the leftmost subtree's sibling
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
		evaluable := parser.ParseFromString("r.sub == p.sub && r.obj == p.obj && r.act == p.act").(ast.Evaluable)
		expectedPruned := parser.ParseFromString("r.act == p.act").(ast.Evaluable)
		expectedRemained := parser.ParseFromString("r.sub == p.sub && r.obj == p.obj").(ast.Evaluable)

		// prunes the rightmost subtree
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
		evaluable := parser.ParseFromString("r.sub == p.sub && r.obj == p.obj && r.act == p.act").(ast.Evaluable)
		expectedPruned := parser.ParseFromString("r.sub == p.sub && r.obj == p.obj").(ast.Evaluable)
		expectedRemained := parser.ParseFromString("r.act == p.act").(ast.Evaluable)

		// prunes the rightmost subtree's sibling

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
		evaluable := parser.ParseFromString("r.sub == p.sub && r.obj == p.obj && r.act == p.act || r.sub == \"root\"").(ast.Evaluable)
		expectedPruned := parser.ParseFromString("r.sub == \"root\"").(ast.Evaluable)
		expectedRemained := parser.ParseFromString("r.sub == p.sub && r.obj == p.obj && r.act == p.act").(ast.Evaluable)

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
		evaluable := parser.ParseFromString("r.sub == p.sub && r.obj == p.obj && r.act == p.act").(ast.Evaluable)

		pruned, remained := PruneSubtree(evaluable, func(evaluable ast.Evaluable) bool {
			members := GetAccessorMembers(evaluable)
			return len(members) == 3 && slices.Contains(members, "obj") && slices.Contains(members, "act")
		})
		assert.Equal(t, evaluable, pruned)
		assert.Equal(t, nil, remained)
	})
	t.Run("prunes nothing", func(t *testing.T) {
		evaluable := parser.ParseFromString("r.sub == p.sub && r.obj == p.obj && r.act == p.act").(ast.Evaluable)

		pruned, remained := PruneSubtree(evaluable, func(evaluable ast.Evaluable) bool {
			return false
		})
		assert.Equal(t, nil, pruned)
		assert.Equal(t, evaluable, remained)
	})

}

func TestFlatAndSubtree(t *testing.T) {
	t.Run("and expr0", func(t *testing.T) {
		tree := parser.MustParseFromString("A")
		expected := []ast.Evaluable{
			&ast.Primitive{Typ: ast.IDENTIFIER, Value: "A"},
		}
		result := FlatAndSubtree(tree)
		assert.Equal(t, expected, result)
	})
	t.Run("A && B && C && D", func(t *testing.T) {
		tree := parser.MustParseFromString("A && B && C && D")
		expected := []ast.Evaluable{
			&ast.Primitive{Typ: ast.IDENTIFIER, Value: "A"},
			&ast.Primitive{Typ: ast.IDENTIFIER, Value: "B"},
			&ast.Primitive{Typ: ast.IDENTIFIER, Value: "C"},
			&ast.Primitive{Typ: ast.IDENTIFIER, Value: "D"},
		}
		result := FlatAndSubtree(tree)
		slices.SortFunc(result, func(a, b ast.Evaluable) bool {
			return strings.Compare(a.(*ast.Primitive).Value.(string), b.(*ast.Primitive).Value.(string)) < 0
		})
		assert.Equal(t, expected, result)
	})
	t.Run("A && B && C && D && E", func(t *testing.T) {
		tree := parser.MustParseFromString("A && B && C && D && E")
		expected := []ast.Evaluable{
			&ast.Primitive{Typ: ast.IDENTIFIER, Value: "A"},
			&ast.Primitive{Typ: ast.IDENTIFIER, Value: "B"},
			&ast.Primitive{Typ: ast.IDENTIFIER, Value: "C"},
			&ast.Primitive{Typ: ast.IDENTIFIER, Value: "D"},
			&ast.Primitive{Typ: ast.IDENTIFIER, Value: "E"},
		}
		result := FlatAndSubtree(tree)
		slices.SortFunc(result, func(a, b ast.Evaluable) bool {
			return strings.Compare(a.(*ast.Primitive).Value.(string), b.(*ast.Primitive).Value.(string)) < 0
		})
		assert.Equal(t, expected, result)
	})
	t.Run("A && (B || D && E)", func(t *testing.T) {
		tree := parser.MustParseFromString("A && (B || D && E)")
		expected := []ast.Evaluable{
			parser.MustParseFromString("A"),
			parser.MustParseFromString("(B || D && E)"),
		}
		result := FlatAndSubtree(tree)
		assert.Equal(t, expected, result)
	})
	t.Run("(A || !B || !C) && (!D || E ||F)", func(t *testing.T) {
		tree := parser.MustParseFromString("(A || !B || !C) && (!D || E ||F)")
		expected := []ast.Evaluable{
			parser.MustParseFromString("(A || !B || !C)"),
			parser.MustParseFromString("(!D || E ||F)"),
		}
		result := FlatAndSubtree(tree)
		assert.Equal(t, expected, result)
	})
}

func TestFlatOrSubtree(t *testing.T) {
	t.Run("or expr0", func(t *testing.T) {
		tree := parser.MustParseFromString("A")
		expected := []ast.Evaluable{
			&ast.Primitive{Typ: ast.IDENTIFIER, Value: "A"},
		}
		result := FlatOrSubtree(tree)
		assert.Equal(t, expected, result)
	})
	t.Run("or expr1", func(t *testing.T) {
		tree := parser.MustParseFromString("A || B || C || D")
		expected := []ast.Evaluable{
			&ast.Primitive{Typ: ast.IDENTIFIER, Value: "A"},
			&ast.Primitive{Typ: ast.IDENTIFIER, Value: "B"},
			&ast.Primitive{Typ: ast.IDENTIFIER, Value: "C"},
			&ast.Primitive{Typ: ast.IDENTIFIER, Value: "D"},
		}
		result := FlatOrSubtree(tree)
		slices.SortFunc(result, func(a, b ast.Evaluable) bool {
			return strings.Compare(a.(*ast.Primitive).Value.(string), b.(*ast.Primitive).Value.(string)) < 0
		})
		assert.Equal(t, expected, result)
	})
	t.Run("or expr2", func(t *testing.T) {
		tree := parser.MustParseFromString("A || B || C || D || E")
		expected := []ast.Evaluable{
			&ast.Primitive{Typ: ast.IDENTIFIER, Value: "A"},
			&ast.Primitive{Typ: ast.IDENTIFIER, Value: "B"},
			&ast.Primitive{Typ: ast.IDENTIFIER, Value: "C"},
			&ast.Primitive{Typ: ast.IDENTIFIER, Value: "D"},
			&ast.Primitive{Typ: ast.IDENTIFIER, Value: "E"},
		}
		result := FlatOrSubtree(tree)
		slices.SortFunc(result, func(a, b ast.Evaluable) bool {
			return strings.Compare(a.(*ast.Primitive).Value.(string), b.(*ast.Primitive).Value.(string)) < 0
		})
		assert.Equal(t, expected, result)
	})
	t.Run("or expr3", func(t *testing.T) {
		tree := parser.MustParseFromString("A && B || C")
		expected := []ast.Evaluable{
			parser.MustParseFromString("A && B"),
			parser.MustParseFromString("C"),
		}
		result := FlatOrSubtree(tree)
		assert.Equal(t, expected, result)
	})
	t.Run("or expr3", func(t *testing.T) {
		tree := parser.MustParseFromString("(A && !B && !C) || (!D && E && F)")
		expected := []ast.Evaluable{
			parser.MustParseFromString("(A && !B && !C)"),
			parser.MustParseFromString("(!D && E && F)"),
		}
		result := FlatOrSubtree(tree)
		assert.Equal(t, expected, result)
	})
}
