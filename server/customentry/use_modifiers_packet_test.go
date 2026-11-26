package customentry

import (
	"bytes"
	"image"
	"math"
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/category"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type registryTestItem struct {
	id       string
	duration time.Duration
	modifier float64
}

func (i registryTestItem) EncodeItem() (string, int16) { return i.id, 0 }
func (i registryTestItem) Name() string                { return i.id }
func (i registryTestItem) Texture() image.Image        { return image.NewNRGBA(image.Rect(0, 0, 1, 1)) }
func (i registryTestItem) Category() category.Category { return category.Items() }
func (i registryTestItem) AlwaysConsumable() bool      { return true }
func (i registryTestItem) ConsumeDuration() time.Duration {
	return i.duration
}
func (i registryTestItem) Consume(*world.Tx, item.Consumer) item.Stack { return item.Stack{} }
func (i registryTestItem) MovementModifier() float64                   { return i.modifier }

func registryNumber(t *testing.T, v any) float64 {
	t.Helper()
	switch x := v.(type) {
	case float64:
		return x
	case float32:
		return float64(x)
	case int32:
		return float64(x)
	case int:
		return float64(x)
	default:
		t.Fatalf("value %#v is not a number (%T)", v, v)
		return 0
	}
}

func registryUseModifiers(t *testing.T, entry protocol.ItemEntry) map[string]any {
	t.Helper()
	components, ok := entry.Data["components"].(map[string]any)
	if !ok {
		t.Fatalf("entry Data has no components map: %#v", entry.Data)
	}
	modifiers, ok := components["minecraft:use_modifiers"].(map[string]any)
	if !ok {
		t.Fatalf("minecraft:use_modifiers missing from registry entry: %#v", components)
	}
	return modifiers
}

// TestItemRegistryCarriesUseModifiers proves the component is present both in
// the protocol entry built for the item registry and after an actual packet
// marshal/unmarshal round trip, i.e. on the wire the client receives.
func TestItemRegistryCarriesUseModifiers(t *testing.T) {
	it := registryTestItem{"krep:minigun", 1500 * time.Millisecond, 0.2}
	entry := Item(it, 42)
	if entry.Name != "krep:minigun" || entry.RuntimeID != 42 {
		t.Fatalf("unexpected entry identity: %#v", entry)
	}
	if !entry.ComponentBased || entry.Version != protocol.ItemEntryVersionDataDriven {
		t.Fatalf("entry is not a component-based data-driven item: %#v", entry)
	}
	modifiers := registryUseModifiers(t, entry)
	if got := registryNumber(t, modifiers["movement_modifier"]); math.Abs(got-0.2) > 1e-9 {
		t.Errorf("entry movement_modifier = %v, want 0.2", got)
	}
	if got := registryNumber(t, modifiers["use_duration"]); math.Abs(got-1.5) > 1e-9 {
		t.Errorf("entry use_duration = %v, want 1.5 seconds", got)
	}

	pk := &packet.ItemRegistry{Items: []protocol.ItemEntry{entry}}
	buf := &bytes.Buffer{}
	pk.Marshal(protocol.NewWriter(buf, 0))
	if buf.Len() == 0 {
		t.Fatal("ItemRegistry marshal produced no bytes")
	}
	var decoded packet.ItemRegistry
	decoded.Marshal(protocol.NewReader(bytes.NewReader(buf.Bytes()), 0, true))
	if len(decoded.Items) != 1 {
		t.Fatalf("decoded %d items, want 1", len(decoded.Items))
	}
	wire := registryUseModifiers(t, decoded.Items[0])
	if got := registryNumber(t, wire["movement_modifier"]); math.Abs(got-0.2) > 1e-6 {
		t.Errorf("wire movement_modifier = %v, want 0.2", got)
	}
	if got := registryNumber(t, wire["use_duration"]); math.Abs(got-1.5) > 1e-6 {
		t.Errorf("wire use_duration = %v, want 1.5 seconds", got)
	}
}
