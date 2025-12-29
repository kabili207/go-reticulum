package rns

import (
	"os"
	"path/filepath"
	"testing"

	configobj "github.com/svanichkin/configobj"
)

func TestDaemonManagement_HaltResumeReload(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config")

	cfgText := `
[interfaces]
  [[TestIF]]
    enabled = yes
    type = UDPInterface
    listen_ip = 127.0.0.1
    listen_port = 0
    forward_ip = 127.0.0.1
    forward_port = 0
`
	if err := os.WriteFile(configPath, []byte(cfgText), 0o600); err != nil {
		t.Fatalf("WriteFile(config): %v", err)
	}
	cfg, err := configobj.Load(configPath)
	if err != nil {
		t.Fatalf("Load(config): %v", err)
	}

	prev := Interfaces
	Interfaces = nil
	t.Cleanup(func() {
		for _, ifc := range Interfaces {
			if ifc != nil {
				ifc.Detach()
			}
		}
		Interfaces = prev
	})

	r := &Reticulum{
		Config:        cfg,
		ConfigPath:    configPath,
		InterfacePath: filepath.Join(dir, "interfaces"),
	}

	if err := r.bringUpSystemInterfaces(); err != nil {
		t.Fatalf("bringUpSystemInterfaces: %v", err)
	}
	if len(Interfaces) != 1 || Interfaces[0] == nil {
		t.Fatalf("expected 1 started interface, got %#v", Interfaces)
	}
	first := Interfaces[0]

	if err := r.HaltInterface("TestIF"); err != nil {
		t.Fatalf("HaltInterface: %v", err)
	}
	if len(Interfaces) != 0 {
		t.Fatalf("expected 0 interfaces after halt, got %d", len(Interfaces))
	}

	if err := r.ResumeInterface("TestIF"); err != nil {
		t.Fatalf("ResumeInterface: %v", err)
	}
	if len(Interfaces) != 1 || Interfaces[0] == nil {
		t.Fatalf("expected 1 interface after resume, got %#v", Interfaces)
	}
	second := Interfaces[0]
	if second == first {
		t.Fatalf("expected resumed interface to be a new instance")
	}

	if err := r.ReloadInterface("TestIF"); err != nil {
		t.Fatalf("ReloadInterface: %v", err)
	}
	if len(Interfaces) != 1 || Interfaces[0] == nil {
		t.Fatalf("expected 1 interface after reload, got %#v", Interfaces)
	}
	third := Interfaces[0]
	if third == second {
		t.Fatalf("expected reloaded interface to be a new instance")
	}
}
