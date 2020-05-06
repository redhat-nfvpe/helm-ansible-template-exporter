## Ansible Operator Builder

Transitioning a Helm Chart to an Ansible Playbook Role is useful since many users prefer Ansible to Go Templating.
However, a transition to Ansible Playbook Role does not address Day 2 facets of
[Operational Lifecycle Management (OLM)](https://github.com/operator-framework/operator-lifecycle-manager).
The operator-framework [operator-sdk](https://github.com/operator-framework/operator-sdk) includes functionality to
develop Lifecycle Management using an Ansible Playbook Role.

[`build-operator.sh`](./hack/build-operator.sh) is a tool that takes an exported Ansible Role and builds an Ansible
Operator.

### Build An Operator From An Existing Helm Chart

build-operator comes with a number of command line options described below:

```
usage: build-operator.sh  -r -b -d -e -c  | -h
-e | --export : Export helmcharts and create Ansible operator
-b | --build  : Build Operator image and push it to quay.io
-d | --deploy:  Deploy this operator to existing cluster
-r | --run : Option to run the operator outside the cluster
-c | --delete : Clean the cluster by deleteing the operator
```

#### Configuring Environment Variables
build-operator.sh is configured through setting a number of environment variables in [`env.sh`](env.sh).

1. Change following variables:
   * `role`:  The Helm Chart name, which becomes Ansible Role Name.
   * `worskpace`:  An arbitrary directory location to export the target Operator and Ansible Playbook.
   * `helm_chart`:  File Path location to the original helm charts.
   * `quay_namespace`:  Your quay.io namespace.
   * `kind`:  The Kind of the Custom Resource(CR) to create.
   * `apiVersion`:  Version of the CR to create.
   * `INSTALL_OPERATOR_SDK=1`:  If you want the script to install operator-sdk.

2. Source ./env.sh to set environment variables or set following variables in command line.

```shell script
source env.sh
```

*Note*: To unset all exported variables, run this command:

```shell script
unset $(awk -F'[ =]+' '/^export/{print $2}' ./env.sh)
```

#### Running build-operator script to create an operator

Export the Helm Chart as an Ansible Playbook

```shell script
./hack/build-operator.sh --export
```  

The `--export` option will generate an Ansible Playbook Role and corresponding Operator implementation in the
`workspace` directory.  The `workspace` directory will contain two folders:
* `role`:          The Ansible Playbook Role, which is exported using Helm Template Ansible Exporter.  This directory
                   can optionally be deleted, since its contents are copied into `workspace/{role}-operator`. 
* `role-operator`: The generated operator, which can be deployed to a K8S cluster.
                                  
The converted operator in `workspace/role-operator` may not yet be usable since Helm Template Ansible Exporter cannot
automate 100% of the conversion process.  See the "Helm To Ansible Exporter Known Limitations" for more information.

### Building the Ansible Operator

Invoke `build-operator.sh` using the build argument:

```shell script
./hack/build-operator.sh --build
```

The `--build` argument builds the Docker container image, uploads the built image to the quay repository,  and update
`operator.yaml` with the uploaded image name.

### Deploying the Ansible Operator

Invoke `build-operator.sh` uisng the deploy argument:
 
```shell script
./hack/build-operator.sh --deploy
```

The `--deploy` argument deploys the operator to the K8S cluster.
 
You can now override the Ansible defaults variables `roles/defaults/main.yaml` from your CR file
`deploy/crds/foo.example.com_foos_cr.yaml`.  The variables that can be overridden are in the `spec` metadata of the CR.
After making appropriate edits to the defaults, you can apply the CR using:

```shell script
kubectl apply -f deploy/crds/foo.example.com_foos_cr.yaml
```

Applying the CR will cause the operator to deploy all K8S resources from the Ansible templates.  For more information
consult the [operator-sdk Ansible Developer Guide](https://github.com/operator-framework/operator-sdk/blob/master/doc/ansible/dev/developer_guide.md).

### Debugging the Running Ansible Operator

Use commands similar to the following in order to debug the running Ansible Operator: 

```shell script
kubectl get pods -n default
kubectl logs -f <operator pod> -n default -c operator
kubectl logs -f <operator pod> -n default -c ansible
```

### Delete the Ansible Operator from the K8S Cluster

To delete the Ansible Operator from the K8S cluster, run the following:
 
```shell script
./hack/build-operator.sh --delete
```

This will delete the deployed operator from the cluster.
