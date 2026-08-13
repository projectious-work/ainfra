package project

import "testing"

func FuzzDecodeManifest(f *testing.F) {
	f.Add([]byte(`apiVersion: ainfra.projectious.work/v1
kind: Deployment
metadata:
  name: fuzz-example
spec:
  template:
    source: local:../template
`))
	f.Add([]byte("apiVersion: [\x00\xff\n---\nkind: Deployment\n"))
	f.Fuzz(func(t *testing.T, contents []byte) {
		_, _ = decodeManifest(contents)
	})
}
