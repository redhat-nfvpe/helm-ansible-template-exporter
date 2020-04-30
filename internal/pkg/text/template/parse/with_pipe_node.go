package parse

import "strings"

// WithPipeNode holds a "with" pipeline (i.e., everything after "{{ with ").  IfPipeNode was abstracted from the generic
// PipeNode in order to hand tailor Jinja2 output, which varies greatly from other branch structures in the Jinja2
// template language.
type WithPipeNode struct {
	NodeType
	Pos
	tr       *Tree
	Line     int
	IsAssign bool
	Decl     []*VariableNode
	Cmds     []*CommandNode
}

func (p *WithPipeNode) append(command *CommandNode) {
	p.Cmds = append(p.Cmds, command)
}

func (p *WithPipeNode) String() string {
	var sb strings.Builder
	p.writeTo(&sb)
	return sb.String()
}

func (t *Tree) newWithPipeline(pos Pos, line int, vars []*VariableNode) *WithPipeNode {
	return &WithPipeNode{tr: t, NodeType: NodePipeWith, Pos: pos, Line: line, Decl: vars}
}

func (p *WithPipeNode) writeTo(sb *strings.Builder) {
	if len(p.Decl) > 0 {
		for i, v := range p.Decl {
			if i > 0 {
				sb.WriteString(", ")
			}
			v.writeTo(sb)
		}
		sb.WriteString(" := ")
	}
	for i, c := range p.Cmds {
		if i > 0 {
			sb.WriteString(" | ")
		}
		c.writeTo(sb)
	}
}

func (p *WithPipeNode) tree() *Tree {
	return p.tr
}

func (p *WithPipeNode) CopyPipe() *WithPipeNode {
	if p == nil {
		return p
	}
	vars := make([]*VariableNode, len(p.Decl))
	for i, d := range p.Decl {
		vars[i] = d.Copy().(*VariableNode)
	}
	n := p.tr.newWithPipeline(p.Pos, p.Line, vars)
	n.IsAssign = p.IsAssign
	for _, c := range p.Cmds {
		n.append(c.Copy().(*CommandNode))
	}
	return n
}

func (p *WithPipeNode) Copy() Node {
	return p.CopyPipe()
}
