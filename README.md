# Helm Ansible Template Exporter (WIP)

This tool is aimed at automating the conversion from Helm to Ansible Playbook Role.  Based on the diverging operating
models, some aspects of Helm charts are difficult or impossible to convert directly to Ansible Playbook roles.  This
tool makes a best attempt to automatically convert a Helm chart into an Ansible Playbook Role.

Caveat Emptor:  This code is very much a work in progress.

## Current Capabilities
The current offering does the following:
1) Creates a role in the "./workspace" directory using ansible-galaxy.
2) Copies templates into the generated Ansible Playbook Role templates directory, renaming files with a ".j2" extension.
3) Merges values.yml (or values.yaml) into the generated Ansible Playbook Role defaults/main.yml file.
4) Searches the generated Ansible Playbook Role's defaults/main.yml file for self references (i.e., references to
.Values.) and comments them out.  Ansible Playbook is incapable of expressing self references in defaults/main.yml,
a clear technology gap between Ansible Playbook and Helm charts.  A "WARN" message is output indicating that a manual
change is required to defaults/main.yml on the appropriate lines.
5) Removes references to ".Values." in the generated Ansible Playbook's Roles' templates.  Ansible Playbook can directly
reference the values in defaults/main.yml, so this reference is not needed.

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
Additionally, ansible-galaxy must be included in the $PATH.

### Run Instructions

```shell script
make run
``` 
This will run the current implementation of code.
