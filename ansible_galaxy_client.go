package main

import (
    "github.com/sirupsen/logrus"
    "os/exec"
)

const AnsibleGalaxyCommand = "ansible-galaxy"
const AnsibleGalaxyRoleCommand = "role"
const AnsibleGalaxyInitCommand = "init"
const AnsibleGalaxyInitPathOption = "--init-path"
const KeyValueSeparator = "="

// Given a directory to store Ansible roles, form the ansible-galaxy --init-path option appropriately.
func formAnsibleGalaxyInitPathOption(rolesDirectory string) string {
    return AnsibleGalaxyInitPathOption + KeyValueSeparator + rolesDirectory
}

// Generates an Ansible Playbook Role using ansible-galaxy in the rolesDirectory directory.
func InstallAnsibleRole(roleName string, rolesDirectory string) {
    _, lookErr := exec.LookPath(AnsibleGalaxyCommand)
    if lookErr != nil {
        logrus.Errorf("Cannot find %s;  is it installed?", AnsibleGalaxyCommand)
        logrus.Fatalln(lookErr)
    }

    output, execErr := exec.Command(AnsibleGalaxyCommand, AnsibleGalaxyRoleCommand, AnsibleGalaxyInitCommand,
        formAnsibleGalaxyInitPathOption(rolesDirectory), roleName).CombinedOutput()
    if execErr != nil {
        logrus.Error(string(output))
        logrus.Fatalln(execErr)
    } else {
        logrus.Infof("Successfully initialized the Ansible Role \"%s\" in %s", roleName, rolesDirectory)
    }
}