package session

import (
	"strings"
	"testing"

	"github.com/df-mc/dragonfly/server/item/creative"
	"github.com/df-mc/dragonfly/server/world"
)

// TestCreativeContentPreservesAnonymousGroupIndices checks the exact packet
// payload produced for the client: anonymous group names must stay empty and
// their items must keep their own group_index instead of all collapsing into
// the first anonymous group (index 10).
func TestCreativeContentPreservesAnonymousGroupIndices(t *testing.T) {
	world.DefaultBlockRegistry.Finalize()
	if len(creative.Groups()) == 0 {
		groups, items := creative.VanillaItems()
		for _, group := range groups {
			creative.RegisterGroup(group)
		}
		for _, i := range items {
			creative.RegisterItem(i)
		}
	}

	groups, items := creativeContent(world.DefaultBlockRegistry)
	if len(groups) != 124 {
		t.Fatalf("packet group count = %d, want 124", len(groups))
	}
	if len(items) != 1562 {
		t.Fatalf("packet item count = %d, want 1562", len(items))
	}

	anonymous := map[uint32]bool{10: true, 13: true, 15: true, 17: true, 24: true, 25: true, 28: true, 33: true, 36: true, 41: true, 44: true, 46: true, 48: true, 51: true, 62: true, 64: true, 67: true, 69: true, 73: true, 75: true, 80: true, 81: true, 84: true, 87: true, 89: true, 91: true, 93: true, 95: true, 97: true, 99: true, 102: true, 105: true, 111: true, 113: true, 115: true, 123: true}
	if len(anonymous) != 36 {
		t.Fatalf("fixture has %d anonymous indices, want 36", len(anonymous))
	}

	empty, synthetic := 0, 0
	var firstSynthetic string
	for i, group := range groups {
		if group.Name == "" {
			empty++
			if !anonymous[uint32(i)] {
				t.Errorf("packet group %d has an unexpected empty name", i)
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
		t.Errorf("packet groups with empty name = %d, want 36", empty)
	}
	if synthetic != 0 {
		t.Errorf("packet groups with synthetic anon* name = %d, want 0 (first: %q)", synthetic, firstSynthetic)
	}

	total := 0
	perGroup := make(map[uint32]int, len(anonymous))
	for _, i := range items {
		if anonymous[i.GroupIndex] {
			total++
			perGroup[i.GroupIndex]++
		}
	}
	if total != 223 {
		t.Errorf("packet items in anonymous groups = %d, want 223", total)
	}
	// All 36 anonymous groups are present in the packet, but only 32 of them
	// have items that survive registry resolution (groups 33, 69, 73 and 95
	// contribute none), so 223 packet items occupy 32 distinct indices.
	if len(perGroup) != 32 {
		t.Errorf("anonymous group indices used by packet items = %d, want 32 (distribution: %v)", len(perGroup), perGroup)
	}
	if perGroup[10] == 223 {
		t.Errorf("all %d anonymous items collapsed into group index 10", perGroup[10])
	}
}
