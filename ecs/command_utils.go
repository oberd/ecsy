package ecs

import (
	"encoding/csv"
	"strings"
)

func parseCommandOverride(command string) ([]string, error) {
	reader := csv.NewReader(strings.NewReader(command))
	reader.Comma = ' '
	commands, err := reader.Read()
	if err != nil {
		return nil, err
	}
	return commands, nil
}
