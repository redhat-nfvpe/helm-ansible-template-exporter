# Helm Ansible Template Exporter

Helm Ansible Template Exporter is a set of tools aimed at helping K8S application developers transition from Helm to
Ansible Role Based Operators.  This project is separated into two main tools:

1) Helm To Ansible Exporter
2) [Ansible Operator Builder](./hack/README.md)

The Ansible Operator Builder is targeted towards developer usage, and is located in the [hack directory](./hack).

## Helm To Ansible Exporter

This tool automates conversion from [Helm Chart](https://helm.sh/) to
[Ansible Playbook Role](https://docs.ansible.com/ansible/latest/user_guide/playbooks_reuse_roles.html).  Helm Charts are
defined utilizing [Go Template Language](https://golang.org/pkg/text/template/), and Ansible Playbook Roles are
defined using Yaml and [Jinja2 Template Language](https://jinja.palletsprojects.com).  Due to the fact that Go Template
Language and Jinja2 Template Language are completely separate languages, some facets of Helm charts are difficult or
impossible to convert automatically.  This tool performs a best attempt conversion of a Helm chart into an Ansible
Playbook Role, but some aspects of the conversion process must be performed manually by hand after utilizing this tool.
Additionally, when ambiguities in conversion arise, this tool makes a best attempt to emit warnings where a manual
conversion or inspection is needed after running the tool.

The output of this tool is an Ansible Role which can be installed using
[Ansible Playbook](https://docs.ansible.com/ansible/latest/cli/ansible-playbook.html), or built into an operator using
the Ansible Operator Builder.

### Helm To Ansible Exporter Current Capabilities
The current offering does the following:
1)  Creates a role in the workspace directory using ansible-galaxy.
2)  Raw copies templates into the generated Ansible Playbook Role templates directory, renaming each template with a
    ".j2" extension.
3)  Merges values.yml (or values.yaml) into the generated Ansible Playbook Role defaults/main.yml file.
4)  Searches the generated Ansible Playbook Role's defaults/main.yml file for self references (i.e., references to
    .Values.) and comments them out.  Ansible Playbook is incapable of expressing self references in defaults/main.yml,
    a clear technology gap between Ansible Playbook and Helm charts.  A "WARN" message is output indicating that a
    manual change is required to defaults/main.yml on the appropriate lines.
5)  Convert Branch syntax for `if`, `range <list>` and `range <map>` in each template to utilize proper Jinja2 syntax.
    This includes a heuristic which attempts to determine if conditionals are checking for definition v.s. boolean
    evaluation.
6)  Convert boolean composition ordering.  Go Templating utilizes "and <condition1> <condition2>" format.  On the other
    hand, Jinja2 utilizes "<condition1> and <condition2>" formatting.  helmExport handles this conversion automatically.
7)  Template functions invocations are converted to Jinja2 Ansible Filter invocations.  This requires converting direct
    function invocations to piped invocations, as well as inserting the appropriate parentheses and commas for argument
    lists.
8)  Removes references to ".Values." in the generated Ansible Playbook's Roles' templates.  Ansible Playbook can directly
    reference the values in defaults/main.yml, so this reference isn't needed.
9)  If the "generateFilters" flag is set to true, some stub Ansible Filter implementations are installed for use in
    converting Go template functions to Ansible Playbook Filters.
10) If the "emitKeysSnakeCase" flag is set to true (default), all Ansible keys are converted to snake_case.  This option
    is especially useful if you plan to generate a K8S operator, since the operator-sdk converts Custom Resource
    variables to snake_case.  This tool directly invokes the ToSnake provided by operator-sdk in an attempt to exactly
    match the conversion functionality.
   
### Helm To Ansible Exporter Known Limitations

#### Go/Sprig Template Function Limitations

[Ansible Playbook Filter(s)](https://docs.ansible.com/ansible/latest/user_guide/playbooks_filters.html) are one
candidate to replace Go/Sprig Template Functions.  However, they are not a direct 1:1 mapping, so a few extra steps must
be taken after conversion.

#### Implement or Replace Template Function Invocations with Ansible Filters
There are a number of
[Ansible Playbook Filters](https://docs.ansible.com/ansible/latest/user_guide/playbooks_filters.html) built directly
into Ansible.  For instances where there does not seem to be a replacement, you will need to
[create your own replacement](https://www.dasblinkenlichten.com/creating-ansible-filter-plugins/).  If you utilize the
"generateFilters" command line argument, some example filters will be installed into your generated Ansible Playbook
Role.

#### "template" and "include" are not supported

Go Templates provide "template" and "include" in order to support dynamic template creation.  Ansible has no direct
replacement, and although there are some similar constructs, helmExport currently doesn't support any conversion.
Instead, consider using Ansible defaults as a replacement.

### Building The Exporter

To build the code, use the following command:

```shell script
make
```

To clean the code base, use the following command:
```shell script
make clean
```

### Running the Exporter

#### Runtime Dependencies

Helm Ansible Template Exporter requires [ansible-galaxy](https://galaxy.ansible.com/) to initialize exported roles.
Additionally, ansible-galaxy must be included in the $PATH.  If you utilize the "generateFilters" flag, then a Go
compiler must be installed.

#### Run Instructions

```shell script
make run
``` 
This will run the current implementation of code.

Alternatively:

```shell script
./helmExport export nginx --helm-chart=./example --workspace=./workspace --generateFilters=true --emitKeysSnakeCase=true
```

### Testing the Ansible Playbook Role

Ansible Operators are deployed using the
[`k8s` Ansible Module](https://docs.ansible.com/ansible/latest/modules/k8s_module.html).  The `k8s` Ansible Module
allows developers to leverage existing Kubernetes resource YAML files, and express lifecycle management in native
Ansible.  One benefit of using Ansible in conjunction with existing Kubernetes resource files is the ability to use
Jinja2 templating to customize deployments.

#### Installing the k8s Ansible Module

The k8s Ansible Module requires Ansible.  To install on Fedora/CentOS:

```shell script
sudo dnf install ansible
```

The OpenShift Restclient Python package is also required.  Install using pip:

```shell script
pip3 install openshift
```

Lastly, the Ansible Kubernetes collection is required. Install using ansible-galaxy:

```shell script
ansible-galaxy collection install community.kubernetes
```

#### Testing the Ansible Playbook Role Using the k8s Ansible Module

1. Start a K8S cluster.  For example, using [minikube](https://minikube.sigs.k8s.io/docs/):

```shell script
minikube start
```
 
2. Create an Ansible Playbook called `workspace/role/playbook.yaml`.

```yaml
---
- hosts: localhost
  roles:
  - Foo
```

3. Debug any issues not covered by the Helm Ansible Template Exporter by running this diagnostic command:

```shell script
ansible-playbook playbook.yaml -vvv
```

## Known Test Environments

This tool has been tested using:
* Fedora 31
* macOS Catalina Version 10.15.4
