package protocol

import (
	"fmt"
	"regexp"
	"strings"
)

type IRCMessage struct {
	Prefix  string
	Command string
	Params  []string
}

type ProtocolParser struct{}

func NewProtocolParser() *ProtocolParser {
	return &ProtocolParser{}
}

func (p *ProtocolParser) Parse(raw string) (*IRCMessage, error) {
	// Regex for matching the prefix (if present)
	prefixRegex := regexp.MustCompile(`^:([^ ]+) `)
	
	// Check for prefix
	var prefix string
	if strings.HasPrefix(raw, ":") {
		matches := prefixRegex.FindStringSubmatch(raw)
		if matches != nil {
			prefix = matches[1]
			raw = raw[len(matches[0]):]
		}
	}

	// Split the remaining message by spaces
	parts := strings.SplitN(raw, " ", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid message format")
	}
	command := parts[0]
	rawParams := parts[1]

	// Extract params and handle trailing params that start with ":"
	var params []string
	if strings.Contains(rawParams, " :") {
		parts := strings.SplitN(rawParams, " :", 2)
		params = append(strings.Split(parts[0], " "), parts[1])
	} else {
		params = strings.Split(rawParams, " ")
	}

	// Ensure the message ends with CRLF
	if !strings.HasSuffix(params[len(params)-1], "\r\n") {
		return nil, fmt.Errorf("message must end with CRLF")
	}

	// Clean up the CRLF from the last param
	params[len(params)-1] = strings.TrimSuffix(params[len(params)-1], "\r\n")

	return &IRCMessage{
		Prefix:  prefix,
		Command: strings.ToUpper(command),
		Params:  params,
	}, nil
}
