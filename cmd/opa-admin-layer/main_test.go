package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunRenderWithOutDirWritesPlanTree(t *testing.T) {
	specPath := filepath.Join(t.TempDir(), "spec.json")
	spec := `{
  "name": "demo",
  "controlPlane": {
    "baseServiceURL": "https://control.example.com"
  },
  "tenants": [
    {
      "name": "tenant-a",
      "topics": [
        {
          "name": "billing"
        }
      ]
    }
  ]
}`
	if err := os.WriteFile(specPath, []byte(spec), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	stdoutPath := filepath.Join(t.TempDir(), "plan.stdout.json")
	outDir := filepath.Join(t.TempDir(), "tree")
	if err := run([]string{"render", "-input", specPath, "-output", stdoutPath, "-outdir", outDir}); err != nil {
		t.Fatalf("run render: %v", err)
	}

	for _, path := range []string{
		stdoutPath,
		filepath.Join(outDir, "plan.json"),
		filepath.Join(outDir, "tenant-a", "billing", "opa-config.yaml"),
		filepath.Join(outDir, "tenant-a", "billing", "deployment.yaml"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s to exist: %v", path, err)
		}
	}
}
