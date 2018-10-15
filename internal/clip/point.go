package clip

import "github.com/tidwall/geojson"

func clipPoint(point *geojson.Point, clipper geojson.Object) geojson.Object {
	if point.IntersectsRect(clipper.Rect()) {
		return point
	}
	return geojson.NewMultiPoint(nil)
}
