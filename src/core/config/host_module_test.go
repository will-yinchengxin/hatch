package config

import (
	"encoding/json"
	"testing"
)

func TestHostModuleDefaultsAndNormalization(t *testing.T) {
	modules := DefaultModules()
	if err := json.Unmarshal([]byte(`{"workbench":false,"ai":false,"vpn":false,"memo":false}`), &modules); err != nil {
		t.Fatal(err)
	}
	if !modules.Host || modules.Normalized().Workbench {
		t.Fatalf("old config must show host without activating workbench: %+v", modules)
	}
	modules.Host = false
	if !modules.Normalized().Workbench {
		t.Fatal("all disabled modules must fall back to workbench")
	}
}
