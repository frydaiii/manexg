package vietnam

import (
	"github.com/banbox/banexg/utils"
	"path/filepath"
	"testing"
)

func TestVietnamNew(t *testing.T) {
	options := map[string]interface{}{
		"consumerID":     "test-consumer-id",
		"consumerSecret": "test-consumer-secret",
		"sandbox":        true,
	}

	exg, err := New(options)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if exg == nil {
		t.Fatal("expected exchange instance, got nil")
	}

	if exg.ExgInfo.ID != "vietnam" {
		t.Errorf("expected ID='vietnam', got '%s'", exg.ExgInfo.ID)
	}

	if exg.ExgInfo.Name != "Vietnam Stock Market" {
		t.Errorf("expected Name='Vietnam Stock Market', got '%s'", exg.ExgInfo.Name)
	}
}

func loadTestConfig() (map[string]interface{}, error) {
	configPath := filepath.Join(".", "local.json")
	var config map[string]interface{}
	err := utils.ReadJsonFile(configPath, &config, utils.JsonNumDefault)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func skipIfNoCredentials(t *testing.T) map[string]interface{} {
	config, err := loadTestConfig()
	if err != nil {
		t.Skipf("Skipping test: local.json not found or invalid: %v", err)
	}
	return config
}

func TestVietnamAuth(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := skipIfNoCredentials(t)

	exg, err := New(config)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if exg.AccessToken == "" {
		t.Error("expected access token after Init(), got empty string")
	}

	if exg.TokenExpiry == 0 {
		t.Error("expected token expiry to be set after Init(), got 0")
	}

	t.Logf("Successfully authenticated. Token expiry: %d", exg.TokenExpiry)
}
