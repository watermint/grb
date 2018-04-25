package grb

import (
	"reflect"
	"testing"
	"github.com/mitchellh/go-mruby"
	"encoding/json"
	"fmt"
)

func floatEquals(a, b float64) bool {
	var EPSILON = 0.000001
	return (a-b) < EPSILON && (b-a) < EPSILON
}

func TestDecodeMrbValue(t *testing.T) {
	mrb := mruby.NewMrb()
	defer mrb.Close()

	if rv, err := mrb.LoadString(`nil`); err != nil {
		t.Error(err)
	} else {
		if dv, err := DecodeMrbValue(rv); err != nil {
			t.Error(err)
		} else {
			if dv != nil {
				t.Error("invalid type")
			}
		}
	}

	if rv, err := mrb.LoadString(`true`); err != nil {
		t.Error(err)
	} else {
		if dv, err := DecodeMrbValue(rv); err != nil {
			t.Error(err)
		} else {
			if dv != true {
				t.Error("invalid type")
			}
		}
	}

	if rv, err := mrb.LoadString(`false`); err != nil {
		t.Error(err)
	} else {
		if dv, err := DecodeMrbValue(rv); err != nil {
			t.Error(err)
		} else {
			if dv != false {
				t.Error("invalid type")
			}
		}
	}

	if rv, err := mrb.LoadString(`1`); err != nil {
		t.Error(err)
	} else {
		if dv, err := DecodeMrbValue(rv); err != nil {
			t.Error(err)
		} else {
			if dv != 1 {
				t.Error("invalid type")
			}
		}
	}

	if rv, err := mrb.LoadString(`1.0`); err != nil {
		t.Error(err)
	} else {
		if dv, err := DecodeMrbValue(rv); err != nil {
			t.Error(err)
		} else {
			if dv != 1.0 {
				t.Error("invalid type")
			}
		}
	}

	if rv, err := mrb.LoadString(`"water"`); err != nil {
		t.Error(err)
	} else {
		if dv, err := DecodeMrbValue(rv); err != nil {
			t.Error(err)
		} else {
			if dv != "water" {
				t.Error("invalid type")
			}
		}
	}

	if rv, err := mrb.LoadString(`["water", "mint"]`); err != nil {
		t.Error(err)
	} else {
		if dv, err := DecodeMrbValue(rv); err != nil {
			t.Error(err)
		} else {
			switch dvv := dv.(type) {
			case []interface{}:
				switch dvv0 := dvv[0].(type) {
				case string:
					if dvv0 != "water" {
						t.Error("invalid value")
					} else {
						switch dvv1 := dvv[1].(type) {
						case string:
							if dvv1 != "mint" {
								t.Error("invalid value")
							}
						default:
							t.Error("Invalid type")
						}
					}
				default:
					t.Error("Invalid type")
				}
			default:
				t.Error("Invalid type")
			}
		}
	}
}

func TestEncodeMrbValue(t *testing.T) {
	mrb := mruby.NewMrb()
	defer mrb.Close()

	if ev, err := EncodeMrbValue(mrb, 123); err != nil {
		t.Error(err)
	} else {
		if ev.Type() != mruby.TypeFixnum || ev.Fixnum() != 123 {
			t.Error("Invalid value")
		}
	}

	if ev, err := EncodeMrbValue(mrb, 3.14); err != nil {
		t.Error(err)
	} else {
		if ev.Type() != mruby.TypeFloat || !floatEquals(ev.Float(), 3.14) {
			t.Errorf("Invalid value : Type:%d, Value:%f", ev.Type(), ev.Float())
		}
	}

	if ev, err := EncodeMrbValue(mrb, "water"); err != nil {
		t.Error(err)
	} else {
		if ev.Type() != mruby.TypeString || ev.String() != "water" {
			t.Error("Invalid value")
		}
	}

	if ev, err := EncodeMrbValue(mrb, "123"); err != nil {
		t.Error(err)
	} else {
		if ev.Type() != mruby.TypeString || ev.String() != "123" {
			t.Error("Invalid value")
		}
	}

	if ev, err := EncodeMrbValue(mrb, []string{"water", "mint"}); err != nil {
		t.Error(err)
	} else {
		if ev.Type() != mruby.TypeArray {
			t.Error("Invalid type")
		} else {
			ar := ev.Array()
			if ar.Len() != 2 {
				t.Error("Invalid value")
			}
			if ar0, err := ar.Get(0); err != nil {
				t.Error(err)
			} else if ar0.String() != "water" {
				t.Error("Invalid value")
			}
			if ar1, err := ar.Get(1); err != nil {
				t.Error(err)
			} else if ar1.String() != "mint" {
				t.Error("Invalid value")
			}

		}
	}

	if ev, err := EncodeMrbValue(mrb, map[string]interface{}{"Earth": 6371, "Moon": 1737}); err != nil {
		t.Error(err)
	} else {
		if ev.Type() != mruby.TypeHash {
			t.Error("invalid type")
		} else {
			ha := ev.Hash()
			if he, err := ha.Get(mrb.StringValue("Earth")); err != nil {
				t.Error("Missing key")
			} else if he.Type() != mruby.TypeFixnum || he.Fixnum() != 6371 {
				t.Error("Invalid value")
			}
			if he, err := ha.Get(mrb.StringValue("Moon")); err != nil {
				t.Error("Missing key")
			} else if he.Type() != mruby.TypeFixnum || he.Fixnum() != 1737 {
				t.Error("Invalid value")
			}
		}
	}
}

func TestEncodeDecodeMrbValue(t *testing.T) {
	mrb := mruby.NewMrb()
	defer mrb.Close()

	samples := []string{
		`{}`,
		`[]`,
		`123`,
		`"Planet"`,
		`null`,
		`true`,
		`false`,
		`{"Earth":6371,"Moon":1737}`,
		`{"Earth":{"Radius":6371,"HasMoon":true}, "Venus":{"Radius":6051, "HasMoon":false}}`,
		`{"Earth":{"Radius":6371,"Moon":[{"Moon":{"Radius":1737}}]}, "Venus":{"Radius":6051, "Moon":null}}`,
		`[{"Name":"Earth","Radius":6371},{"Name":"Mars","Radius":3389}]`,
		`[{"Name":"地球","Radius":6371},{"Name":"火星","Radius":3389}]`,
	}

	for _, sample := range samples {
		var d interface{}
		if err := json.Unmarshal([]byte(sample), &d); err != nil {
			t.Errorf("Broken test data json json[%s]", sample)
			continue
		}
		rv, err := EncodeMrbValue(mrb, d)
		if err != nil {
			t.Errorf("Unable to encode value json[%s]", sample)
		}
		gv, err := DecodeMrbValue(rv)
		if err != nil {
			t.Errorf("Unable to decode value json[%s]", sample)
		}
		if !reflect.DeepEqual(d, gv) {
			t.Errorf("Values not mached Orig[%s] Decoded[%s]", d, gv)
		}
	}
}

func TestUnmarshal(t *testing.T) {
	mrb := mruby.NewMrb()
	defer mrb.Close()

	_, err := mrb.LoadString(`
class Planet
  attr_accessor :id           # int
  attr_accessor :name         # string
  attr_accessor :radius       # float
  attr_accessor :has_moon     # bool
  attr_accessor :moon         # []string
  attr_accessor :moon_radius  # map[string]int
end
`)
	if err != nil {
		t.Error(err)
	}

	type Planet struct {
		Id         int            `mruby:"id"`
		Name       string         `mruby:"name"`
		Radius     float64        `mruby:"radius"`
		HasMoon    bool           `mruby:"has_moon"`
		Moon       []string       `mruby:"moon"`
		MoonRadius map[string]int `mruby:"moon_radius"`
	}

	v, err := mrb.LoadString(`
d = Planet.new
d.id          = 3
d.name        = "Earth"
d.radius      = 6371.0
d.has_moon    = true
d.moon        = ["Moon"]
d.moon_radius = {"Moon" => 1737}
d
`)
	if err != nil {
		t.Error(err)
	}

	d := Planet{}
	err = Unmarshal(v, &d)
	fmt.Printf("Planet: %v\n", d)
	if err != nil {
		t.Error(err)
	}
	if d.Id != 3 {
		t.Errorf("Invalid value: %d", d.Id)
	}
	if d.Name != "Earth" {
		t.Errorf("Invalid value: %s", d.Name)
	}
	if !floatEquals(d.Radius, 6371.0) {
		t.Errorf("Invalid value: %f", d.Radius)
	}
	if !d.HasMoon {
		t.Errorf("Invalid value: %t", d.HasMoon)
	}
	if !reflect.DeepEqual(d.Moon, []string{"Moon"}) {
		t.Errorf("Invalid value: %s", d.Moon)
	}
	if !reflect.DeepEqual(d.MoonRadius, map[string]int{"Moon": 1737}) {
		t.Errorf("Invalid value: %v", d.MoonRadius)
	}

	v, err = mrb.LoadString(`
d = Planet.new
d.id          = 4
d.name        = "火星"
d.radius      = 3389.5
d.has_moon    = false
d.moon        = []
d.moon_radius = {}
d
`)
	if err != nil {
		t.Error(err)
	}
	err = Unmarshal(v, &d)
	if err != nil {
		t.Error(err)
	}
	if d.Id != 4 {
		t.Errorf("Invalid value: %d", d.Id)
	}
	if d.Name != "火星" {
		t.Errorf("Invalid value: %s", d.Name)
	}
	if !floatEquals(d.Radius, 3389.5) {
		t.Errorf("Invalid value: %f", d.Radius)
	}
	if d.HasMoon {
		t.Errorf("Invalid value: %t", d.HasMoon)
	}
	if !reflect.DeepEqual(d.Moon, []string{}) {
		t.Errorf("Invalid value: %s", d.Moon)
	}
	if !reflect.DeepEqual(d.MoonRadius, map[string]int{}) {
		t.Errorf("Invalid value: %v", d.MoonRadius)
	}

	{
mrb := mruby.NewMrb()
defer mrb.Close()

type Planet struct {
	Id   int
	Name string
}

rv, err := mrb.LoadString(`
class Planet
  attr_accessor :id
  attr_accessor :name
end

p = Planet.new
p.id   = 3
p.name = "Earth"
p
`)
if err != nil {
	panic(err.Error())
}

p := Planet{}
if rId, err := rv.Call("id"); err != nil {
	panic(err.Error())
} else {
	p.Id = rId.Fixnum()
}
if rName, err := rv.Call("name"); err != nil {
	panic(err.Error())
} else {
	p.Name = rName.String()
}
fmt.Printf("Planet: %v\n", p)
	}
}
