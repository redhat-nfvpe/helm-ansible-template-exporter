// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Parse nodes fork.  The intent of this implementation is to hijack the existing Abstract Syntax Tree implementation,
// and in particular "Node.writeTo(...)" implementation, in order to output valid Jinja2 Syntax.  This particular
// implementation is targeted for conversion from Helm Chart to Ansible Role, and is not a targeted Go Template to
// Jinja2 conversion utility.  A better solution may have abstracted an enhanced Parse Tree, but that was not timely for
// this particular task.  Thus, the existing Parse Tree is utilized in a best attempt to convert to valid Jinja2 syntax,
// which Ansible Playbook is capable of templating.  This solution is not all encompassing, and there are certainly
// cases that will require by-hand modifications.

package parse

import (
	"fmt"
	"github.com/redhat-nfvpe/helm-ansible-template-exporter/internal/pkg/helm"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

var textFormat = "%s" // Changed to "%q" in tests for better error messages.

// A Node is an element in the parse tree. The interface is trivial.
// The interface contains an unexported method so that only
// types local to this package can satisfy it.
type Node interface {
	Type() NodeType
	String() string
	// Copy does a deep copy of the Node and all its components.
	// To avoid type assertions, some XxxNodes also have specialized
	// CopyXxx methods that return *XxxNode.
	Copy() Node
	Position() Pos // byte position of start of node in full original input string
	// tree returns the containing *Tree.
	// It is unexported so all implementations of Node are in this package.
	tree() *Tree
	// writeTo writes the String output to the builder.
	writeTo(*strings.Builder, *j2Context)
}

// NodeType identifies the type of a parse tree node.
type NodeType int

// Pos represents a byte position in the original input text from which
// this template was parsed.
type Pos int

func (p Pos) Position() Pos {
	return p
}

// Type returns itself andq provides an easy default implementation
// for embedding in a Node. Embedded in all non-trivial Nodes.
func (t NodeType) Type() NodeType {
	return t
}

const (
	NodeText       NodeType = iota // Plain text.
	NodeAction                     // A non-control action such as a field evaluation.
	NodeBool                       // A boolean constant.
	NodeChain                      // A sequence of field accesses.
	NodeCommand                    // An element of a pipeline.
	NodeDot                        // The cursor, dot.
	nodeElse                       // An else action. Not added to tree.
	nodeEnd                        // An end action. Not added to tree.
	NodeField                      // A field or method name.
	NodeIdentifier                 // An identifier; always a function name.
	NodeIf                         // An if action.
	NodeList                       // A list of Nodes.
	NodeNil                        // An untyped nil constant.
	NodeNumber                     // A numerical constant.
	NodePipe                       // A pipeline of commands.
	NodeRange                      // A range action.
	NodeString                     // A string constant.
	NodeTemplate                   // A template invocation action.
	NodeVariable                   // A $ variable.
	NodeWith                       // A with action.
)

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

// Nodes.

// ListNode holds a sequence of nodes.
type ListNode struct {
	NodeType
	Pos
	tr    *Tree
	Nodes []Node // The element nodes in lexical order.
}

func (t *Tree) newList(pos Pos) *ListNode {
	return &ListNode{tr: t, NodeType: NodeList, Pos: pos}
}

func (l *ListNode) append(n Node) {
	l.Nodes = append(l.Nodes, n)
}

func (l *ListNode) tree() *Tree {
	return l.tr
}

func (l *ListNode) String() string {
	var sb strings.Builder
	l.writeTo(&sb, &j2Context{})
	return sb.String()
}

func (l *ListNode) writeTo(sb *strings.Builder, context *j2Context) {
	for _, n := range l.Nodes {
		n.writeTo(sb, context)
	}
}

func (l *ListNode) CopyList() *ListNode {
	if l == nil {
		return l
	}
	n := l.tr.newList(l.Pos)
	for _, elem := range l.Nodes {
		n.append(elem.Copy())
	}
	return n
}

func (l *ListNode) Copy() Node {
	return l.CopyList()
}

// TextNode holds plain text.
type TextNode struct {
	NodeType
	Pos
	tr   *Tree
	Text []byte // The text; may span newlines.
}

func (t *Tree) newText(pos Pos, text string) *TextNode {
	return &TextNode{tr: t, NodeType: NodeText, Pos: pos, Text: []byte(text)}
}

func (t *TextNode) String() string {
	return fmt.Sprintf(textFormat, t.Text)
}

func (t *TextNode) writeTo(sb *strings.Builder, context *j2Context) {
	sb.WriteString(t.String())
}

func (t *TextNode) tree() *Tree {
	return t.tr
}

func (t *TextNode) Copy() Node {
	return &TextNode{tr: t.tr, NodeType: NodeText, Pos: t.Pos, Text: append([]byte{}, t.Text...)}
}

// PipeNode holds a pipeline with optional declaration
type PipeNode struct {
	NodeType
	Pos
	tr       *Tree
	Line     int             // The line number in the input. Deprecated: Kept for compatibility.
	IsAssign bool            // The variables are being assigned, not declared.
	Decl     []*VariableNode // Variables in lexical order.
	Cmds     []*CommandNode  // The commands in lexical order.
}

func (t *Tree) newPipeline(pos Pos, line int, vars []*VariableNode) *PipeNode {
	return &PipeNode{tr: t, NodeType: NodePipe, Pos: pos, Line: line, Decl: vars}
}

func (p *PipeNode) append(command *CommandNode) {
	p.Cmds = append(p.Cmds, command)
}

func (p *PipeNode) String() string {
	var sb strings.Builder
	p.writeTo(&sb, &j2Context{})
	return sb.String()
}

// The Go Template Parser forms an Abstract Syntax Tree that is particular to the Go Template Language specification.
// In the context of this project, the text/template node.go writeTo(...) implementation has been hijacked to output
// valid Jinja2 syntax instead of Go Template Parser Syntax.  Go Template Parser reuses several Node implementations
// for truly orthogonal concepts.  For example, CommandNode is used to represent if-conditional statements as well as
// Go template function invocations.  While this is a truly admirable and brilliant aspect of the Go Template language
// (i.e., keep it simple), Jinja2 is much more strict in terms of syntax.  As such, it is impossible to represent
// Jinja2 using the existing Parse tree short of injecting additional information from time to time, or completely
// modifying the Parse Tree with more granular Node type definitions (i.e., IfCommandNode and
// FunctionInvocationCommandNode).  Although a better solution would likely involve the latter, the former was chosen
// for this project in the interest of expediting a solution.  As such, the Node.writeTo(...) signature was modified to
// include a pointer to a j2Context, which is meant to represent a way of passing enhanced information down to child
// Nodes.  Child nodes can utilize this information to determine a strategy for generating Node output.
type j2Context struct {
	isConditional bool // Represents whether or not we are dealing with a conditional context.
	isFunc        bool // Represents whether or not we are dealing with a function context.
	pipeNodeCount int  // Assuming the above is true, this stores the current level of nesting.  Function invocations
	                   // can be and often are highly nested.  This context clue helps to determine whether the given
	                   // context may be a direct function invocation, which then needs to be converted to piped.
}

func (p *PipeNode) writeTo(sb *strings.Builder, context *j2Context) {
	if len(p.Decl) > 0 {
		for i, v := range p.Decl {
			if i > 0 {
				sb.WriteString(", ")
			}
			v.writeTo(sb, context)
		}
		sb.WriteString(" := ")
	}
	for i, c := range p.Cmds {
		if i > 0 {
			sb.WriteString(" | ")
		}
		// Persist the isConditional context, as it matters here.
		isConditional := context.isConditional
		injectedContext := j2Context{
			isConditional: isConditional,
			isFunc:        !isConditional,
			pipeNodeCount: i,
		}
		c.writeTo(sb, &injectedContext)
	}
}

//writeFor is used used by  loop overriding writeFor to change `:=` to `in`
func (p *PipeNode) writeForTo(sb *strings.Builder) {
	if len(p.Decl) > 0 {
		for i, v := range p.Decl {
			if i > 0 {
				sb.WriteString(", ")
			}
			v.writeTo(sb, &j2Context{})
		}
		sb.WriteString(" in ")
	}
	for i, c := range p.Cmds {
		if i > 0 {
			sb.WriteString(" | ")
		}
		c.writeTo(sb, &j2Context{})
	}
}

func (p *PipeNode) tree() *Tree {
	return p.tr
}

func (p *PipeNode) CopyPipe() *PipeNode {
	if p == nil {
		return p
	}
	vars := make([]*VariableNode, len(p.Decl))
	for i, d := range p.Decl {
		vars[i] = d.Copy().(*VariableNode)
	}
	n := p.tr.newPipeline(p.Pos, p.Line, vars)
	n.IsAssign = p.IsAssign
	for _, c := range p.Cmds {
		n.append(c.Copy().(*CommandNode))
	}
	return n
}

func (p *PipeNode) Copy() Node {
	return p.CopyPipe()
}

// ActionNode holds an action (something bounded by delimiters).
// Control actions have their own nodes; ActionNode represents simple
// ones such as field evaluations and parenthesized pipelines.
type ActionNode struct {
	NodeType
	Pos
	tr   *Tree
	Line int       // The line number in the input. Deprecated: Kept for compatibility.
	Pipe *PipeNode // The pipeline in the action.
}

func (t *Tree) newAction(pos Pos, line int, pipe *PipeNode) *ActionNode {
	return &ActionNode{tr: t, NodeType: NodeAction, Pos: pos, Line: line, Pipe: pipe}
}

func (a *ActionNode) String() string {
	var sb strings.Builder
	a.writeTo(&sb, &j2Context{})
	return sb.String()
}

func (a *ActionNode) writeTo(sb *strings.Builder, context *j2Context) {
	sb.WriteString("{{ ")
	a.Pipe.writeTo(sb, context)
	sb.WriteString(" }}")
}

func (a *ActionNode) tree() *Tree {
	return a.tr
}

func (a *ActionNode) Copy() Node {
	return a.tr.newAction(a.Pos, a.Line, a.Pipe.CopyPipe())

}

// CommandNode holds a command (a pipeline inside an evaluating action).
type CommandNode struct {
	NodeType
	Pos
	tr   *Tree
	Args []Node // Arguments in lexical order: Identifier, field, or constant.
}

func (t *Tree) newCommand(pos Pos) *CommandNode {
	return &CommandNode{tr: t, NodeType: NodeCommand, Pos: pos}
}

func (c *CommandNode) append(arg Node) {
	c.Args = append(c.Args, arg)
}

func (c *CommandNode) String() string {
	var sb strings.Builder
	c.writeTo(&sb, &j2Context{})
	return sb.String()
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
		// output on line #325-326
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

// Determines whether the go template is likely representative of direct function invocation.  For example, consider
// "{{ toYaml .Values.someYamlVariable }}".  Jinja2 is unable to render direct function invocation, so when rendering
// the Jinja2 translation, additional steps must be taken to re-arrange nodes in the Abstract Syntax Tree.  This
// function just determines whether the given context is representative of direct function invocation.
func isCandidateForDirectFunctionInvocation(argsPointer *[]Node, context  *j2Context) bool {
	args := *argsPointer
	numArgs := len(args)
	pipeNodeDepth := (*context).pipeNodeCount
	// pipeNodeDepth represents the functional nesting level.  Golang only allows for direct function invocation at
	// level 0, so ensure that we are dealing with a level 0 context.
	if pipeNodeDepth == 0 && numArgs > 1 {
		if _, ok := args[0].(*IdentifierNode); ok {
			ctx := *context
			if ctx.isFunc {
				logrus.Infof("Found a direct function invocation which must be translated to a pipe equivalent: %s", args)
				return true
			}
		}
	}
	return false
}

// Rewrite direct function invocation utilizing a pipe, to adhere to Jinija2 requirement which requires Ansible Filters
// to utilize piping.  For example:
// {{ toYaml .Values.someValue '.' }}
// becomes:
// {{ .Values.someValue | toYaml('.') }}
func writePipedVersionOfDirectFunctionInvocation(sb *strings.Builder, argsPointer *[]Node) {
	args := *argsPointer
	argsLen := len(args)
	if argsLen >= 2 {
		// direct function call with arguments
		firstArgument := args[1]
		functionName := args[0]
		// TODO:  I could not utilize just "b" then invoke sb.WriteString(b.String()).  Fix this later.
		var b strings.Builder
		sb.WriteString(firstArgument.String())
		b.WriteString(firstArgument.String())
		sb.WriteString(" | ")
		b.WriteString(" | ")
		sb.WriteString(functionName.String())
		b.WriteString(functionName.String())

		if argsLen > 2 {
			sb.WriteString("(")
			b.WriteString("(")
			for i := 2; i < len(args); i++ {
				if i > 2 {
					sb.WriteString(", ")
					b.WriteString(", ")
				}
				sb.WriteString(args[i].String())
				b.WriteString(args[i].String())
			}
			sb.WriteString(")")
			b.WriteString(")")
			logrus.Infof("Template Function Converted: %s %s", args, b.String())
		}
	} else {
		// direct function call without arguments
		sb.WriteString("'' | ")
		sb.WriteString(args[0].String())
		logrus.Infof("Template Function Converted: %s %s", args, "'' | " + args[0].String())
	}
}

// GoLang Templates allow specification of arguments without parentheses or comma separated values.  Jinja2 templates
// do not have this flexibility.  Thus, format the output for valid Jinja2 syntax.
func rewritePipedFunctionOutput(sb *strings.Builder, argsPointer *[]Node) {
	args := *argsPointer
	argsLen := len(args)
	sb.WriteString(args[0].String())
	// there are function arguments
	if argsLen > 1 {
		sb.WriteString("(")
		for i := 1; i < argsLen; i++ {
			if i > 1 {
				sb.WriteString(", ")
			}
			sb.WriteString(args[i].String())
		}
		sb.WriteString(")")
	}
}

func (c *CommandNode) writeTo(sb *strings.Builder, context *j2Context) {
	// Handles problem #2 of if-conditional conversion;  the "boolean composition problem".
	if commandNodeInversionIsRequired(&c.Args) {
		positionInFile := c.Position()

		logrus.Infof("Found an Argument sequence at position %d that requires Jinja2 syntax normalization: %s",
			positionInFile, c.Args)
		invertCommandNodes(&c.Args)
		logrus.Infof("Conversion at position %d became %s", positionInFile, c.Args)
	}

	// Such as: "{{ toYaml .Values.something '.' }}
	if isCandidateForDirectFunctionInvocation(&c.Args, context) {
		writePipedVersionOfDirectFunctionInvocation(sb, &c.Args)
		return
	} else {
		isFunc := (*context).isFunc
		// rewrite function output in the form func(arg1, arg2, ... argN)
		if isFunc {
			rewritePipedFunctionOutput(sb, &c.Args)
			return
		}

		for i, arg := range c.Args {
			if i > 0 {
				sb.WriteByte(' ')
			}
			if arg, ok := arg.(*PipeNode); ok {
				sb.WriteByte('(')
				arg.writeTo(sb, context)
				sb.WriteByte(')')
				continue
			}
			if context.isConditional && isValueNode(arg.String()) {
				writeValueNode(&arg, sb)
			} else {
				arg.writeTo(sb, context)
			}
		}
	}
}

func (c *CommandNode) tree() *Tree {
	return c.tr
}

func (c *CommandNode) Copy() Node {
	if c == nil {
		return c
	}
	n := c.tr.newCommand(c.Pos)
	for _, c := range c.Args {
		n.append(c.Copy())
	}
	return n
}

// IdentifierNode holds an identifier.
type IdentifierNode struct {
	NodeType
	Pos
	tr    *Tree
	Ident string // The identifier's name.
}

// NewIdentifier returns a new IdentifierNode with the given identifier name.
func NewIdentifier(ident string) *IdentifierNode {
	return &IdentifierNode{NodeType: NodeIdentifier, Ident: ident}
}

// SetPos sets the position. NewIdentifier is a public method so we can't modify its signature.
// Chained for convenience.
// TODO: fix one day?
func (i *IdentifierNode) SetPos(pos Pos) *IdentifierNode {
	i.Pos = pos
	return i
}

// SetTree sets the parent tree for the node. NewIdentifier is a public method so we can't modify its signature.
// Chained for convenience.
// TODO: fix one day?
func (i *IdentifierNode) SetTree(t *Tree) *IdentifierNode {
	i.tr = t
	return i
}

func (i *IdentifierNode) String() string {
	return i.Ident
}

func (i *IdentifierNode) writeTo(sb *strings.Builder, context *j2Context) {
	sb.WriteString(i.String())
}

func (i *IdentifierNode) tree() *Tree {
	return i.tr
}

func (i *IdentifierNode) Copy() Node {
	return NewIdentifier(i.Ident).SetTree(i.tr).SetPos(i.Pos)
}

// AssignNode holds a list of variable names, possibly with chained field
// accesses. The dollar sign is part of the (first) name.
type VariableNode struct {
	NodeType
	Pos
	tr    *Tree
	Ident []string // Variable name and fields in lexical order.
}

func (t *Tree) newVariable(pos Pos, ident string) *VariableNode {
	return &VariableNode{tr: t, NodeType: NodeVariable, Pos: pos, Ident: strings.Split(ident, ".")}
}

func (v *VariableNode) String() string {
	var sb strings.Builder
	v.writeTo(&sb, &j2Context{})
	return sb.String()
}

func (v *VariableNode) writeTo(sb *strings.Builder, context *j2Context) {
	for i, id := range v.Ident {
		if i > 0 {
			sb.WriteByte('.')
		}
		sb.WriteString(id)
	}
}

func (v *VariableNode) tree() *Tree {
	return v.tr
}

func (v *VariableNode) Copy() Node {
	return &VariableNode{tr: v.tr, NodeType: NodeVariable, Pos: v.Pos, Ident: append([]string{}, v.Ident...)}
}

// DotNode holds the special identifier '.'.
type DotNode struct {
	NodeType
	Pos
	tr *Tree
}

func (t *Tree) newDot(pos Pos) *DotNode {
	return &DotNode{tr: t, NodeType: NodeDot, Pos: pos}
}

func (d *DotNode) Type() NodeType {
	// Override method on embedded NodeType for API compatibility.
	// TODO: Not really a problem; could change API without effect but
	// api tool complains.
	return NodeDot
}

func (d *DotNode) String() string {
	return "."
}

func (d *DotNode) writeTo(sb *strings.Builder, context *j2Context) {
	sb.WriteString(d.String())
}

func (d *DotNode) tree() *Tree {
	return d.tr
}

func (d *DotNode) Copy() Node {
	return d.tr.newDot(d.Pos)
}

// NilNode holds the special identifier 'nil' representing an untyped nil constant.
type NilNode struct {
	NodeType
	Pos
	tr *Tree
}

func (t *Tree) newNil(pos Pos) *NilNode {
	return &NilNode{tr: t, NodeType: NodeNil, Pos: pos}
}

func (n *NilNode) Type() NodeType {
	// Override method on embedded NodeType for API compatibility.
	// TODO: Not really a problem; could change API without effect but
	// api tool complains.
	return NodeNil
}

func (n *NilNode) String() string {
	return "nil"
}

func (n *NilNode) writeTo(sb *strings.Builder, context *j2Context) {
	sb.WriteString(n.String())
}

func (n *NilNode) tree() *Tree {
	return n.tr
}

func (n *NilNode) Copy() Node {
	return n.tr.newNil(n.Pos)
}

// FieldNode holds a field (identifier starting with '.').
// The names may be chained ('.x.y').
// The period is dropped from each ident.
type FieldNode struct {
	NodeType
	Pos
	tr    *Tree
	Ident []string // The identifiers in lexical order.
}

func (t *Tree) newField(pos Pos, ident string) *FieldNode {
	return &FieldNode{tr: t, NodeType: NodeField, Pos: pos, Ident: strings.Split(ident[1:], ".")} // [1:] to drop leading period
}

func (f *FieldNode) String() string {
	var sb strings.Builder
	f.writeTo(&sb, &j2Context{})
	return sb.String()
}

func (f *FieldNode) writeTo(sb *strings.Builder, context *j2Context) {
	for _, id := range f.Ident {
		sb.WriteByte('.')
		sb.WriteString(id)
	}
}

func (f *FieldNode) tree() *Tree {
	return f.tr
}

func (f *FieldNode) Copy() Node {
	return &FieldNode{tr: f.tr, NodeType: NodeField, Pos: f.Pos, Ident: append([]string{}, f.Ident...)}
}

// ChainNode holds a term followed by a chain of field accesses (identifier starting with '.').
// The names may be chained ('.x.y').
// The periods are dropped from each ident.
type ChainNode struct {
	NodeType
	Pos
	tr    *Tree
	Node  Node
	Field []string // The identifiers in lexical order.
}

func (t *Tree) newChain(pos Pos, node Node) *ChainNode {
	return &ChainNode{tr: t, NodeType: NodeChain, Pos: pos, Node: node}
}

// Add adds the named field (which should start with a period) to the end of the chain.
func (c *ChainNode) Add(field string) {
	if len(field) == 0 || field[0] != '.' {
		panic("no dot in field")
	}
	field = field[1:] // Remove leading dot.
	if field == "" {
		panic("empty field")
	}
	c.Field = append(c.Field, field)
}

func (c *ChainNode) String() string {
	var sb strings.Builder
	c.writeTo(&sb, &j2Context{})
	return sb.String()
}

func (c *ChainNode) writeTo(sb *strings.Builder, context *j2Context) {
	if _, ok := c.Node.(*PipeNode); ok {
		sb.WriteByte('(')
		c.Node.writeTo(sb, context)
		sb.WriteByte(')')
	} else {
		c.Node.writeTo(sb, context)
	}
	for _, field := range c.Field {
		sb.WriteByte('.')
		sb.WriteString(field)
	}
}

func (c *ChainNode) tree() *Tree {
	return c.tr
}

func (c *ChainNode) Copy() Node {
	return &ChainNode{tr: c.tr, NodeType: NodeChain, Pos: c.Pos, Node: c.Node, Field: append([]string{}, c.Field...)}
}

// BoolNode holds a boolean constant.
type BoolNode struct {
	NodeType
	Pos
	tr   *Tree
	True bool // The value of the boolean constant.
}

func (t *Tree) newBool(pos Pos, true bool) *BoolNode {
	return &BoolNode{tr: t, NodeType: NodeBool, Pos: pos, True: true}
}

func (b *BoolNode) String() string {
	if b.True {
		return "true"
	}
	return "false"
}

func (b *BoolNode) writeTo(sb *strings.Builder, context *j2Context) {
	sb.WriteString(b.String())
}

func (b *BoolNode) tree() *Tree {
	return b.tr
}

func (b *BoolNode) Copy() Node {
	return b.tr.newBool(b.Pos, b.True)
}

// NumberNode holds a number: signed or unsigned integer, float, or complex.
// The value is parsed and stored under all the types that can represent the value.
// This simulates in a small amount of code the behavior of Go's ideal constants.
type NumberNode struct {
	NodeType
	Pos
	tr         *Tree
	IsInt      bool       // Number has an integral value.
	IsUint     bool       // Number has an unsigned integral value.
	IsFloat    bool       // Number has a floating-point value.
	IsComplex  bool       // Number is complex.
	Int64      int64      // The signed integer value.
	Uint64     uint64     // The unsigned integer value.
	Float64    float64    // The floating-point value.
	Complex128 complex128 // The complex value.
	Text       string     // The original textual representation from the input.
}

func (t *Tree) newNumber(pos Pos, text string, typ itemType) (*NumberNode, error) {
	n := &NumberNode{tr: t, NodeType: NodeNumber, Pos: pos, Text: text}
	switch typ {
	case itemCharConstant:
		rune, _, tail, err := strconv.UnquoteChar(text[1:], text[0])
		if err != nil {
			return nil, err
		}
		if tail != "'" {
			return nil, fmt.Errorf("malformed character constant: %s", text)
		}
		n.Int64 = int64(rune)
		n.IsInt = true
		n.Uint64 = uint64(rune)
		n.IsUint = true
		n.Float64 = float64(rune) // odd but those are the rules.
		n.IsFloat = true
		return n, nil
	case itemComplex:
		// fmt.Sscan can parse the pair, so let it do the work.
		if _, err := fmt.Sscan(text, &n.Complex128); err != nil {
			return nil, err
		}
		n.IsComplex = true
		n.simplifyComplex()
		return n, nil
	}
	// Imaginary constants can only be complex unless they are zero.
	if len(text) > 0 && text[len(text)-1] == 'i' {
		f, err := strconv.ParseFloat(text[:len(text)-1], 64)
		if err == nil {
			n.IsComplex = true
			n.Complex128 = complex(0, f)
			n.simplifyComplex()
			return n, nil
		}
	}
	// Do integer test first so we get 0x123 etc.
	u, err := strconv.ParseUint(text, 0, 64) // will fail for -0; fixed below.
	if err == nil {
		n.IsUint = true
		n.Uint64 = u
	}
	i, err := strconv.ParseInt(text, 0, 64)
	if err == nil {
		n.IsInt = true
		n.Int64 = i
		if i == 0 {
			n.IsUint = true // in case of -0.
			n.Uint64 = u
		}
	}
	// If an integer extraction succeeded, promote the float.
	if n.IsInt {
		n.IsFloat = true
		n.Float64 = float64(n.Int64)
	} else if n.IsUint {
		n.IsFloat = true
		n.Float64 = float64(n.Uint64)
	} else {
		f, err := strconv.ParseFloat(text, 64)
		if err == nil {
			// If we parsed it as a float but it looks like an integer,
			// it's a huge number too large to fit in an int. Reject it.
			if !strings.ContainsAny(text, ".eEpP") {
				return nil, fmt.Errorf("integer overflow: %q", text)
			}
			n.IsFloat = true
			n.Float64 = f
			// If a floating-point extraction succeeded, extract the int if needed.
			if !n.IsInt && float64(int64(f)) == f {
				n.IsInt = true
				n.Int64 = int64(f)
			}
			if !n.IsUint && float64(uint64(f)) == f {
				n.IsUint = true
				n.Uint64 = uint64(f)
			}
		}
	}
	if !n.IsInt && !n.IsUint && !n.IsFloat {
		return nil, fmt.Errorf("illegal number syntax: %q", text)
	}
	return n, nil
}

// simplifyComplex pulls out any other types that are represented by the complex number.
// These all require that the imaginary part be zero.
func (n *NumberNode) simplifyComplex() {
	n.IsFloat = imag(n.Complex128) == 0
	if n.IsFloat {
		n.Float64 = real(n.Complex128)
		n.IsInt = float64(int64(n.Float64)) == n.Float64
		if n.IsInt {
			n.Int64 = int64(n.Float64)
		}
		n.IsUint = float64(uint64(n.Float64)) == n.Float64
		if n.IsUint {
			n.Uint64 = uint64(n.Float64)
		}
	}
}

func (n *NumberNode) String() string {
	return n.Text
}

func (n *NumberNode) writeTo(sb *strings.Builder, context *j2Context) {
	sb.WriteString(n.String())
}

func (n *NumberNode) tree() *Tree {
	return n.tr
}

func (n *NumberNode) Copy() Node {
	nn := new(NumberNode)
	*nn = *n // Easy, fast, correct.
	return nn
}

// StringNode holds a string constant. The value has been "unquoted".
type StringNode struct {
	NodeType
	Pos
	tr     *Tree
	Quoted string // The original text of the string, with quotes.
	Text   string // The string, after quote processing.
}

func (t *Tree) newString(pos Pos, orig, text string) *StringNode {
	return &StringNode{tr: t, NodeType: NodeString, Pos: pos, Quoted: orig, Text: text}
}

func (s *StringNode) String() string {
	return s.Quoted
}

func (s *StringNode) writeTo(sb *strings.Builder, context *j2Context) {
	sb.WriteString(s.String())
}

func (s *StringNode) tree() *Tree {
	return s.tr
}

func (s *StringNode) Copy() Node {
	return s.tr.newString(s.Pos, s.Quoted, s.Text)
}

// endNode represents an {{end}} action.
// It does not appear in the final parse tree.
type endNode struct {
	NodeType
	Pos
	tr *Tree
}

func (t *Tree) newEnd(pos Pos) *endNode {
	return &endNode{tr: t, NodeType: nodeEnd, Pos: pos}
}

func (e *endNode) String() string {
	return "{{ end }}"
}

func (e *endNode) writeTo(sb *strings.Builder, context *j2Context) {
	sb.WriteString(e.String())
}

func (e *endNode) tree() *Tree {
	return e.tr
}

func (e *endNode) Copy() Node {
	return e.tr.newEnd(e.Pos)
}

// elseNode represents an {{else}} action. Does not appear in the final tree.
type elseNode struct {
	NodeType
	Pos
	tr   *Tree
	Line int // The line number in the input. Deprecated: Kept for compatibility.
}

func (t *Tree) newElse(pos Pos, line int) *elseNode {
	return &elseNode{tr: t, NodeType: nodeElse, Pos: pos, Line: line}
}

func (e *elseNode) Type() NodeType {
	return nodeElse
}

func (e *elseNode) String() string {
	return "{% else %}"
}

func (e *elseNode) writeTo(sb *strings.Builder, context *j2Context) {
	sb.WriteString(e.String())
}

func (e *elseNode) tree() *Tree {
	return e.tr
}

func (e *elseNode) Copy() Node {
	return e.tr.newElse(e.Pos, e.Line)
}

// BranchNode is the common representation of if, range, and with.
type BranchNode struct {
	NodeType
	Pos
	tr       *Tree
	Line     int       // The line number in the input. Deprecated: Kept for compatibility.
	Pipe     *PipeNode // The pipeline to be evaluated.
	List     *ListNode // What to execute if the value is non-empty.
	ElseList *ListNode // What to execute if the value is empty (nil if absent).
}

func (b *BranchNode) String() string {
	var sb strings.Builder
	b.writeTo(&sb, &j2Context{})
	return sb.String()
}

func (b *BranchNode) writeTo(sb *strings.Builder, context *j2Context) {
	logrus.Info("---------------- Control Flow (if, range,for ,with) ------------------------")
	logrus.Infof("Reading control flow (if, range,for ,with). Template name %s", b.tr.Name)
	var rangeValues *map[string][]*helm.LogHelmReport
	var dotValue = make(map[string][]*helm.LogHelmReport)
	dotValue["."] = []*helm.LogHelmReport{}
	var itemField string
	removeBodyVariablePrefix := false
	name := ""
	switch b.NodeType {
	case NodeIf:
		name = "if"
	case NodeRange:
		name = "for"
	case NodeWith:
		name = "with"
	default:
		panic("unknown branch type")
	}

	if name == "for" {
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
		//b) if you have two variables then assume  you have key and value don't have to do anything
		//c) if you have  one variable then
		// 1. derive the item_name and prefix the variables under the cmds variable found  in values with $item_name
		switch b.GetRangeUseCaseType() {
		case UseCaseNoVariables:
			{ //{{- range .Values.ingress.secrets }}
				//extract item from single argument
				sb.WriteString("for")
				sb.WriteByte(' ')
				ss := strings.Split(b.Pipe.Cmds[0].Args[0].String(), ".")
				itemField = "item_" + ss[len(ss)-1]
				sb.WriteString(itemField)
				sb.WriteByte(' ')
				sb.WriteString("in")
				sb.WriteByte(' ')
				b.Pipe.writeForTo(sb)
				rangeValues, _ = helm.GetValues(b.Pipe.Cmds[0].Args[0].String())
				logrus.Info("Attempting to prefix variables with loop variable,example: `value` becomes `item.value` for template", b.tr.Name)
				//attempting for dot values
				if rangeValues == nil {
					logrus.Warnf("Failed to prefix variables with loop variable value for template %s", b.tr.Name)
					logrus.Warnf("%s at position %d was not found in Helm chart's values: invalid path", b.Pipe.Cmds[0].Args[0].String(), b.Pos)
					rangeValues = &dotValue
				} else {
					(*rangeValues)["."] = []*helm.LogHelmReport{}
				}

			}
		case UseCaseKeyValue, UseCaseSingleValue:
			{
				sb.WriteString("for")
				sb.WriteByte(' ')
				b.RemoveVarPrefix("$")
				b.Pipe.writeForTo(sb)
				removeBodyVariablePrefix = true
			}
		case UseCaseTuple:
			{
				sb.WriteString("for")
				sb.WriteByte(' ')
				b.Pipe.writeForTo(sb)
			}
		default:
			{
				sb.WriteString("for")
				sb.WriteString(name)
				sb.WriteByte(' ')
				b.Pipe.writeForTo(sb)
			}
		}
	} else {
		sb.WriteString("{% ")
		sb.WriteString(name)
		sb.WriteByte(' ')
	}
	if name == "if" {
		ctx := &j2Context{
			isConditional: true,
		}
		b.Pipe.writeTo(sb, ctx)
	} else if name != "for" {
		ctx := &j2Context{
			isConditional: false,
		}
		b.Pipe.writeTo(sb, ctx)
	}
	sb.WriteString(" %}")
	ctx := &j2Context{
		isConditional: false,
	}
	b.List.writeTo(sb, ctx)
	// all the things if the conditional is true
	//prefix  range variables with item
	if rangeValues != nil {
		helm.PreFixValuesWithItems(sb, itemField, rangeValues)
		logrus.Infof("**************************************************************")
		logrus.Info("       for loop variables replacement inside for loop : 	", b.tr.Name)
		logrus.Infof("**************************************************************")
		logrus.Infof("For loop: Following variables were prefixed with %s in template %s", itemField, b.tr.Name)
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

	if b.ElseList != nil {
		sb.WriteString("{% else %}")
		b.ElseList.writeTo(sb, ctx)
	}
	switch b.NodeType {
	case NodeIf:
		sb.WriteString("{% endif %}")
	case NodeRange:
		sb.WriteString("{% endfor %}")
	case NodeWith:
		sb.WriteString("{% with %}")
	default:
		panic("unknown branch type")
	}
}

func (b *BranchNode) tree() *Tree {
	return b.tr
}

func (b *BranchNode) Copy() Node {
	switch b.NodeType {
	case NodeIf:
		return b.tr.newIf(b.Pos, b.Line, b.Pipe, b.List, b.ElseList)
	case NodeRange:
		return b.tr.newRange(b.Pos, b.Line, b.Pipe, b.List, b.ElseList)
	case NodeWith:
		return b.tr.newWith(b.Pos, b.Line, b.Pipe, b.List, b.ElseList)
	default:
		panic("unknown branch type")
	}
}

// Remove $From $key $value
func (b *BranchNode) RemoveVarPrefix(prefix string) {
	for _, v := range b.Pipe.Decl {
		for i, _ := range v.Ident {
				v.Ident[i] = strings.TrimPrefix(v.Ident[i], prefix)
		}
	}
}


// GetRangeUseCaseType ... get difference cases for range flow
func (b *BranchNode) GetRangeUseCaseType() RangeUseCaseType {
	if len(b.Pipe.Decl) == 0 && len(b.Pipe.Cmds[0].Args) == 1 {
		return UseCaseNoVariables
	} else if len(b.Pipe.Decl) == 2 { //key value
		return UseCaseKeyValue
	} else if len(b.Pipe.Decl) == 1 && len(b.Pipe.Cmds[0].Args) == 1 { //$host and one value
		return UseCaseSingleValue
	} else if len(b.Pipe.Decl) == 0 && len(b.Pipe.Cmds[0].Args) == 1 && b.Pipe.Cmds[0].Args[0].String() == "tuple" {
		return UseCaseTuple
	}
	return UseCaseDefault
}

// IfNode represents an {{if}} action and its commands.
type IfNode struct {
	BranchNode
}

func (t *Tree) newIf(pos Pos, line int, pipe *PipeNode, list, elseList *ListNode) *IfNode {
	return &IfNode{BranchNode{tr: t, NodeType: NodeIf, Pos: pos, Line: line, Pipe: pipe, List: list, ElseList: elseList}}
}

func (i *IfNode) Copy() Node {
	return i.tr.newIf(i.Pos, i.Line, i.Pipe.CopyPipe(), i.List.CopyList(), i.ElseList.CopyList())
}

// RangeNode represents a {{range}} action and its commands.
type RangeNode struct {
	BranchNode
}

func (t *Tree) newRange(pos Pos, line int, pipe *PipeNode, list, elseList *ListNode) *RangeNode {
	return &RangeNode{BranchNode{tr: t, NodeType: NodeRange, Pos: pos, Line: line, Pipe: pipe, List: list, ElseList: elseList}}
}

func (r *RangeNode) Copy() Node {
	return r.tr.newRange(r.Pos, r.Line, r.Pipe.CopyPipe(), r.List.CopyList(), r.ElseList.CopyList())
}

// WithNode represents a {{with}} action and its commands.
type WithNode struct {
	BranchNode
}

func (t *Tree) newWith(pos Pos, line int, pipe *PipeNode, list, elseList *ListNode) *WithNode {
	return &WithNode{BranchNode{tr: t, NodeType: NodeWith, Pos: pos, Line: line, Pipe: pipe, List: list, ElseList: elseList}}
}

func (w *WithNode) Copy() Node {
	return w.tr.newWith(w.Pos, w.Line, w.Pipe.CopyPipe(), w.List.CopyList(), w.ElseList.CopyList())
}

// TemplateNode represents a {{template}} action.
type TemplateNode struct {
	NodeType
	Pos
	tr   *Tree
	Line int       // The line number in the input. Deprecated: Kept for compatibility.
	Name string    // The name of the template (unquoted).
	Pipe *PipeNode // The command to evaluate as dot for the template.
}

func (t *Tree) newTemplate(pos Pos, line int, name string, pipe *PipeNode) *TemplateNode {
	return &TemplateNode{tr: t, NodeType: NodeTemplate, Pos: pos, Line: line, Name: name, Pipe: pipe}
}

func (t *TemplateNode) String() string {
	var sb strings.Builder
	t.writeTo(&sb, &j2Context{})
	return sb.String()
}

func (t *TemplateNode) writeTo(sb *strings.Builder, context *j2Context) {
	sb.WriteString("{{ template ")
	sb.WriteString(strconv.Quote(t.Name))
	if t.Pipe != nil {
		sb.WriteByte(' ')
		t.Pipe.writeTo(sb, context)
	}
	sb.WriteString(" }}")
}

func (t *TemplateNode) tree() *Tree {
	return t.tr
}

func (t *TemplateNode) Copy() Node {
	return t.tr.newTemplate(t.Pos, t.Line, t.Name, t.Pipe.CopyPipe())
}
