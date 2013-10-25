Package j2n allows arbitrary JSON to be marshaled into structs. Any JSON fields
that are not marshaled directly into the fields of the struct are put into a 
field called `Overflow`, of type `map[string]*json.RawMessage`.

This means that fields that are not explicitly named in the struct will survive 
an Unmarshal/Marshal round trip.

To avoid recursive calls to MarshalJSON/UnmarshalJSON, use the following 
pattern:

```
type CatData struct {
	Name     string                      `json:"name"`
	Overflow map[string]*json.RawMessage `json:"-"`
}

type Cat struct {
	CatData
}

func (c *Cat) UnmarshalJSON(data []byte) error {
	return j2n.UnmarshalJSON(data, &c.CatData)
}

func (c Cat) MarshalJSON() ([]byte, error) {
	return j2n.MarshalJSON(c.CatData)
}
```
