###  Steps to build an operator from an existing helm charts

##### ./hack/build-operator.sh -h 
~~~
usage: build-operator.sh  -r -b -d -e -c  | -h
-e | --export : Export helmcharts and create Ansible operator
-b | --build  : Build Operator image and push it to quay.io
-d | --deploy:  Deploy this operator to existing cluster
-r | --run : Option to run the operator outside the cluster
-c | --delete : Clean the cluster by deleteing the operator
~~~
#####Setting up environment variables
 Step 1: update and source ./env.sh to set environment variables or 
 set following variables in command line.
 * Note: If you override any variables via CLI, first source env.sh, which will unset some env variables.
 Example: 
 ```
source env.sh
export role=foo
export workspace=./workspace
export helm_chart=path to your helmchart
#Change this to your namespace
export quay_namespace="YOUR_NAMESPACE"
export INSTALL_OPERATOR_SDK=1

```
  Note : To unset all exported variables run this command.
 ` unset $(awk -F'[ =]+' '/^export/{print $2}' ./env.sh)`
 
##### Change following variables.
* `role` : should be your helm chart name, which becomes Ansible role.
* `worskpace` : Location where the operator and exported templates will be created.
* `helm_chart`: Local path to the helm charts to export.
* INSTALL_OPERATOR_SDK=1 #if you want the script to install operator sdk.
 
 ###### Variables required by operator-sdk
*`quay_namespace`: Your quay namespace to build and push the operator image
* `kind` : Kind of the CR to be created
* `apiVersion` : Version of the CR to be created.
 
Example
```
apiVersion: "foo.example.com/v1alpha1"
kind: "Foo"
metadata:
  name: "{{ meta.name }}" 
```
Once you are satisfied with above variables, source the file(if you not overriding via cli)
`
source ./env.sh
`
#####Running build-operator script to create an operator
1. ``./hack/build-operator.sh --export``  
The `--export` option will export the existing helm charts into an operator, by creating a scaffold using operator sdk and 
parsing helm templates and converting them to Ansible templates.
The output will create two folder under the workspace folder as set via workspace env variable.
```
{workspace}-
      - {Role} # This is exported helm files generated but this tool for you reference. 
               Operator sdk will copy the contents  into {role}-operator folder and create an operator.
               You can delete this folder or keep it for your reference.
      - {Operator} #This is the final operator folder structure. 
               The Anisible file structure within the operator will be same as mentioned in main README file for the outputed exported ansible role.
 ```
                                  
The converted operator is not yet usable since exporting tool is unable to convert all the templates fields.
For more details , read main [README.md](../README.md) file.
### Testing  Exported K8s Ansible template
Now you have an operator created , Lets see how to test the exported Ansible templates 
and make changes to the templates and make it work.

 Step 1 : Have a Cluster up and running , example using minikube
`minikube start`
 
 Step 2 Create a playbook.yaml (Which you can delete after testing and fixing Ansible templates) and copy it under roles folder
 `Create an Ansible playbook playbook.yml in the top-level directory which includes role example Foo:` 
```
---
- hosts: localhost
  roles:
  - Foo
```
Getting started with the k8s Ansible modules

Since we are interested in using Ansible for the lifecycle management of our application on Kubernetes, 
it is beneficial for a developer to get a good grasp of the k8s Ansible module. 
This Ansible module allows a developer to either leverage their existing Kubernetes resource files (written in YaML) or express the lifecycle management in native Ansible.
One of the biggest benefits of using Ansible in conjunction with existing Kubernetes resource files is the ability to 
use Jinja templating so that you can customize deployments with the simplicity of a few variables in Ansible.

The easiest way to get started is to install the modules on your local machine and test them using a playbook.
Installing the k8s Ansible modules

To install the k8s Ansible modules, one must first install Ansible 2.9+. On Fedora/Centos:

`$ sudo dnf install ansible`

In addition to Ansible, a user must install the OpenShift Restclient Python package. This can be installed from pip:

`$ pip3 install openshift`

Finally, a user must install the Ansible Kubernetes collection from ansible-galaxy:

`$ ansible-galaxy collection install community.kubernetes`

Run `$ansible-playbook playbook.yaml -vvv` 

Debug the output and fix all the changes required to make this template working as mentioned in README.

2. Edit all the  templates and change the name of the resource to `name: {{ meta.name }}` and any namespace mentioned to 
`namespace: {{ meta.namespace }}`, This will give the k8s unique name based off its custom resource name and namespace.
   
3. ``./hack/build-operator.sh --build``
This  will build the operator image and deploy to a quay repository and update the operator.yaml with the uploaded 
image name.
Under the hood, the script will be running following commands
```
$ operator-sdk build quay.io/example/foo-operator:v0.0.1
$ docker push quay.io/example/foo-operator:v0.0.1
```

4. ``./hack/build-operator.sh --deploy`` 
This is will deploy the operator to running cluster
The above script will be running similar commands
```
sed -i 's|{{REPLACE_IMAGE}}|quay.io/example/foo-operator:v0.0.1|g' deploy/operator.yaml
$ kubectl create -f deploy/crds/foo.example.com_foos_crd.yaml # if CRD doesn't exist already
$ kubectl create -f deploy/service_account.yaml
$ kubectl create -f deploy/role.yaml
$ kubectl create -f deploy/role_binding.yaml
$ kubectl create -f deploy/operator.yaml
 ```
 
Now you can now override the Ansible variables ( roles/deafault/main.yaml) from your CR file which is found under deploy/crds/foo.example.com_foos_cr.yaml
All the variables you wish to override  should be under the spec: metadata of the this CR
Then apply the CR
` kubectl apply -f deploy/crds/foo.example.com_foos_cr.yaml `
Applying the CR will make the operator to deploy all kubernetes resources from the ansible templates.

For more information read this [operator sdk ansible developers guide](https://github.com/operator-framework/operator-sdk/blob/master/doc/ansible/dev/developer_guide.md)


5. Debugging Operator and checking if its running fine 
* checking if operator is running
    * kubectl get pods -n default #namespace can be default or the namespace that was deployed to
    * kubectl logs -f {the operator pod} -n default -c operator
    * kubectl logs -f {the operator pod} -n default -c ansible

 
6. ``./hack/build-operator.sh --delete ``
This will delete the deployed operator from the cluster.

         

 

