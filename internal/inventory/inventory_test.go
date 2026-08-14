package inventory

import (
	"strings"
	"testing"
)

func TestParseAndYAMLDeterministic(t *testing.T) {
	contents := []byte(`{"schema_version":"1","hosts":{"z":{"groups":["web"],"connection":{"type":"ssh","address":"10.0.0.2","user":"ops","port":22}},"a":{"groups":["web","control"],"connection":{"type":"local"}}}}`)
	output, err := Parse(contents)
	if err != nil {
		t.Fatal(err)
	}
	first, err := YAML(output)
	if err != nil {
		t.Fatal(err)
	}
	second, err := YAML(output)
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Fatal("inventory serialization is not deterministic")
	}
	want := "control:\n  hosts:\n    a:\n      ansible_connection: local\nweb:\n  hosts:\n    a:\n      ansible_connection: local\n    z:\n      ansible_connection: ssh\n      ansible_host: 10.0.0.2\n      ansible_user: ops\n      ansible_port: 22\n"
	if string(first) != want {
		t.Fatalf("inventory mismatch:\n%s", first)
	}
}

func TestParseRejectsUnknownAndUnsafeFields(t *testing.T) {
	cases := []string{
		`{"schema_version":"2","hosts":{"a":{"groups":["g"],"connection":{"type":"local"}}}}`,
		`{"schema_version":"1","hosts":{"a":{"groups":["bad group"],"connection":{"type":"local"}}}}`,
		`{"schema_version":"1","hosts":{"a":{"groups":["g"],"connection":{"type":"local","password":"secret"}}}}`,
		`{"schema_version":"1","hosts":{"a":{"groups":["g"],"connection":{"type":"ssh","address":"x","user":"u","port":0}}}}`,
		`{"schema_version":"1","hosts":{"a":{"groups":["g"],"connection":{"type":"ssh","address":"ghp_abcdefghijklmnopqrstuvwxyz","user":"ops","port":22}}}}`,
	}
	for _, value := range cases {
		if _, err := Parse([]byte(value)); err == nil {
			t.Fatalf("accepted invalid output %s", strings.TrimSpace(value))
		}
	}
}
