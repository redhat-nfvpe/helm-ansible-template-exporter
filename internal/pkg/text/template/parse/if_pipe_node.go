package parse

import "strings"

// IfPipeNode holds an "if" pipeline (i.e., everything after "{{ if ").  IfPipeNode was abstracted from the generic
// PipeNode in order to hand tailor Jinja2 output, which varies greatly from other branch structures in the Jinja2
// template language.
type IfPipeNode struct {
	NodeType
	Pos
	tr       *Tree
	Line     int
	IsAssign bool
	Decl     []*VariableNode  // Kept for backwards compatibility.  Since PipeNode was once overloaded to mean several
	                          // things, declarations were optional (i.e., "range").
	Cmds     []*IfCommandNode // Commands for "if" statements are now of type IfCommandNode, which has greater
	                          // adequate functionality to output in a Jinja2 compliant manner.  (i.e., it will solve
	                          // the if-ambiguity and boolean-composition problems.
}

func (p *IfPipeNode) append(command *IfCommandNode) {
	p.Cmds = append(p.Cmds, command)
}

func (p *IfPipeNode) String() string {
	var sb strings.Builder
	p.writeTo(&sb)
	return sb.String()
}

func (t *Tree) newIfPipeline(pos Pos, line int, vars []*VariableNode) *IfPipeNode {
	return &IfPipeNode{tr: t, NodeType: NodePipeIf, Pos: pos, Line: line, Decl: vars}
}

func (p *IfPipeNode) writeTo(sb *strings.Builder) {
	for i, c := range p.Cmds {
		if i > 0 {
			sb.WriteString(" | ")
		}
		c.writeTo(sb)
	}
}

func (p *IfPipeNode) tree() *Tree {
	return p.tr
}

func (p *IfPipeNode) CopyPipe() *IfPipeNode {
	if p == nil {
		return p
	}
	vars := make([]*VariableNode, len(p.Decl))
	for i, d := range p.Decl {
		vars[i] = d.Copy().(*VariableNode)
	}
	n := p.tr.newIfPipeline(p.Pos, p.Line, vars)
	n.IsAssign = p.IsAssign
	for _, c := range p.Cmds {
		n.append(c.Copy().(*IfCommandNode))
	}
	return n
}

func (p *IfPipeNode) Copy() Node {
	return p.CopyPipe()
}
