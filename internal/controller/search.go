package controller

import (
	"bytes"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mmcloughlin/geohash"
	"github.com/tidwall/geojson"
	"github.com/tidwall/geojson/geometry"
	"github.com/tidwall/resp"
	"github.com/tidwall/tile38/internal/bing"
	"github.com/tidwall/tile38/internal/glob"
	"github.com/tidwall/tile38/internal/server"
)

const defaultCircleSteps = 64

type liveFenceSwitches struct {
	searchScanBaseTokens
	obj    geojson.Object
	cmd    string
	roam   roamSwitches
	knn    bool
	groups map[string]string
}

type roamSwitches struct {
	on      bool
	key     string
	id      string
	pattern bool
	meters  float64
	scan    string
	nearbys map[string]map[string]bool
}

type roamMatch struct {
	id     string
	obj    geojson.Object
	meters float64
}

func (s liveFenceSwitches) Error() string {
	return goingLive
}

func (s liveFenceSwitches) Close() {
	for _, whereeval := range s.whereevals {
		whereeval.Close()
	}
}

func (s liveFenceSwitches) usingLua() bool {
	return len(s.whereevals) > 0
}

func (c *Controller) cmdSearchArgs(
	fromFenceCmd bool, cmd string, vs []resp.Value, types []string,
) (s liveFenceSwitches, err error) {
	var t searchScanBaseTokens
	if fromFenceCmd {
		t.fence = true
	}
	vs, t, err = c.parseSearchScanBaseTokens(cmd, t, vs)
	if err != nil {
		return
	}
	s.searchScanBaseTokens = t
	var typ string
	var ok bool
	if vs, typ, ok = tokenval(vs); !ok || typ == "" {
		err = errInvalidNumberOfArguments
		return
	}
	if s.searchScanBaseTokens.output == outputBounds {
		if cmd == "within" || cmd == "intersects" {
			if _, err := strconv.ParseFloat(typ, 64); err == nil {
				// It's likely that the output was not specified, but rather the search bounds.
				s.searchScanBaseTokens.output = defaultSearchOutput
				vs = append([]resp.Value{resp.StringValue(typ)}, vs...)
				typ = "BOUNDS"
			}
		}
	}
	ltyp := strings.ToLower(typ)
	var found bool
	for _, t := range types {
		if ltyp == t {
			found = true
			break
		}
	}
	if !found && s.searchScanBaseTokens.fence && ltyp == "roam" && cmd == "nearby" {
		// allow roaming for nearby fence searches.
		found = true
	}
	if !found {
		err = errInvalidArgument(typ)
		return
	}
	switch ltyp {
	case "point":
		fallthrough
	case "circle":
		if s.clip {
			err = errInvalidArgument("cannnot clip with " + ltyp)
			return
		}
		var slat, slon, smeters string
		if vs, slat, ok = tokenval(vs); !ok || slat == "" {
			err = errInvalidNumberOfArguments
			return
		}
		if vs, slon, ok = tokenval(vs); !ok || slon == "" {
			err = errInvalidNumberOfArguments
			return
		}
		umeters := true
		if vs, smeters, ok = tokenval(vs); !ok || smeters == "" {
			umeters = false
			if cmd == "nearby" {
				// possible that this is KNN search
				s.knn = s.searchScanBaseTokens.ulimit && // must be true
					!s.searchScanBaseTokens.usparse // must be false
			}
			if !s.knn {
				err = errInvalidArgument(slat)
				return
			}
		}
		var lat, lon, meters float64
		if lat, err = strconv.ParseFloat(slat, 64); err != nil {
			err = errInvalidArgument(slat)
			return
		}
		if lon, err = strconv.ParseFloat(slon, 64); err != nil {
			err = errInvalidArgument(slon)
			return
		}
		if umeters {
			if meters, err = strconv.ParseFloat(smeters, 64); err != nil {
				err = errInvalidArgument(smeters)
				return
			}
			if meters < 0 {
				err = errInvalidArgument(smeters)
				return
			}
		}
		if s.knn {
			s.obj = geojson.NewPoint(geometry.Point{X: lon, Y: lat})
		} else {
			s.obj = geojson.NewCircle(geometry.Point{X: lon, Y: lat},
				meters, defaultCircleSteps)
		}
	case "object":
		if s.clip {
			err = errInvalidArgument("cannnot clip with object")
			return
		}
		var obj string
		if vs, obj, ok = tokenval(vs); !ok || obj == "" {
			err = errInvalidNumberOfArguments
			return
		}
		s.obj, err = geojson.Parse(obj, nil)
		if err != nil {
			return
		}
	case "bounds":
		var sminLat, sminLon, smaxlat, smaxlon string
		if vs, sminLat, ok = tokenval(vs); !ok || sminLat == "" {
			err = errInvalidNumberOfArguments
			return
		}
		if vs, sminLon, ok = tokenval(vs); !ok || sminLon == "" {
			err = errInvalidNumberOfArguments
			return
		}
		if vs, smaxlat, ok = tokenval(vs); !ok || smaxlat == "" {
			err = errInvalidNumberOfArguments
			return
		}
		if vs, smaxlon, ok = tokenval(vs); !ok || smaxlon == "" {
			err = errInvalidNumberOfArguments
			return
		}
		var minLat, minLon, maxLat, maxLon float64
		if minLat, err = strconv.ParseFloat(sminLat, 64); err != nil {
			err = errInvalidArgument(sminLat)
			return
		}
		if minLon, err = strconv.ParseFloat(sminLon, 64); err != nil {
			err = errInvalidArgument(sminLon)
			return
		}
		if maxLat, err = strconv.ParseFloat(smaxlat, 64); err != nil {
			err = errInvalidArgument(smaxlat)
			return
		}
		if maxLon, err = strconv.ParseFloat(smaxlon, 64); err != nil {
			err = errInvalidArgument(smaxlon)
			return
		}
		s.obj = geojson.NewRect(geometry.Rect{
			Min: geometry.Point{X: minLon, Y: minLat},
			Max: geometry.Point{X: maxLon, Y: maxLat},
		})
	case "hash":
		var hash string
		if vs, hash, ok = tokenval(vs); !ok || hash == "" {
			err = errInvalidNumberOfArguments
			return
		}
		box := geohash.BoundingBox(hash)
		s.obj = geojson.NewRect(geometry.Rect{
			Min: geometry.Point{X: box.MinLng, Y: box.MinLat},
			Max: geometry.Point{X: box.MaxLng, Y: box.MaxLat},
		})
	case "quadkey":
		var key string
		if vs, key, ok = tokenval(vs); !ok || key == "" {
			err = errInvalidNumberOfArguments
			return
		}
		var minLat, minLon, maxLat, maxLon float64
		minLat, minLon, maxLat, maxLon, err = bing.QuadKeyToBounds(key)
		if err != nil {
			err = errInvalidArgument(key)
			return
		}
		s.obj = geojson.NewRect(geometry.Rect{
			Min: geometry.Point{X: minLon, Y: minLat},
			Max: geometry.Point{X: maxLon, Y: maxLat},
		})
	case "tile":
		var sx, sy, sz string
		if vs, sx, ok = tokenval(vs); !ok || sx == "" {
			err = errInvalidNumberOfArguments
			return
		}
		if vs, sy, ok = tokenval(vs); !ok || sy == "" {
			err = errInvalidNumberOfArguments
			return
		}
		if vs, sz, ok = tokenval(vs); !ok || sz == "" {
			err = errInvalidNumberOfArguments
			return
		}
		var x, y int64
		var z uint64
		if x, err = strconv.ParseInt(sx, 10, 64); err != nil {
			err = errInvalidArgument(sx)
			return
		}
		if y, err = strconv.ParseInt(sy, 10, 64); err != nil {
			err = errInvalidArgument(sy)
			return
		}
		if z, err = strconv.ParseUint(sz, 10, 64); err != nil {
			err = errInvalidArgument(sz)
			return
		}
		var minLat, minLon, maxLat, maxLon float64
		minLat, minLon, maxLat, maxLon = bing.TileXYToBounds(x, y, z)
		s.obj = geojson.NewRect(geometry.Rect{
			Min: geometry.Point{X: minLon, Y: minLat},
			Max: geometry.Point{X: maxLon, Y: maxLat},
		})
	case "get":
		if s.clip {
			err = errInvalidArgument("cannnot clip with get")
		}
		var key, id string
		if vs, key, ok = tokenval(vs); !ok || key == "" {
			err = errInvalidNumberOfArguments
			return
		}
		if vs, id, ok = tokenval(vs); !ok || id == "" {
			err = errInvalidNumberOfArguments
			return
		}
		col := c.getCol(key)
		if col == nil {
			err = errKeyNotFound
			return
		}
		s.obj, _, ok = col.Get(id)
		if !ok {
			err = errIDNotFound
			return
		}
	case "roam":
		s.roam.on = true
		if vs, s.roam.key, ok = tokenval(vs); !ok || s.roam.key == "" {
			err = errInvalidNumberOfArguments
			return
		}
		if vs, s.roam.id, ok = tokenval(vs); !ok || s.roam.id == "" {
			err = errInvalidNumberOfArguments
			return
		}
		s.roam.pattern = glob.IsGlob(s.roam.id)
		var smeters string
		if vs, smeters, ok = tokenval(vs); !ok || smeters == "" {
			err = errInvalidNumberOfArguments
			return
		}
		if s.roam.meters, err = strconv.ParseFloat(smeters, 64); err != nil {
			err = errInvalidArgument(smeters)
			return
		}
		var scan string
		if vs, scan, ok = tokenval(vs); ok {
			if strings.ToLower(scan) != "scan" {
				err = errInvalidArgument(scan)
				return
			}
			if vs, scan, ok = tokenval(vs); !ok || scan == "" {
				err = errInvalidNumberOfArguments
				return
			}
			s.roam.scan = scan
		}
	}
	if len(vs) != 0 {
		err = errInvalidNumberOfArguments
		return
	}
	return
}

var nearbyTypes = []string{"point"}
var withinOrIntersectsTypes = []string{
	"geo", "bounds", "hash", "tile", "quadkey", "get", "object", "circle"}

func (c *Controller) cmdNearby(msg *server.Message) (res resp.Value, err error) {
	start := time.Now()
	vs := msg.Values[1:]
	wr := &bytes.Buffer{}
	s, err := c.cmdSearchArgs(false, "nearby", vs, nearbyTypes)
	if s.usingLua() {
		defer s.Close()
		defer func() {
			if r := recover(); r != nil {
				res = server.NOMessage
				err = errors.New(r.(string))
				return
			}
		}()
	}
	if err != nil {
		return server.NOMessage, err
	}
	s.cmd = "nearby"
	if s.fence {
		return server.NOMessage, s
	}
	sw, err := c.newScanWriter(
		wr, msg, s.key, s.output, s.precision, s.glob, false,
		s.cursor, s.limit, s.wheres, s.whereins, s.whereevals, s.nofields)
	if err != nil {
		return server.NOMessage, err
	}
	if msg.OutputType == server.JSON {
		wr.WriteString(`{"ok":true`)
	}
	sw.writeHead()
	if sw.col != nil {
		var matched uint32
		iter := func(id string, o geojson.Object, fields []float64, dist *float64) bool {
			// Calculate distance if we need to
			distance := 0.0
			if s.distance {
				if dist != nil {
					distance = *dist
				} else {
					distance = o.Distance(s.obj)
				}
			}
			return sw.writeObject(ScanWriterParams{
				id:              id,
				o:               o,
				fields:          fields,
				distance:        distance,
				noLock:          true,
				ignoreGlobMatch: s.knn,
			})
		}
		if s.knn {
			c.nearestNeighbors(&s, sw, s.obj, &matched, iter)
		} else {
			sw.col.Intersects(s.obj, s.sparse, func(
				id string, o geojson.Object, fields []float64,
			) bool {
				if c.hasExpired(s.key, id) {
					return true
				}
				return iter(id, o, fields, nil)
			})
		}
	}
	sw.writeFoot()
	if msg.OutputType == server.JSON {
		wr.WriteString(`,"elapsed":"` + time.Now().Sub(start).String() + "\"}")
		return resp.BytesValue(wr.Bytes()), nil
	}
	return sw.respOut, nil
}

type iterItem struct {
	id     string
	o      geojson.Object
	fields []float64
	dist   float64
}

func (c *Controller) nearestNeighbors(
	s *liveFenceSwitches, sw *scanWriter, target geojson.Object, matched *uint32,
	iter func(id string, o geojson.Object, fields []float64, dist *float64,
	) bool) {
	limit := int(sw.cursor + sw.limit)
	var items []iterItem
	sw.col.Nearby(target, func(id string, o geojson.Object, fields []float64) bool {
		if c.hasExpired(s.key, id) {
			return true
		}
		if _, ok := sw.fieldMatch(fields, o); !ok {
			return true
		}
		match, keepGoing := sw.globMatch(id, o)
		if !match {
			return true
		}
		dist := o.Distance(target)
		items = append(items, iterItem{id: id, o: o, fields: fields, dist: dist})
		if !keepGoing {
			return false
		}
		return len(items) < limit
	})
	sort.Slice(items, func(i, j int) bool {
		return items[i].dist < items[j].dist
	})
	for _, item := range items {
		if !iter(item.id, item.o, item.fields, &item.dist) {
			return
		}
	}
}

func (c *Controller) cmdWithin(msg *server.Message) (res resp.Value, err error) {
	return c.cmdWithinOrIntersects("within", msg)
}

func (c *Controller) cmdIntersects(msg *server.Message) (res resp.Value, err error) {
	return c.cmdWithinOrIntersects("intersects", msg)
}

func (c *Controller) cmdWithinOrIntersects(cmd string, msg *server.Message) (res resp.Value, err error) {
	start := time.Now()
	vs := msg.Values[1:]

	wr := &bytes.Buffer{}
	s, err := c.cmdSearchArgs(false, cmd, vs, withinOrIntersectsTypes)
	if s.usingLua() {
		defer s.Close()
		defer func() {
			if r := recover(); r != nil {
				res = server.NOMessage
				err = errors.New(r.(string))
				return
			}
		}()
	}
	if err != nil {
		return server.NOMessage, err
	}
	s.cmd = cmd
	if s.fence {
		return server.NOMessage, s
	}
	sw, err := c.newScanWriter(
		wr, msg, s.key, s.output, s.precision, s.glob, false,
		s.cursor, s.limit, s.wheres, s.whereins, s.whereevals, s.nofields)
	if err != nil {
		return server.NOMessage, err
	}
	if msg.OutputType == server.JSON {
		wr.WriteString(`{"ok":true`)
	}
	sw.writeHead()
	if sw.col != nil {
		if cmd == "within" {
			sw.col.Within(s.obj, s.sparse, func(
				id string, o geojson.Object, fields []float64,
			) bool {
				if c.hasExpired(s.key, id) {
					return true
				}
				return sw.writeObject(ScanWriterParams{
					id:     id,
					o:      o,
					fields: fields,
					noLock: true,
				})
			})
		} else if cmd == "intersects" {
			sw.col.Intersects(s.obj, s.sparse, func(
				id string,
				o geojson.Object,
				fields []float64,
			) bool {
				if c.hasExpired(s.key, id) {
					return true
				}
				params := ScanWriterParams{
					id:     id,
					o:      o,
					fields: fields,
					noLock: true,
				}
				if s.clip {
					params.clip = s.obj
				}
				return sw.writeObject(params)
			})
		}
	}
	sw.writeFoot()
	if msg.OutputType == server.JSON {
		wr.WriteString(`,"elapsed":"` + time.Now().Sub(start).String() + "\"}")
		return resp.BytesValue(wr.Bytes()), nil
	}
	return sw.respOut, nil
}

func (c *Controller) cmdSeachValuesArgs(vs []resp.Value) (
	s liveFenceSwitches, err error,
) {
	var t searchScanBaseTokens
	vs, t, err = c.parseSearchScanBaseTokens("search", t, vs)
	if err != nil {
		return
	}
	s.searchScanBaseTokens = t
	if len(vs) != 0 {
		err = errInvalidNumberOfArguments
		return
	}
	return
}

func (c *Controller) cmdSearch(msg *server.Message) (res resp.Value, err error) {
	start := time.Now()
	vs := msg.Values[1:]

	wr := &bytes.Buffer{}
	s, err := c.cmdSeachValuesArgs(vs)
	if s.usingLua() {
		defer s.Close()
		defer func() {
			if r := recover(); r != nil {
				res = server.NOMessage
				err = errors.New(r.(string))
				return
			}
		}()
	}
	if err != nil {
		return server.NOMessage, err
	}
	sw, err := c.newScanWriter(
		wr, msg, s.key, s.output, s.precision, s.glob, true,
		s.cursor, s.limit, s.wheres, s.whereins, s.whereevals, s.nofields)
	if err != nil {
		return server.NOMessage, err
	}
	if msg.OutputType == server.JSON {
		wr.WriteString(`{"ok":true`)
	}
	sw.writeHead()
	if sw.col != nil {
		if sw.output == outputCount && len(sw.wheres) == 0 && sw.globEverything == true {
			count := sw.col.Count() - int(s.cursor)
			if count < 0 {
				count = 0
			}
			sw.count = uint64(count)
		} else {
			g := glob.Parse(sw.globPattern, s.desc)
			if g.Limits[0] == "" && g.Limits[1] == "" {
				sw.col.SearchValues(s.desc,
					func(id string, o geojson.Object, fields []float64) bool {
						return sw.writeObject(ScanWriterParams{
							id:     id,
							o:      o,
							fields: fields,
							noLock: true,
						})
					},
				)
			} else {
				// must disable globSingle for string value type matching because
				// globSingle is only for ID matches, not values.
				sw.globSingle = false
				sw.col.SearchValuesRange(g.Limits[0], g.Limits[1], s.desc,
					func(id string, o geojson.Object, fields []float64) bool {
						return sw.writeObject(ScanWriterParams{
							id:     id,
							o:      o,
							fields: fields,
							noLock: true,
						})
					},
				)
			}
		}
	}
	sw.writeFoot()
	if msg.OutputType == server.JSON {
		wr.WriteString(`,"elapsed":"` + time.Now().Sub(start).String() + "\"}")
		return resp.BytesValue(wr.Bytes()), nil
	}
	return sw.respOut, nil
}
