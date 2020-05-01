package parse

import (
	"github.com/redhat-nfvpe/helm-ansible-template-exporter/internal/pkg/helm"
	"github.com/sirupsen/logrus"
	"strings"
)

// IfCommandNode holds a command (a pipeline inside an "if" statement).  Since
type IfCommandNode struct {
	NodeType
	Pos
	tr   *Tree
	Args []Node // Arguments in lexical order: Identifier, field, or constant.
}

func (t *Tree) newIfCommand(pos Pos) *IfCommandNode {
	return &IfCommandNode{tr: t, NodeType: NodeIfCommand, Pos: pos}
}

func (c *IfCommandNode) append(arg Node) {
	c.Args = append(c.Args, arg)
}

func (c *IfCommandNode) String() string {
	var sb strings.Builder
	c.writeTo(&sb)
	return sb.String()
}

func (c *IfCommandNode) writeTo(sb *strings.Builder) {
	// Handles problem #2 of if-conditional conversion;  the "boolean composition problem".
	if commandNodeInversionIsRequired(&c.Args) {
		positionInFile := c.Position()

		logrus.Infof("Found an Argument sequence at position %d that requires Jinja2 syntax normalization: %s",
			positionInFile, c.Args)
		invertCommandNodes(&c.Args)
		logrus.Infof("Conversion at position %d became %s", positionInFile, c.Args)
	}

	for i, arg := range c.Args {
		if i > 0 {
			sb.WriteByte(' ')
		}
		if arg, ok := arg.(*IfPipeNode); ok {
			sb.WriteByte('(')
			arg.writeTo(sb)
			sb.WriteByte(')')
			continue
		}
		if isValueNode(arg.String()) {
			writeValueNode(&arg, sb)
		} else {
			arg.writeTo(sb)
		}
	}
}

func (c *IfCommandNode) tree() *Tree {
	return c.tr
}

func (c *IfCommandNode) Copy() Node {
	if c == nil {
		return c
	}
	n := c.tr.newIfCommand(c.Pos)
	for _, c := range c.Args {
		n.append(c.Copy())
	}
	return n
}

// Determine whether we need to invert nodes of the subtree to output in a Jinja2 compliant way.
func commandNodeInversionIsRequired(args *[]Node) bool {
	// CommandNode abstractions are used as the conditional clause in "if" statements.  In order to process an Args
	// array for "if" containing "and", "or", or "eq" we must invert the first and second elements of the Args array.
	// In other words, ["and", "condition1", "condition2"] will become ["condition1", "and", "condition2"].  This
	// functionality determines whether inversion is necessary.
	argsArray := *args
	return len(argsArray) > 2 &&
		(argsArray[0].String() == "and" || argsArray[0].String() == "or" || argsArray[0].String() == "eq")
}

// Side-effects the input array in order to swap elements at indexes 0 and 1.  This method does not check array length
// and expects well formed input.
func invertCommandNodes(args *[]Node) {
	argsArray := *args
	temp := argsArray[0]
	argsArray[0] = argsArray[1]
	argsArray[1] = temp
}

func isValueNode(nodeString string) bool {
	return strings.HasPrefix(nodeString, ".Values.")
}

func removeValuesPrefix(fieldString string) string {
	return strings.ReplaceAll(fieldString, ".Values.", "")
}

func writeValueNode(fieldNodeRef *Node, sb *strings.Builder) {
	fieldNode := *fieldNodeRef
	fieldNodePosition := fieldNode.Position()
	fieldNodeString := fieldNode.String()
	logrus.Infof("Found a candidate for conversion: %s", fieldNodeString)
	unqualifiedName := removeValuesPrefix(fieldNodeString)
	fieldIsLikelyBoolean, err := helm.ArgIsLikelyBooleanYamlValue(unqualifiedName)
	if err != nil {
		logrus.Warnf("\"%s\" at position %d was not found in Helm chart's values: %s.  Defaulting to definition conversion",
			fieldNodeString, fieldNodePosition, err)
	}
	if fieldIsLikelyBoolean {
		logrus.Infof("Determined %s at position %d is likely a boolean", fieldNodeString, fieldNodePosition)
		sb.WriteString(unqualifiedName)
	} else {
		logrus.Infof("Determined %s at position %d is likely checking for definition, not boolean evaluation",
			fieldNodeString, fieldNodePosition)
		sb.WriteString(unqualifiedName)
		sb.WriteString(" is defined")
	}
}