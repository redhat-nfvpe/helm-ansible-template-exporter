/*
Package ansiblegalaxy provides a GO wrapper interface to the ansible-galaxy client library.
 */
package ansiblegalaxy

import (
	"github.com/sirupsen/logrus"
    "os/exec"
)

const ansibleGalaxyCommand = "ansible-galaxy"
const ansibleGalaxyRoleCommand = "role"
const ansibleGalaxyInitCommand = "init"
const ansibleGalaxyInitPathOption = "--init-path"
const keyValueSeparator = "="

// Given a directory to store Ansible roles, form the ansible-galaxy --init-path option appropriately.
func formAnsibleGalaxyInitPathOption(rolesDirectory string) string {
    return ansibleGalaxyInitPathOption + keyValueSeparator + rolesDirectory
}

// Ensures that ansible-galaxy is resolvable on the given $PATH.  If ansible-galaxy is not found, error information is
// logged followed by a fatal exit.
func ensureAnsibleGalaxyIsInstalled() {
    _, lookErr := exec.LookPath(ansibleGalaxyCommand)
    if lookErr != nil {
        logrus.Errorf("Cannot find %s;  is it installed?", ansibleGalaxyCommand)
        logrus.Fatalln(lookErr)
    }
}

// Generates an Ansible Playbook Role using ansible-galaxy in the rolesDirectory directory.
func InstallAnsibleRole(roleName string, rolesDirectory string) {
    ensureAnsibleGalaxyIsInstalled()

    output, execErr := exec.Command(ansibleGalaxyCommand, ansibleGalaxyRoleCommand, ansibleGalaxyInitCommand,
        formAnsibleGalaxyInitPathOption(rolesDirectory), roleName).CombinedOutput()
    if execErr != nil {
        logrus.Error(string(output))
        logrus.Fatalln(execErr)
    } else {
        logrus.Infof("Successfully initialized the Ansible Role \"%s\" in %s", roleName, rolesDirectory)
    }
}