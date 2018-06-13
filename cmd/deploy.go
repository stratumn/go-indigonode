// Copyright © 2018 Stratumn SAS
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

const (
	// InventoryFlagName is the name of the flag to pass the ansible inventory file.
	InventoryFlagName = "inventory"

	// DefaultInventory is the default name for the alice test network inventory file.
	DefaultInventory = "network.ini"

	// PlaybookFlagName is the name of the flag for the ansible playbook.
	PlaybookFlagName = "playbook"

	// DefaultPlaybook is the path to the ansible deployment playbook.
	DefaultPlaybook = "main.yml"

	// RemoteUserFlagName is the name of the flag for setting the remote user's name.
	RemoteUserFlagName = "user"

	// DefaultRemoteUser is the name of the user on the remote hosts.
	DefaultRemoteUser = "ubuntu"

	// DeploymentKeyFlagName is the name of the flag for the deployment key.
	DeploymentKeyFlagName = "key"

	// DefaultDeploymentKey is the path to the private key setup on the cloud platform.
	DefaultDeploymentKey = "/keybase/team/stratumn_eng/alice_test_network/alice-test-key.pem"

	// AnsibleConfigFlagName is the name of the flag ansible configuration file path.
	AnsibleConfigFlagName = "ansible-cfg"

	// DefaultAnsibleConfig is the path to the ansible configuration file.
	DefaultAnsibleConfig = "ansible.cfg"

	// DefaultPythonRequirements is the path to the python requirements file.
	DefaultPythonRequirements = "requirements.txt"

	// DefaultAnsibleRequirements is the path to the ansible requirements file.
	DefaultAnsibleRequirements = "requirements.yml"
)

var (
	inventoryFile string

	playbookFile string

	remoteUser string

	deploymentKeyPath string

	ansibleConfig string

	pythonDeps string

	ansibleDeps string
)

func setupEnv(cmd *exec.Cmd) {
	_, err := os.Stat(deploymentKeyPath)
	if err != nil {
		osExit(1, "a private key is needed to authenticate on the cloud platform")
	}
	cmd.Env = append(os.Environ(), fmt.Sprintf("ANSIBLE_CONFIG=%s", ansibleConfig))
}

func getDependencies() {
	_, err := exec.LookPath("python")
	if err != nil {
		osExit(1, "python must be installed to run alice deploy")
	}

	// Install python dependencies.
	fmt.Println("Updating dependencies...")
	err = exec.Command("python", "-m", "pip", "install", "-r", DefaultPythonRequirements).Run()
	if err != nil {
		osExit(1, fmt.Sprintf("could not install python dependencies: %s", err.Error()))
	}

	// Install ansible dependencies.
	err = exec.Command("ansible-galaxy", "install", "-r", DefaultAnsibleRequirements, "-p", "roles").Run()
	if err != nil {
		osExit(1, fmt.Sprintf("could not install ansible roles: %s", err.Error()))
	}
	fmt.Println("Done!")
}

func runDeploy(cmd *exec.Cmd) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		osExit(1, err.Error())
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		osExit(1, err.Error())
	}
	err = cmd.Start()
	if err != nil {
		osExit(1, fmt.Sprintf("error while running ansible: %s", err))
	}

	reader := func(stream io.ReadCloser) {
		streamScanner := bufio.NewScanner(stream)
		for streamScanner.Scan() {
			fmt.Println(streamScanner.Text())
		}
	}
	go reader(stdout)
	go reader(stderr)
	if err := cmd.Wait(); err != nil {
		osExit(1, err.Error())
	}
}

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploys an alice network on the cloud",
	Run: func(cmd *cobra.Command, args []string) {
		getDependencies()
		ansibleCmd := exec.Command(
			"ansible-playbook",
			"--inventory", inventoryFile,
			"--user", remoteUser,
			"--private-key", deploymentKeyPath,
			playbookFile)
		setupEnv(ansibleCmd)
		runDeploy(ansibleCmd)
	},
}

func init() {
	RootCmd.AddCommand(deployCmd)

	deployCmd.Flags().StringVar(
		&inventoryFile,
		InventoryFlagName,
		DefaultInventory,
		"ansible inventory file",
	)

	deployCmd.Flags().StringVar(
		&playbookFile,
		PlaybookFlagName,
		DefaultPlaybook,
		"ansible playbook",
	)

	deployCmd.Flags().StringVar(
		&deploymentKeyPath,
		DeploymentKeyFlagName,
		DefaultDeploymentKey,
		"private key used to authenticate to the cloud platform",
	)

	deployCmd.Flags().StringVar(
		&ansibleConfig,
		AnsibleConfigFlagName,
		DefaultAnsibleConfig,
		"ansible configuration file",
	)

	deployCmd.Flags().StringVar(
		&remoteUser,
		RemoteUserFlagName,
		DefaultRemoteUser,
		"user name on the remote hosts",
	)
}
