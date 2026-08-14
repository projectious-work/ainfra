// Package inventory validates standardized OpenTofu output and converts it to
// deterministic, closed Ansible inventory.
package inventory

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	"go.yaml.in/yaml/v3"
)

const maxOutputBytes = 16 << 20

var namePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,127}$`)
var userPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_.-]{0,63}$`)
var secretPattern = regexp.MustCompile(`(?i)(-----BEGIN|AKIA[0-9A-Z]{16}|gh[pousr]_[A-Za-z0-9_]{20,}|xox[baprs]-|eyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\.|://[^/@:]+:[^/@]+@)`)

// StandardOutput is the closed, versioned provider-to-configuration handoff.
type StandardOutput struct {
	SchemaVersion string          `json:"schema_version"`
	Hosts         map[string]Host `json:"hosts"`
}

// Host contains only inventory identity and closed connection facts.
type Host struct {
	Groups     []string   `json:"groups"`
	Connection Connection `json:"connection"`
}

// Connection is either local or a complete SSH connection.
type Connection struct {
	Type    string `json:"type"`
	Address string `json:"address,omitempty"`
	User    string `json:"user,omitempty"`
	Port    int    `json:"port,omitempty"`
}

// Parse strictly decodes and validates standardized output.
func Parse(contents []byte) (StandardOutput, error) {
	if len(contents) == 0 || len(contents) > maxOutputBytes {
		return StandardOutput{}, errors.New("standard output has invalid size")
	}
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.DisallowUnknownFields()
	var output StandardOutput
	if err := decoder.Decode(&output); err != nil {
		return StandardOutput{}, fmt.Errorf("decode standard output: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return StandardOutput{}, errors.New("standard output must contain exactly one JSON value")
	}
	if err := Validate(output); err != nil {
		return StandardOutput{}, err
	}
	return output, nil
}

// Validate enforces the v1 closed output contract.
func Validate(output StandardOutput) error {
	if output.SchemaVersion != "1" {
		return fmt.Errorf("unsupported standard output schema version %q", output.SchemaVersion)
	}
	if len(output.Hosts) == 0 || len(output.Hosts) > 10000 {
		return errors.New("standard output must contain 1 to 10000 hosts")
	}
	for name, host := range output.Hosts {
		if !namePattern.MatchString(name) || len(host.Groups) == 0 || len(host.Groups) > 256 {
			return fmt.Errorf("invalid inventory host %q", name)
		}
		seen := make(map[string]struct{}, len(host.Groups))
		for _, group := range host.Groups {
			if !namePattern.MatchString(group) {
				return fmt.Errorf("invalid inventory group %q", group)
			}
			if _, exists := seen[group]; exists {
				return fmt.Errorf("host %q contains duplicate group %q", name, group)
			}
			seen[group] = struct{}{}
		}
		switch host.Connection.Type {
		case "local":
			if host.Connection.Address != "" || host.Connection.User != "" || host.Connection.Port != 0 {
				return fmt.Errorf("local host %q contains SSH fields", name)
			}
		case "ssh":
			if host.Connection.Address == "" || host.Connection.User == "" ||
				host.Connection.Port < 1 || host.Connection.Port > 65535 {
				return fmt.Errorf("SSH host %q has incomplete connection fields", name)
			}
			if host.Connection.Address[0] == '-' || strings.ContainsAny(host.Connection.Address, " \t\r\n") ||
				len(host.Connection.Address) > 253 || !userPattern.MatchString(host.Connection.User) {
				return fmt.Errorf("SSH host %q has invalid connection fields", name)
			}
		default:
			return fmt.Errorf("host %q uses unsupported connection type %q", name, host.Connection.Type)
		}
		if secretPattern.MatchString(name + " " + strings.Join(host.Groups, " ") + " " + host.Connection.Address + " " + host.Connection.User) {
			return fmt.Errorf("host %q contains secret-shaped material", name)
		}
	}
	return nil
}

// YAML returns stable YAML with sorted groups, hosts, and map keys.
func YAML(output StandardOutput) ([]byte, error) {
	if err := Validate(output); err != nil {
		return nil, err
	}
	groups := make(map[string][]string)
	for hostName, host := range output.Hosts {
		for _, group := range host.Groups {
			groups[group] = append(groups[group], hostName)
		}
	}
	groupNames := sortedKeys(groups)
	root := mapping()
	for _, group := range groupNames {
		hosts := mapping()
		sort.Strings(groups[group])
		for _, hostName := range groups[group] {
			host := output.Hosts[hostName]
			variables := mapping()
			appendScalar(variables, "ansible_connection", host.Connection.Type)
			if host.Connection.Type == "ssh" {
				appendScalar(variables, "ansible_host", host.Connection.Address)
				appendScalar(variables, "ansible_user", host.Connection.User)
				appendInt(variables, "ansible_port", host.Connection.Port)
			}
			hosts.Content = append(hosts.Content, scalar(hostName), variables)
		}
		children := mapping()
		children.Content = append(children.Content, scalar("hosts"), hosts)
		root.Content = append(root.Content, scalar(group), children)
	}
	document := &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{root}}
	var buffer bytes.Buffer
	encoder := yaml.NewEncoder(&buffer)
	encoder.SetIndent(2)
	if err := encoder.Encode(document); err != nil {
		return nil, err
	}
	return buffer.Bytes(), encoder.Close()
}

func sortedKeys(values map[string][]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func mapping() *yaml.Node { return &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"} }
func scalar(value string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value}
}
func appendScalar(node *yaml.Node, key, value string) {
	node.Content = append(node.Content, scalar(key), scalar(value))
}
func appendInt(node *yaml.Node, key string, value int) {
	node.Content = append(node.Content, scalar(key), &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: fmt.Sprint(value)})
}
