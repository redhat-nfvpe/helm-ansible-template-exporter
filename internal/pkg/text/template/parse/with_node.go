package parse

import (
	"github.com/sirupsen/logrus"
	"strings"
)

// WithNode is the representation of a "with" statement.  This is a tailored version of text/template's BranchNode.
// The text/template package "if", "for" and "with" utilizing the BranchNode abstraction.  This makes sense, since the
// types are all very similar in Go Templating, and they are all output in very similar manners.  However, Jinja2 has
// greater requirements surrounding output of these Branch structures. This abstraction is introduced to handle the
// "with" BranchNode.
//
// Currently, the implementation outputs go template code, as Ansible does not support the "with" block.  A warning
// including the source line number is emitted to the user that a manual conversion is required.  In the future, this
// should be improved to support a more automated form of conversion.
type WithNode struct {
	NodeType
	Pos
	tr       *Tree
	Line     int           // The line number in the input. Deprecated: Kept for compatibility.
	Pipe     *WithPipeNode // The *with* pipeline to be evaluated.
	List     *ListNode     // What to execute if the value is non-empty.
	ElseList *ListNode     // What to execute if the value is empty (nil if absent).
}

func (w *WithNode) String() string {
	var sb strings.Builder
	w.writeTo(&sb)
	return sb.String()
}

func (w *WithNode) writeTo(sb *strings.Builder) {
	// TODO:  This implementation will not work in Ansible.  This is not a regression;  no existing solution has
	// handled with appropriately.
	logrus.Warnf("\"with\" block on line %d outputting as in Go Template format;  manual conversion required",
		w.Line)
	sb.WriteString("{% with ")
	w.Pipe.writeTo(sb)
	sb.WriteString(" %}")
	w.List.writeTo(sb)
	if w.ElseList != nil {
		sb.WriteString("{% else %}")
		w.ElseList.writeTo(sb)
	}
	sb.WriteString("{% endwith %}")
}

func (t *Tree) newWith(pos Pos, line int, pipe *WithPipeNode, list, elseList *ListNode) *WithNode {
	return &WithNode{tr: t, NodeType: NodeWith, Pos: pos, Line: line, Pipe: pipe, List: list, ElseList: elseList}
}

func (w *WithNode) Copy() Node {
	return w.tr.newWith(w.Pos, w.Line, w.Pipe.CopyPipe(), w.List.CopyList(), w.ElseList.CopyList())
}

func (w *WithNode) tree() *Tree {
	return w.tr
}
