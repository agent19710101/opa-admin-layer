package admin

import (
	"fmt"
	"os"
	"path/filepath"
)

// WritePlanTree materializes a rendered plan and per-topic YAML artifacts.
func WritePlanTree(plan Plan, outDir string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	encoded, err := MarshalPlan(plan)
	if err != nil {
		return fmt.Errorf("encode plan: %w", err)
	}
	if err := os.WriteFile(filepath.Join(outDir, "plan.json"), append(encoded, '\n'), 0o644); err != nil {
		return fmt.Errorf("write plan.json: %w", err)
	}

	for _, tenant := range plan.Tenants {
		for _, topic := range tenant.Topics {
			topicDir := filepath.Join(outDir, tenant.Name, topic.Name)
			if err := os.MkdirAll(topicDir, 0o755); err != nil {
				return fmt.Errorf("create topic directory %q: %w", topicDir, err)
			}
			if err := os.WriteFile(filepath.Join(topicDir, "opa-config.yaml"), []byte(topic.OPAConfigYAML), 0o644); err != nil {
				return fmt.Errorf("write opa-config.yaml for %s/%s: %w", tenant.Name, topic.Name, err)
			}
			if err := os.WriteFile(filepath.Join(topicDir, "configmap.yaml"), []byte(topic.ConfigMapManifestYAML), 0o644); err != nil {
				return fmt.Errorf("write configmap.yaml for %s/%s: %w", tenant.Name, topic.Name, err)
			}
			if err := os.WriteFile(filepath.Join(topicDir, "deployment.yaml"), []byte(topic.DeploymentManifestYAML), 0o644); err != nil {
				return fmt.Errorf("write deployment.yaml for %s/%s: %w", tenant.Name, topic.Name, err)
			}
		}
	}

	return nil
}
