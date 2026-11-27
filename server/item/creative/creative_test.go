package creative

import (
	"strings"
	"testing"

	"github.com/df-mc/dragonfly/server/world"
)

func TestVanillaItems(t *testing.T) {
	world.DefaultBlockRegistry.Finalize()
	groups, items := VanillaItems()
	if len(groups) == 0 || len(items) == 0 {
		t.Fatalf("empty vanilla catalog: %d groups, %d items", len(groups), len(items))
	}
	categories := make(map[string]Category, len(groups))
	for _, group := range groups {
		categories[group.Name] = group.Category
	}
	for _, entry := range items {
		block, ok := entry.Stack.Item().(world.Block)
		if !ok {
			continue
		}
		name, _ := block.EncodeBlock()
		if name == "minecraft:diamond_ore" {
			if got := categories[entry.Group]; got != NatureCategory() {
				t.Fatalf("diamond ore category = %d, want nature", got.Uint8())
			}
			return
		}
	}
	t.Fatal("diamond ore missing from vanilla creative catalog")
}

// TestVanillaItemsAnonymousGroupsKeepEmptyNames verifies that the vanilla
// catalog keeps empty group names empty instead of exposing synthetic anonN
// keys. Behavioural check: the catalog must still contain every group and item.
func TestVanillaItemsAnonymousGroupsKeepEmptyNames(t *testing.T) {
	world.DefaultBlockRegistry.Finalize()
	groups, items := VanillaItems()
	if len(groups) != 124 {
		t.Fatalf("group count = %d, want 124", len(groups))
	}
	if len(items) != 1562 {
		t.Fatalf("item count = %d, want 1562", len(items))
	}

	anonymous := map[int]bool{10: true, 13: true, 15: true, 17: true, 24: true, 25: true, 28: true, 33: true, 36: true, 41: true, 44: true, 46: true, 48: true, 51: true, 62: true, 64: true, 67: true, 69: true, 73: true, 75: true, 80: true, 81: true, 84: true, 87: true, 89: true, 91: true, 93: true, 95: true, 97: true, 99: true, 102: true, 105: true, 111: true, 113: true, 115: true, 123: true}
	if len(anonymous) != 36 {
		t.Fatalf("fixture has %d anonymous indices, want 36", len(anonymous))
	}

	empty, synthetic := 0, 0
	var firstSynthetic string
	for i, group := range groups {
		if group.Name == "" {
			empty++
			if !anonymous[i] {
				t.Errorf("group %d has an unexpected empty name", i)
			}
			continue
		}
		if strings.HasPrefix(group.Name, "anon") {
			synthetic++
			if firstSynthetic == "" {
				firstSynthetic = group.Name
			}
		}
	}
	if empty != 36 {
		t.Errorf("groups with empty name = %d, want 36", empty)
	}
	if synthetic != 0 {
		t.Errorf("groups with synthetic anon* name = %d, want 0 (first: %q)", synthetic, firstSynthetic)
	}
	for i := range anonymous {
		if groups[i].Name != "" {
			t.Errorf("anonymous group %d name = %q, want empty", i, groups[i].Name)
		}
	}
}
