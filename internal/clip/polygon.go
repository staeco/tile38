package clip

import (
	"github.com/tidwall/geojson"
	"github.com/tidwall/geojson/geometry"
)

func clipPolygon(
	polygon *geojson.Polygon, clipper geojson.Object,
) geojson.Object {
	rect := clipper.Rect()
	var newPoints [][]geometry.Point
	base := polygon.Base()
	rings := []geometry.Ring{base.Exterior}
	rings = append(rings, base.Holes...)
	for _, ring := range rings {
		ringPoints := make([]geometry.Point, ring.NumPoints())
		for i := 0; i < len(ringPoints); i++ {
			ringPoints[i] = ring.PointAt(i)
		}
		newPoints = append(newPoints, clipRing(ringPoints, rect))
	}

	var exterior []geometry.Point
	var holes [][]geometry.Point
	if len(newPoints) > 0 {
		exterior = newPoints[0]
	}
	if len(newPoints) > 1 {
		holes = newPoints[1:]
	}
	newPoly := geojson.NewPolygon(
		geometry.NewPoly(exterior, holes, nil),
	)
	if newPoly.Empty() {
		return geojson.NewMultiPolygon(nil)
	}
	return polygon
}
