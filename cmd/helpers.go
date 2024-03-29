package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/oberd/ecsy/ecs"
	"github.com/spf13/cobra"
)

// ServiceChooser enforces an arguments array have a cluster and service
// if it does not, it will search and prompt
func ServiceChooser(args []string) (string, string) {
	var cluster, service string
	if len(args) > 0 {
		cluster = args[0]
	}
	if len(args) > 1 {
		service = args[1]
	}
	if cluster == "" {
		names, err := ecs.GetClusterNames()
		failOnError(err, "Error finding clusters")
		cluster = StringChooser(names, "Please choose a cluster")
	}
	if service == "" {
		services, err := ecs.ListServices(cluster)
		service = StringChooser(services, "Please choose a service")
		failOnError(err, "Error finding services")
	}
	return cluster, service
}

// StringChooser asks a user to pick from a series of strings
func StringChooser(options []string, title string) string {
	if len(options) == 1 {
		return options[0]
	}
	for {
		fmt.Printf("%s:\n", title)
		reader := bufio.NewReader(os.Stdin)
		sort.Strings(options)
		for choice, disp := range options {
			fmt.Printf("[%d]: %s\n", choice+1, disp)
		}
		text, _ := reader.ReadString('\n')
		choice, err := strconv.Atoi(text[:len(text)-1])
		choice--
		if err == nil && choice >= 0 && choice < len(options) {
			return options[choice]
		}
	}
}

// AskForConfirmation asks the user for confirmation. A user must type in "yes" or "no" and
// then press enter. It has fuzzy matching, so "y", "Y", "yes", "YES", and "Yes" all count as
// confirmations. If the input is not recognized, it will ask again. The function does not return
// until it gets a valid response from the user.
func AskForConfirmation(s string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/n]: ", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}

// Validate2ArgumentsCount simply validates that there are two arguments
func Validate2ArgumentsCount(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("not enough arguments supplied")
	}
	return nil
}

// Validate4ArgumentsCount simply validates that there are two arguments
func Validate4ArgumentsCount(cmd *cobra.Command, args []string) error {
	if len(args) != 4 {
		return fmt.Errorf("not enough arguments supplied")
	}
	return nil
}

// EditStringBlock delegates the editing of a string block
// to an editor of choice, similar to git commit, or git rebase
func EditStringBlock(input string) (output string, err error) {
	tmpfile, err := ioutil.TempFile("", "ecsy")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpfile.Name())
	if _, err = tmpfile.Write([]byte(input)); err != nil {
		return "", err
	}
	cmd := exec.Command("vi", tmpfile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err = cmd.Start()
	if err != nil {
		return "", err
	}
	err = cmd.Wait()
	if err != nil {
		return "", err
	}
	if err = tmpfile.Close(); err != nil {
		return "", err
	}
	edited, err := ioutil.ReadFile(tmpfile.Name())
	if err != nil {
		return "", err
	}
	return string(edited), nil
}

func failOnError(err error, message string) {
	if err != nil {
		fmt.Printf("%s: %v\n", message, err)
		os.Exit(1)
	}
}
