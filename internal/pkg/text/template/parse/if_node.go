package parse

import "strings"

// IfNode is the representation of a "if" statement.  This is a tailored version of text/template's BranchNode.
// The text/template package "if", "for" and "with" utilizing the BranchNode abstraction.  This makes sense, since the
// types are all very similar in Go Templating, and they are all output in very similar manners.  However, Jinja2 has
// greater requirements surrounding output of these Branch structures.  This abstraction is introduced to handle the
// "if" BranchNode.
//
// Currently, the implementation makes a best guess to fully translate "if" conditionals.  There is not always a
// deterministic conversion, which is discussed in much greater detail below.  However, this implementation makes a best
// attempt approach to resolve a proper translation.  There are two major difference between Go Template conditionals
// and Jinja2 conditionals:
//
// 1. If-Ambiguity:  In Go template language, a conditional such as "{{ if something }}" is overloaded and may mean:
//
//    "if something is true" (boolean evaluation) or "if something is defined" (definition check)
//
//    Jinja2 does not support this overloaded conditional evaluation behavior.  There is no deterministic way to tell
//    whether an arbitrary condition is expressing boolean evaluation or checking for definition without further
//    information.  Thus, a heuristic is introduced to make a best guess.  The input Helm Chart's Values.yaml file is
//    interrogated for the "something" key.  If "something" exists and its value is a boolean, then the conditional is
//    classified as boolean evaluation, and will output as:
//
//    {% if something %}
//
//    Otherwise, if the key is not found or the value is not a boolean, then the conditional is classified as a
//    definition check.  In this case, the Jinja2 output is:
//
//    {% if something is defined %}
//
//    Note:  You must uncomment optional configuration in order for this heuristic to work accurately.
//
// 2. Boolean-Composition: Go Template language treats boolean operators ("and", "or", "not") as function calls.  Thus,
//    the syntax is "<booleanOperator> <condition1> <condition2>".  Treating such operators as function invocations is
//    a common tactic in language development, as it significantly reduces the complexity of the parser.  However,
//    Jinja2 treats boolean operators in a more traditional way, and expects the syntax
//    "<condition1> <booleanOperator> <condition2>".  Additionally, Go template language implements an "eq" equality
//    operator, which works in a similar way to the Go template boolean operator implementation.  In these cases, the
//    Abstract Syntax Tree nodes must be re-ordered in order to output proper Jinja2.  The "swapping" of nodes in memory
//    is necessary since these statements can, and often are, heavily nested.
type IfNode struct {
	NodeType
	Pos
	tr       *Tree
	Line     int          // The line number in the input. Deprecated: Kept for compatibility.
	Pipe     *IfPipeNode // The pipeline to be evaluated.
	List     *ListNode   // What to execute if the value is non-empty.
	ElseList *ListNode   // What to execute if the value is empty (nil if absent).
}

func (n *IfNode) String() string {
	var sb strings.Builder
	n.writeTo(&sb)
	return sb.String()
}

func (n *IfNode) writeTo(sb *strings.Builder) {
    // This implementation has been greatly simplified from the original BranchNode.writeTo(*strings.Builder) impl:
	// https://github.com/golang/go/blob/master/src/text/template/parse/node.go#L822
	// This is largely due to the fact that the "IfNode" type is abstracted to handle "if" statements specifically.
	// Removal of the overloaded functionality allows greater specificity in outputting the Node.
	sb.WriteString("{% if ")
	n.Pipe.writeTo(sb)
	sb.WriteString(" %}")
	n.List.writeTo(sb)
	if n.ElseList != nil {
		sb.WriteString("{% else %}")
		n.ElseList.writeTo(sb)
	}
	sb.WriteString("{% endif %}")
}

func (t *Tree) newIf(pos Pos, line int, pipe *IfPipeNode, list, elseList *ListNode) *IfNode {
	return &IfNode{tr: t, NodeType: NodeIf, Pos: pos, Line: line, Pipe: pipe, List: list, ElseList: elseList}
}

func (n *IfNode) Copy() Node {
	return n.tr.newIf(n.Pos, n.Line, n.Pipe.CopyPipe(), n.List.CopyList(), n.ElseList.CopyList())
}

func (n *IfNode) tree() *Tree {
	return n.tr
}
