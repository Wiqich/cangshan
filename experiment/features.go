package experiment

type Features interface {
	UpdateFeature(name string, value interface{}, operation string)
	GetFeature(name string) interface{}
}

type MapFeatures map[string]interface{}

func (features MapFeatures) GetFeature(name string) interface{} {
	return features[name]
}

func (features MapFeatures) UpdateFeature(name string, value interface{}, operation string) {
	switch operation {
	case "append":
		if values, ok := features[name].([]interface{}); ok {
			features[name] = append(values, value)
		}
	default:
		features[name] = value
	}
}
