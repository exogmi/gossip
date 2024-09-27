package protocol

import (
	"strings"
)

type Command struct {
	Name   string
	Params []string
}

type ProtocolParser struct{}

func NewProtocolParser() *ProtocolParser {
	return &ProtocolParser{}
}

func (p *ProtocolParser) Parse(message string) (*Command, error) {
	parts := strings.SplitN(strings.TrimSpace(message), " ", 2)
	command := &Command{
		Name: strings.ToUpper(parts[0]),
	}

	if len(parts) > 1 {
		// Check if there's a colon in the parameters
		colonIndex := strings.Index(parts[1], ":")
		if colonIndex != -1 {
			// Split the parameters before the colon
			command.Params = strings.Fields(parts[1][:colonIndex])
			// Add the rest as a single parameter (including the colon)
			command.Params = append(command.Params, parts[1][colonIndex:])
		} else {
			command.Params = strings.Fields(parts[1])
		}
	}

	return command, nil
}
