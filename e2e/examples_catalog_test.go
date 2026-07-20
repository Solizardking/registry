//go:build e2e

// TestExamplesCatalogApply registers every Agent-bearing example under
// examples/ via the real arctl apply path and asserts each agent is
// queryable with arctl get. Catalog registration only — no deploy/runtime.
package e2e

import (
	"path/filepath"
	"strings"
	"testing"
)

// exampleAgent is an agent identity expected after applying an example file.
type exampleAgent struct {
	name string
	tag  string
}

// TestExamplesCatalogApply applies the repo's example manifests against a live
// registry and verifies the example Agents are registered and gettable.
//
// Order matches examples/README.md dependency-safe apply order.
func TestExamplesCatalogApply(t *testing.T) {
	regURL := RegistryURL(t)
	// e2e runs with cwd = package dir (e2e/); examples/ is one level up.
	examplesDir, err := filepath.Abs(filepath.Join("..", "examples"))
	if err != nil {
		t.Fatalf("resolve examples dir: %v", err)
	}

	// Files in dependency-safe order. Metaplex MCP stubs must precede the agent.
	applyFiles := []string{
		filepath.Join(examplesDir, "agent.yaml"),
		filepath.Join(examplesDir, "full-stack.yaml"),
		filepath.Join(examplesDir, "clawd-perps-agent.yaml"),
		filepath.Join(examplesDir, "clawd-perps-mcp", "mcp.yaml"),
		filepath.Join(examplesDir, "clawd-perps-agent-project", "agent.yaml"),
		filepath.Join(examplesDir, "metaplex-mcp-stubs.yaml"),
		filepath.Join(examplesDir, "metaplex-agent.yaml"),
	}

	wantAgents := []exampleAgent{
		{name: "summarizer", tag: "latest"},
		{name: "summarizer", tag: "stable"},
		{name: "clawd-perps-trader", tag: "stable"},
		{name: "metaplex-trader", tag: "stable"},
	}

	for _, path := range applyFiles {
		t.Run("apply_"+filepath.Base(path), func(t *testing.T) {
			result := RunArctl(t, "", "apply", "-f", path, "--registry-url", regURL)
			RequireSuccess(t, result)
			// Apply may report created / configured / unchanged; never a hard fail.
			if strings.Contains(result.Stdout+result.Stderr, "failed") {
				// Soft failures print "failed:" per resource; RequireSuccess already
				// gates exit code, but double-check no resource-level fail marker.
				combined := result.Stdout + result.Stderr
				if strings.Contains(combined, "✗") {
					t.Fatalf("apply reported resource failure for %s:\n%s\n%s",
						path, result.Stdout, result.Stderr)
				}
			}
		})
	}

	for _, agent := range wantAgents {
		t.Run("get_"+agent.name+"_"+agent.tag, func(t *testing.T) {
			result := RunArctl(t, "", "get", "agent", agent.name,
				"--tag", agent.tag, "-o", "yaml", "--registry-url", regURL)
			RequireSuccess(t, result)
			RequireOutputContains(t, result, "kind: Agent")
			RequireOutputContains(t, result, "name: "+agent.name)
			// Tag appears under metadata; accept either yaml form.
			if !strings.Contains(result.Stdout, "tag: "+agent.tag) {
				t.Fatalf("expected tag %q in get output for %s:\n%s",
					agent.tag, agent.name, result.Stdout)
			}
			// HTTP path must also resolve (same live registry, not a reimplementation).
			verifyAgentExists(t, regURL, agent.name, agent.tag)
		})
	}

	t.Run("list_contains_example_agents", func(t *testing.T) {
		result := RunArctl(t, "", "get", "agents", "--registry-url", regURL)
		RequireSuccess(t, result)
		for _, agent := range wantAgents {
			RequireOutputContains(t, result, agent.name)
		}
	})
}
