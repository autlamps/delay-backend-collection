package realtime

import (
	"reflect"
	"testing"
)

func TestVLEntities_ToMap(t *testing.T) {

	vl1 := VLEntity{
		"1234",
		false,
		VLVehicle{VLTrip{"1234", "1234", "1234", 1},
			VLVehicleID{"1234"},
			VLPosition{1.23, 2.14},
			-1.234,
		}}

	vl2 := VLEntity{
		"5678",
		false,
		VLVehicle{VLTrip{"5678", "5678", "5678", 1},
			VLVehicleID{"5678"},
			VLPosition{1.23, 2.14},
			-1.234,
		}}

	vles := VLEntities{vl1, vl2}
	vlm := make(map[string]VLEntity, 1)

	vlm["1234"] = vl1
	vlm["5678"] = vl2

	tvlm := vles.ToMap()

	if !reflect.DeepEqual(vlm, tvlm) {
		t.Errorf("Vehicle.ToMap not the same as normal map! Expected %v, got %v", vlm, tvlm)
	}
}
