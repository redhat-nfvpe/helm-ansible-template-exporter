# Custom Go text/template implementation (WIP)

## Introduction

The Golang text/template parser is particularly useful for creating templated files.  Helm utilizes the Golang
text/template functionality to output Helm chart YAML.

Ansible Playbook uses YAML roles, and it is capable of utilizing Jinja2 templates.

Jinja2 is not the same as Golang templating, and the languages actually differ quite a bit.  There are numerous
syntactic and semantic differences between the languages, making a 1:1 translation from Golang template to Jinja2
template extremely difficult.  Thus, in order to fully understand the Abstract Syntax Tree, an extended Golang parser is
needed.  This is a fork of the standard Golang text/template module, and it is used to aid in conversion from Golang
templating into Jinja2 templating.

### Status Quo

This work is ongoing.  Currently, the template parser is modified to handle conversion of branch syntax from Golang to
Jinja2 templating.  Since the Abstract Syntax Tree (AST) BranchNode functions are directly modified, this solution
supports nested translation.

#### If-Conditional
```gotemplate
{{ if condition }}
...
{{ end }}
```
Becomes:
```yaml
{% if condition %}
...
{% endif %}
```

#### Range
```gotemplate
{{ range .Values.ingress.secrets }}
...
{{ end }}
```
Becomes:
```yaml
{% range .Values.ingress.secrets %}
...
{% endfor %}
```
Note:  As of now, there is no way to convert range into for-each syntax.  This is future work.


## Changes

The normal changes in an internal fork have been performed, including changing import references etc.  Additionally,
since the upstream text/template occasionally utilizes "internal" imports, some of that functionality needed to be
copied into a place that is accessible from a non-Golang package.  Specifically, the SortedMap implementation and some
related functionality was copied from "internal/fmtsrt" to exec.go.  Since exec.go was the only client, it seemed like
an appropriate place to port the code.

### Changes Related to Conversion

The biggest change thus far has been the treatment of BranchNode output functionality in text/template/parse/node.go.

```go
func (b *BranchNode) writeTo(sb *strings.Builder) {
	name := ""
	switch b.NodeType {
	case NodeIf:
		name = "if"
	case NodeRange:
		name = "range"
	case NodeWith:
		name = "with"
	default:
		panic("unknown branch type")
	}
	sb.WriteString("{% ")
	sb.WriteString(name)
	sb.WriteByte(' ')
	b.Pipe.writeTo(sb)
	sb.WriteString(" %}")
	b.List.writeTo(sb)
	if b.ElseList != nil {
		sb.WriteString("{% else %}")
		b.ElseList.writeTo(sb)
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
```

Essentially, the writeTo method was rewritten in order to output Jinja2 branching syntax.

## Future Work

One big question involves whether we really need a fork for this package.  We may be able to utilize reflection in order
to extend the template engine, which would be more ideal since this approach is fraught with maintainability issues
should the upstream version of text/template ever change significantly.  For now, it is the path of least resistance.

Future work also includes utilizing this forked parser to handle other parts of conversion.  Some big ticket items are:
1) Translate "range" to for-each.
2) Handle translation of Go Template Functions into Ansible Playbook Filter(s).
3) Eliminate "define" function
4) Eliminate "include" function
