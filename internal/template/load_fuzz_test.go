package template

import "testing"

func FuzzDecodeManifest(f *testing.F) {
	f.Add([]byte(`apiVersion: ainfra.projectious.work/v1
kind: Template
metadata:
  name: fuzz-template
  version: 1.0.0
spec:
  engines:
    tofu:
      directory: tofu
      version: ">= 1.9.0"
  outputs:
    inventory: none
`))
	f.Add([]byte("kind: Template\nunknown: &recursive [*recursive]\n"))
	f.Fuzz(func(t *testing.T, contents []byte) {
		_, _ = decode(contents)
	})
}
