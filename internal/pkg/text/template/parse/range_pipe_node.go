package parse

import "strings"

// RangePipeNode holds an "range" pipeline (i.e., everything after "{{ range ").  IfPipeNode was abstracted from the
// generic  PipeNode in order to hand tailor Jinja2 output, which varies greatly from other branch structures in the
// Jinja2 template language.
type RangePipeNode struct {
	NodeType
	Pos
	tr       *Tree
	Line     int
	IsAssign bool
	Decl     []*VariableNode
	Cmds     []*CommandNode
}

func (p *RangePipeNode) append(command *CommandNode) {
	p.Cmds = append(p.Cmds, command)
}

func (p *RangePipeNode) String() string {
	var sb strings.Builder
	p.writeTo(&sb)
	return sb.String()
}

func (t *Tree) newRangePipeline(pos Pos, line int, vars []*VariableNode) *RangePipeNode {
	return &RangePipeNode{tr: t, NodeType: NodePipeRange, Pos: pos, Line: line, Decl: vars}
}

func (p *RangePipeNode) writeTo(sb *strings.Builder) {
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

func (p *RangePipeNode) tree() *Tree {
	return p.tr
}

func (p *RangePipeNode) CopyPipe() *RangePipeNode {
	if p == nil {
		return p
	}
	vars := make([]*VariableNode, len(p.Decl))
	for i, d := range p.Decl {
		vars[i] = d.Copy().(*VariableNode)
	}
	n := p.tr.newRangePipeline(p.Pos, p.Line, vars)
	n.IsAssign = p.IsAssign
	for _, c := range p.Cmds {
		n.append(c.Copy().(*CommandNode))
	}
	return n
}

func (p *RangePipeNode) Copy() Node {
	return p.CopyPipe()
}

// writeFor is used used to convert `:=` to `in`
func (p *RangePipeNode) writeForTo(sb *strings.Builder) {
	if len(p.Decl) > 0 {
		for i, v := range p.Decl {
			if i > 0 {
				sb.WriteString(", ")
			}
			v.writeTo(sb)
		}
		sb.WriteString(" in ")
	}
	for i, c := range p.Cmds {
		if i > 0 {
			sb.WriteString(" | ")
		}
		c.writeTo(sb)
	}
}
