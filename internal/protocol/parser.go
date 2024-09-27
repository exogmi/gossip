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
		command.Params = strings.Split(parts[1], " ")
	}

	return command, nil
}
