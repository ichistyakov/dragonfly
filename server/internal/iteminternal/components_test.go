package iteminternal

import (
	"image"
	"math"
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/category"
	"github.com/df-mc/dragonfly/server/world"
)

// testConsumable is the minimal custom item that reaches the Consumable branch
// of Components. testModifiedItem adds the optional A56 interface on top.
type testConsumable struct {
	id       string
	duration time.Duration
}

func (i testConsumable) EncodeItem() (string, int16) { return i.id, 0 }
func (i testConsumable) Name() string                { return i.id }
func (i testConsumable) Texture() image.Image        { return image.NewNRGBA(image.Rect(0, 0, 1, 1)) }
func (i testConsumable) Category() category.Category { return category.Items() }
func (i testConsumable) AlwaysConsumable() bool      { return true }
func (i testConsumable) ConsumeDuration() time.Duration {
	return i.duration
}
func (i testConsumable) Consume(*world.Tx, item.Consumer) item.Stack { return item.Stack{} }

type testModifiedItem struct {
	testConsumable
	modifier float64
}

func (i testModifiedItem) MovementModifier() float64 { return i.modifier }

func componentsOf(t *testing.T, it world.CustomItem) map[string]any {
	t.Helper()
	data := Components(it)
	components, ok := data["components"].(map[string]any)
	if !ok {
		t.Fatalf("components missing or wrong type in %#v", data)
	}
	return components
}

func TestUseModifiersComponent(t *testing.T) {
	tests := []struct {
		name          string
		item          world.CustomItem
		want          float64
		wantComponent bool
	}{
		{"minigun", testModifiedItem{testConsumable{"krep:minigun", 1500 * time.Millisecond}, 0.2}, 0.2, true},
		{"akm", testModifiedItem{testConsumable{"krep:akm", 2 * time.Second}, 0.7}, 0.7, true},
		{"mp7", testModifiedItem{testConsumable{"krep:mp7", time.Second}, 1.0}, 1.0, true},
		{"empty", testModifiedItem{testConsumable{"krep:akm_emp", time.Second}, 1.0}, 1.0, true},
		{"clamp high", testModifiedItem{testConsumable{"krep:bad", time.Second}, 1.4}, 1.0, true},
		{"clamp low", testModifiedItem{testConsumable{"krep:bad", time.Second}, -0.3}, 0, true},
		{"clamp nan", testModifiedItem{testConsumable{"krep:bad", time.Second}, math.NaN()}, 0, true},
		{"no interface", testConsumable{"krep:plain", time.Second}, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			components := componentsOf(t, tt.item)
			raw, ok := components["minecraft:use_modifiers"]
			if !ok {
				if tt.wantComponent {
					t.Fatalf("minecraft:use_modifiers missing; components=%#v", components)
				}
				return
			}
			if !tt.wantComponent {
				t.Fatalf("minecraft:use_modifiers must not be added without MovementModifier(): %#v", raw)
			}
			modifiers, ok := raw.(map[string]any)
			if !ok {
				t.Fatalf("minecraft:use_modifiers is %T, want map[string]any", raw)
			}
			if got := modifiers["movement_modifier"]; got != tt.want {
				t.Errorf("movement_modifier = %#v, want %v", got, tt.want)
			}
			if got := modifiers["use_duration"]; got != tt.item.(interface{ ConsumeDuration() time.Duration }).ConsumeDuration().Seconds() {
				t.Errorf("use_duration = %#v, want seconds", got)
			}

			// Legacy tick property and the food component must stay intact.
			props := components["item_properties"].(map[string]any)
			if got := props["use_duration"]; got != int32(tt.item.(interface{ ConsumeDuration() time.Duration }).ConsumeDuration().Seconds()*20) {
				t.Errorf("legacy use_duration = %#v, want ticks", got)
			}
			if _, ok := components["minecraft:food"]; !ok {
				t.Error("minecraft:food component lost")
			}
		})
	}
}

func TestUseModifiersSecondsNotTicks(t *testing.T) {
	// A 1500 ms consume time must appear as 1.5 (seconds) in the component and
	// as 30 (ticks) in the legacy property.
	it := testModifiedItem{testConsumable{"krep:minigun", 1500 * time.Millisecond}, 0.2}
	components := componentsOf(t, it)
	modifiers := components["minecraft:use_modifiers"].(map[string]any)
	if modifiers["use_duration"] != 1.5 {
		t.Fatalf("component use_duration = %#v, want 1.5 seconds", modifiers["use_duration"])
	}
	props := components["item_properties"].(map[string]any)
	if props["use_duration"] != int32(30) {
		t.Fatalf("legacy use_duration = %#v, want 30 ticks", props["use_duration"])
	}
}
