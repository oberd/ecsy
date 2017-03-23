package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
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
