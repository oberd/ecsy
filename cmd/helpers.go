package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

// StringChooser asks a user to pick from a series of strings
func StringChooser(options []string, title string) string {
	if len(options) == 1 {
		return options[0]
	}
	for {
		fmt.Printf("%s:\n", title)
		reader := bufio.NewReader(os.Stdin)
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

// ValidateArgumentsCount2 simply validates that there are two arguments
func ValidateArgumentsCount2(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("Not enough arguemnts supplied! %v", args)
	}
	return nil
}
