# rgb

Go/mruby utilities.

# Mapping values

The mapping between mruby and Go values. 

## Unmarshal Ruby values into Go struct data

Mapping values from Ruby into Go struct.
Func `Unmarshal` fetch values for every field in Ruby object that defined by `mruby:"METHOD_NAME"` tag. 

```go
type Planet struct {
    Id         int            `mruby:"id"`
    Name       string         `mruby:"name"`
    Radius     float64        `mruby:"radius"`
    HasMoon    bool           `mruby:"has_moon"`
    Moon       []string       `mruby:"moon"`
    MoonRadius map[string]int `mruby:"moon_radius"`
}
```

Define each value accessor in Ruby class.

```ruby
class Planet
  attr_accessor :id
  attr_accessor :name
  attr_accessor :radius
  attr_accessor :has_moon
  attr_accessor :moon
  attr_accessor :moon_radius
end
```

Retrieve ruby value `*ruby.MrbValue`, then unmarshal it into the Go struct.

```go
rv, _ := mrb.LoadString(`
p = Planet.new
p.id          = 3
p.name        = "Earth"
p.radius      = 6371.0
p.has_moon    = true
p.moon        = ["Moon"]
p.moon_radius = {"Moon" => 1737}
p
`)

p := Planet{}
Unmarshal(rv, &p)

fmt.Printf("Planet: %v\n", p)
// Planet: {3 Earth 6371 true [Moon] map[Moon:1737]}
```

## Decode/encode between Ruby values and Go values

Similar to func `Unmarshal`. Func `DecodeMrbValue` and `EncodeMrbValue` enable mapping values between mruby and Go.
`DecodeMrbValue`/`EncodeMrbValue` have limitation for a type of value.
This supports JSON equivalent types like below.

* nil
* bool
* int
* float
* string
* Array / slice
* Hash / map (Hash/map key must be a string)

```go
mrb := mruby.NewMrb()
defer mrb.Close()

// Mapping to mruby value
rv, err := EncodeMrbValue(mrb, map[string]interface{}{"Earth": 6371, "Moon": 1737})

// Mapping from mruby value
gv, err := DecodeMrbValue(rv)
```

# License

MIT License. See file `LICENSE` for more detail.

