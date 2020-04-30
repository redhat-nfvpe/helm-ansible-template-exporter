package parse

import (
	"github.com/redhat-nfvpe/helm-ansible-template-exporter/internal/pkg/helm"
	"github.com/sirupsen/logrus"
	"strings"
)

// RangeNode is the representation of a "range" statement.  This is a tailored version of text/template's BranchNode.
// The text/template package "if", "for" and "with" utilizing the BranchNode abstraction.  This makes sense, since the
// types are all very similar in Go Templating, and they are all output in very similar manners.  However, Jinja2 has
// greater requirements surrounding output of these Branch structures.  This abstraction is introduced to handle the
// "range" BranchNode.
//
// Currently, the implementation outputs range statements using Ansible for-statement equivalents.  For example, given
// the list-range input:
//
// {{ range list }}
//
// The translation is:
//
// {% for item_list in list %}
//
// Go Template language also supports range over maps.  For the given map-range input:
//
// {{ range $key, $value := someDict }}
//
// The translation is:
//
// {% for key, value in someDict.iterItems() %}
//
// Lastly, in the case of list-range input, Go Template language implies an iterator.  That is, you can access
// properties of the list using the member access operator ".".  For example:
//
// {{ range computerList }}
// {{ .ipAddress }}
// {{ end }}
//
// The RangeNode implementation addresses this by appending the iterator variable name.  The translation is:
//
// {% for item_computerList in computerList %}
// {{ item_computerList.ipAddress }}
// {% endfor %}
type RangeNode struct {
	NodeType
	Pos
	tr       *Tree
	Line     int            // The line number in the input. Deprecated: Kept for compatibility.
	Pipe     *RangePipeNode // The pipeline to be evaluated.
	List     *ListNode      // What to execute if the value is non-empty.
	ElseList *ListNode      // What to execute if the value is empty (nil if absent).
}

func (t *Tree) newRange(pos Pos, line int, pipe *RangePipeNode, list, elseList *ListNode) *RangeNode {
	return &RangeNode{tr: t, NodeType: NodeRange, Pos: pos, Line: line, Pipe: pipe, List: list, ElseList: elseList}
}

func (r *RangeNode) String() string {
	var sb strings.Builder
	r.writeTo(&sb)
	return sb.String()
}

func (r *RangeNode) RemoveVarPrefix(prefix string) {
	for _, v := range r.Pipe.Decl {
		for i := range v.Ident {
			v.Ident[i] = strings.TrimPrefix(v.Ident[i], prefix)
		}
	}
}

// GetRangeUseCaseType ... get difference cases for range flow
func (r *RangeNode) GetRangeUseCaseType() RangeUseCaseType {
	if len(r.Pipe.Decl) == 0 && len(r.Pipe.Cmds[0].Args) == 1 {
		return UseCaseNoVariables
	} else if len(r.Pipe.Decl) == 2 { //key value
		return UseCaseKeyValue
	} else if len(r.Pipe.Decl) == 1 && len(r.Pipe.Cmds[0].Args) == 1 { //$host and one value
		return UseCaseSingleValue
	} else if len(r.Pipe.Decl) == 0 && len(r.Pipe.Cmds[0].Args) == 1 && r.Pipe.Cmds[0].Args[0].String() == "tuple" {
		return UseCaseTuple
	}
	return UseCaseDefault
}

// RangeUseCaseType identifies the types of uses cases for range.
type RangeUseCaseType int

/*-------------------------------------------------------------------------------------------------------
  | INPUT                                  | LHS(vars) & RHS(cmds) |  OUTPUT                              |
   --------------------------------------------------------------------------------------------------------
  | *{{- range .Values.ingress.secrets }}   |       0 & 1          | {% for item_secrets in .Values.ingress.secrets }} |
  ----------------------------------------------------------------------------------------------------------
  | {{range $key, $value := ingress.annotations }}  | 2 & 1 | {% for $key, $value in ingress.annotations %}|
  -----------------------------------------------------------------------------------------------------------
  | {{- range $host := .Values.ingress.hosts }}  | 1 & 1 | {% for $host in .Values.ingress.hosts }}        |
  ----------------------------------------------------------------------------------------------------------
  | {{ range tuple "config1.toml" "config2.toml" "config3.toml" }} |0 & n |
  								{% for tuple "config1.toml" "config2.toml" "config3.toml" %}
*/
const (
	UseCaseDefault     RangeUseCaseType = iota //Unknown use cases
	UseCaseNoVariables                         // *{{- range .Values.ingress.secrets }}
	UseCaseKeyValue                            // {{range $key, $value := ingress.annotations }}
	UseCaseSingleValue                         //{{- range $host := .Values.ingress.hosts }}
	UseCaseTuple                               //range tuple "config1.toml" "config2.toml" "config3.toml" }}
)

func (r *RangeNode) writeTo(sb *strings.Builder) {
	var rangeValues *map[string][]*helm.LogHelmReport
	var dotValue = make(map[string][]*helm.LogHelmReport)
	dotValue["."] = []*helm.LogHelmReport{}
	var itemField string
	removeBodyVariablePrefix := false

	sb.WriteString("{% ")
	/*-------------------------------------------------------------------------------------------------------
	  | INPUT                                  | LHS(vars) & RHS(cmds) |  OUTPUT                              |
	   --------------------------------------------------------------------------------------------------------
	  | *{{- range .Values.ingress.secrets }}   |       0 & 1          | {% for item_secrets in .Values.ingress.secrets }} |
	  ----------------------------------------------------------------------------------------------------------
	  | {{range $key, $value := ingress.annotations }}  | 2 & 1 | {% for $key, $value in ingress.annotations %}|
	  -----------------------------------------------------------------------------------------------------------
	  | {{- range $host := .Values.ingress.hosts }}  | 1 & 1 | {% for $host in .Values.ingress.hosts }}        |
	  ----------------------------------------------------------------------------------------------------------
	  | {{ range tuple "config1.toml" "config2.toml" "config3.toml" }} |0 & n |
									{% for tuple "config1.toml" "config2.toml" "config3.toml" %}
	*/
	//a) if you have zero variables and one command argument then {{- range .Values.ingress.secrets }}
	////1. derive the item_name and prefix the variables under the cmds variable found  in values with $item_name
	//r) if you have two variables then assume  you have key and value don't have to do anything
	//c) if you have  one variable then
	// 1. derive the item_name and prefix the variables under the cmds variable found  in values with $item_name
	switch r.GetRangeUseCaseType() {
	case UseCaseNoVariables:
		{ //{{- range .Values.ingress.secrets }}
			//extract item from single argument
			sb.WriteString("for")
			sb.WriteByte(' ')
			ss := strings.Split(r.Pipe.Cmds[0].Args[0].String(), ".")
			itemField = "item_" + ss[len(ss)-1]
			sb.WriteString(itemField)
			sb.WriteByte(' ')
			sb.WriteString("in")
			sb.WriteByte(' ')
			r.Pipe.writeForTo(sb)
			rangeValues, _ = helm.GetValues(r.Pipe.Cmds[0].Args[0].String())
			logrus.Info("Attempting to prefix variables with loop variable,example: `value` becomes `item.value` for template", r.tr.Name)
			//attempting for dot values
			if rangeValues == nil {
				logrus.Warnf("Failed to prefix variables with loop variable value for template %s", r.tr.Name)
				logrus.Warnf("%s at position %d was not found in Helm chart's values: invalid path", r.Pipe.Cmds[0].Args[0].String(), r.Pos)
				rangeValues = &dotValue
			} else {
				(*rangeValues)["."] = []*helm.LogHelmReport{}
			}

		}
	case UseCaseKeyValue, UseCaseSingleValue:
		{
			sb.WriteString("for")
			sb.WriteByte(' ')
			r.RemoveVarPrefix("$")
			r.Pipe.writeForTo(sb)
			removeBodyVariablePrefix = true
		}
	case UseCaseTuple:
		{
			sb.WriteString("for")
			sb.WriteByte(' ')
			r.Pipe.writeForTo(sb)
		}
	default:
		{
			sb.WriteString("for ")
			r.Pipe.writeForTo(sb)
		}
	}

	sb.WriteString(" %}")
	r.List.writeTo(sb)
	// all the things if the conditional is true
	//prefix  range variables with item
	if rangeValues != nil {
		helm.PreFixValuesWithItems(sb, itemField, rangeValues)
		logrus.Infof("**************************************************************")
		logrus.Info("       for loop variables replacement inside for loop : 	", r.tr.Name)
		logrus.Infof("**************************************************************")
		logrus.Infof("For loop: Following variables were prefixed with %s in template %s", itemField, r.tr.Name)
		for k, v := range *rangeValues {
			helm.PrintReportItems(v)
			delete(*rangeValues, k)
		}
		logrus.Infof("**************************************************************")
	}
	// Clean $ prefixed variable
	if removeBodyVariablePrefix {
		helm.RemoveDollarPrefix(sb)
	}

	if r.ElseList != nil {
		sb.WriteString("{% else %}")
		r.ElseList.writeTo(sb)
	}
	sb.WriteString("{% endfor %}")
}

func (r *RangeNode) tree() *Tree {
	return r.tr
}

func (r *RangeNode) Copy() Node {
	return r.tr.newRange(r.Pos, r.Line, r.Pipe.CopyPipe(), r.List.CopyList(), r.ElseList.CopyList())
}