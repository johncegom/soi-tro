package ui

import "testing"

func historyKeys(dev bool) map[string]bool {
	keys := map[string]bool{}
	for _, o := range historyMenuOptions(dev) {
		keys[o.Value] = true
	}

	return keys
}

func TestHistoryMenuHidesSamplesUnlessDevMode(t *testing.T) {
	if historyKeys(false)["samples"] {
		t.Fatal("sample data entry must be hidden outside dev mode")
	}

	if !historyKeys(true)["samples"] {
		t.Fatal("sample data entry missing in dev mode")
	}
}

func TestHistoryMenuKeepsCoreEntries(t *testing.T) {
	for _, dev := range []bool{false, true} {
		keys := historyKeys(dev)
		for _, want := range []string{"compare", "list", "viewing", "delete", backChoice} {
			if !keys[want] {
				t.Fatalf("dev=%v: missing %q", dev, want)
			}
		}
	}
}

func TestDevModeFollowsEnv(t *testing.T) {
	t.Setenv("SOI_TRO_DEV", "1")
	if !devModeEnabled() {
		t.Fatal("SOI_TRO_DEV=1 should enable dev mode")
	}

	t.Setenv("SOI_TRO_DEV", "")
	if devModeEnabled() {
		t.Fatal("empty SOI_TRO_DEV should not enable dev mode")
	}
}
