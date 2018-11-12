package clip

import "github.com/tidwall/geojson"

func clipFeature(
	feature *geojson.Feature, clipper geojson.Object,
) geojson.Object {
	newFeature := Clip(feature.Base(), clipper)
	if _, ok := newFeature.(*geojson.Feature); !ok {
		newFeature = geojson.NewFeature(newFeature, feature.Members())
	}
	return newFeature
}
