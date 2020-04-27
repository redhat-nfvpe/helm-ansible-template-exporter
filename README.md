# Helm Ansible Template Exporter

## Description

This tool automates conversion from Helm Chart to Ansible Playbook Role.  Helm Charts are defined utilizing Go
templates, while Ansible Playbook Roles are defined using Yaml and Jinja2 templates.  Due to the fact that Go and Jinja2
are completely separate languages, some facets of Helm charts are difficult or impossible to convert directly to
Ansible Playbook roles.  This tool attempts to automatically convert a Helm chart into an Ansible Playbook Role, but
some aspects of the conversion process must be performed manually by hand after utilizing this tool.

## Current Capabilities
The current offering does the following:
1) Creates a role in the workspace directory using ansible-galaxy.
2) Raw copies templates into the generated Ansible Playbook Role templates directory, renaming each template with a
   ".j2" extension.
3) Merges values.yml (or values.yaml) into the generated Ansible Playbook Role defaults/main.yml file.
4) Searches the generated Ansible Playbook Role's defaults/main.yml file for self references (i.e., references to
   .Values.) and comments them out.  Ansible Playbook is incapable of expressing self references in defaults/main.yml,
   a clear technology gap between Ansible Playbook and Helm charts.  A "WARN" message is output indicating that a manual
   change is required to defaults/main.yml on the appropriate lines.
5) Convert Branch syntax for "if", "for" and "for key/value" in each template to utilize proper Jinja2 syntax.  This
   includes a heuristic which attempts to determine if conditionals are checking for definition v.s. boolean evaluation.
6) Convert boolean composition ordering.  Go Templating utilizes "and <condition1> <condition2>" format.  On the other
   hand, Jinja2 utilizes "<condition1> and <condition2>" formatting.  helmExport handles this conversion automatically.
7) Removes references to ".Values." in the generated Ansible Playbook's Roles' templates.  Ansible Playbook can directly
   reference the values in defaults/main.yml, so this reference isn't needed.
8) If the "generateFilters" flag is set to true, some stub Ansible Filter implementations are installed for use in
   converting Go template functions to Ansible Playbook Filters.

## Known Limitations

### Go/Sprig Template Function Limitations

[Ansible Playbook Filter(s)](https://docs.ansible.com/ansible/latest/user_guide/playbooks_filters.html) are one
candidate to replace Go/Sprig Template Functions.  However, they are not a direct 1:1 mapping, so a few extra steps must
be taken after conversion.

#### Convert Syntax To Piped Ansible Filter Invocation

Ansible Playbook Filters utilize piped syntax input.  For example:

```yaml
{{ some_variable | to_json }}
{{ some_variable | to_yaml }}
```

In other words, one cannot invoke an Ansible Filter in the following ways, as input bust be piped:

```yaml
{{ to_json some_variable }}
{{ to_yaml(some_variable) }}
```

On the other hand, Go Template Functions allow inlining function calls.  An example from the
[nginx Helm Chart](https://github.com/bitnami/charts/blob/master/bitnami/nginx/templates/deployment.yaml#L114) is:
```gotemplate
resources: {{- toYaml .Values.metrics.resources | nindent 12 }}
```

Thus, following conversion, you would need to by-hand convert to something similar to:
```yaml
resources: "{{- metrics.resources | toYaml | nindent(12) }}"
```

If no first argument exists, use an empty string:
```yaml
resources: "{{ '' | generateResources }}"
```

#### Convert Syntax to Using Parentheses for Arguments

Unlike Go/Sprig Template functions, Jinja2 requires parentheses for arguments.  Thus expressions such as:
```gotemplate
resource: {{- metrics.resource | nindent 12 }}
```
have to be converted to utilize parentheses after conversion.  The above expression would become:
```yaml
resource: {{- metrics.resource | nindent(12) }}
```

#### Implement or Replace Template Function Invocations with Ansible Filters
There are a number of
[Ansible Playbook Filters](https://docs.ansible.com/ansible/latest/user_guide/playbooks_filters.html) built directly
into Ansible.  For instances where there does not seem to be a replacement, you will need to
[create your own replacement](https://www.dasblinkenlichten.com/creating-ansible-filter-plugins/).  If you utilize the
"generateFilters" command line argument, some example filters will be installed into your generated Ansible Playbook
Role.

### "template" and "include" are not supported

Go Templates provide "template" and "include" in order to support dynamic template creation.  Ansible has no direct
replacement, and although there are some similar constructs, helmExport currently doesn't support any conversion.
Instead, consider using Ansible defaults as a replacement.

## Building

To build the code, use the following command:

```shell script
make
```

To clean the code base, use the following command:
```shell script
make clean
```

## Running the Code

### Runtime Dependencies

Helm Ansible Template Exporter requires [ansible-galaxy](https://galaxy.ansible.com/) to initialize exported roles.
Additionally, ansible-galaxy must be included in the $PATH.  Additionally, if you utilize the "generateFilters" flag,
then a Go compiler must be installed.

### Run Instructions

```shell script
make run
``` 
This will run the current implementation of code.

Alternatively:

```shell script
./helmExport export nginx --helm-chart=./example --workspace=./workspace --generateFilters=true
```
