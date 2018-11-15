package server

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/tidwall/resp"
	"github.com/tidwall/tile38/internal/log"
)

// MASSINSERT num_keys num_points [minx miny maxx maxy]

const useRandField = true

func randMassInsertPosition(minLat, minLon, maxLat, maxLon float64) (float64, float64) {
	lat, lon := (rand.Float64()*(maxLat-minLat))+minLat, (rand.Float64()*(maxLon-minLon))+minLon
	return lat, lon
}

func (c *Server) cmdMassInsert(msg *Message) (res resp.Value, err error) {
	start := time.Now()
	vs := msg.Args[1:]

	minLat, minLon, maxLat, maxLon := -90.0, -180.0, 90.0, 180.0 //37.10776, -122.67145, 38.19502, -121.62775

	var snumCols, snumPoints string
	var cols, objs int
	var ok bool
	if vs, snumCols, ok = tokenval(vs); !ok || snumCols == "" {
		return NOMessage, errInvalidNumberOfArguments
	}
	if vs, snumPoints, ok = tokenval(vs); !ok || snumPoints == "" {
		return NOMessage, errInvalidNumberOfArguments
	}
	if len(vs) != 0 {
		var sminLat, sminLon, smaxLat, smaxLon string
		if vs, sminLat, ok = tokenval(vs); !ok || sminLat == "" {
			return NOMessage, errInvalidNumberOfArguments
		}
		if vs, sminLon, ok = tokenval(vs); !ok || sminLon == "" {
			return NOMessage, errInvalidNumberOfArguments
		}
		if vs, smaxLat, ok = tokenval(vs); !ok || smaxLat == "" {
			return NOMessage, errInvalidNumberOfArguments
		}
		if vs, smaxLon, ok = tokenval(vs); !ok || smaxLon == "" {
			return NOMessage, errInvalidNumberOfArguments
		}
		var err error
		if minLat, err = strconv.ParseFloat(sminLat, 64); err != nil {
			return NOMessage, err
		}
		if minLon, err = strconv.ParseFloat(sminLon, 64); err != nil {
			return NOMessage, err
		}
		if maxLat, err = strconv.ParseFloat(smaxLat, 64); err != nil {
			return NOMessage, err
		}
		if maxLon, err = strconv.ParseFloat(smaxLon, 64); err != nil {
			return NOMessage, err
		}
		if len(vs) != 0 {
			return NOMessage, errors.New("invalid number of arguments")
		}
	}
	n, err := strconv.ParseUint(snumCols, 10, 64)
	if err != nil {
		return NOMessage, errInvalidArgument(snumCols)
	}
	cols = int(n)
	n, err = strconv.ParseUint(snumPoints, 10, 64)
	if err != nil {
		return NOMessage, errInvalidArgument(snumPoints)
	}
	docmd := func(args []string) error {
		nmsg := &Message{}
		*nmsg = *msg
		nmsg.Args = args
		var d commandDetailsT
		_, d, err = c.command(nmsg, nil)
		if err != nil {
			return err
		}

		return c.writeAOF(nmsg.Args, &d)

	}
	rand.Seed(time.Now().UnixNano())
	objs = int(n)
	var k uint64
	for i := 0; i < cols; i++ {
		key := "mi:" + strconv.FormatInt(int64(i), 10)
		func(key string) {
			// lock cycle
			for j := 0; j < objs; j++ {
				id := strconv.FormatInt(int64(j), 10)
				var values []string
				if j%8 == 0 {
					values = append(values, "set", key, id, "STRING", fmt.Sprintf("str%v", j))
				} else {
					lat, lon := randMassInsertPosition(minLat, minLon, maxLat, maxLon)
					values = make([]string, 0, 16)
					values = append(values, "set", key, id)
					if useRandField {
						values = append(values, "FIELD", "fname",
							strconv.FormatFloat(rand.Float64()*10, 'f', -1, 64))
					}
					values = append(values, "POINT",
						strconv.FormatFloat(lat, 'f', -1, 64),
						strconv.FormatFloat(lon, 'f', -1, 64),
					)
				}
				if err := docmd(values); err != nil {
					log.Fatal(err)
					return
				}
				atomic.AddUint64(&k, 1)
				if j%1000 == 1000-1 {
					log.Infof("massinsert: %s %d/%d", key, atomic.LoadUint64(&k), cols*objs)
				}
			}
		}(key)
	}
	log.Infof("massinsert: done %d objects", atomic.LoadUint64(&k))
	return OKMessage(msg, start), nil
}

func (c *Server) cmdSleep(msg *Message) (res resp.Value, err error) {
	start := time.Now()
	if len(msg.Args) != 2 {
		return NOMessage, errInvalidNumberOfArguments
	}
	d, _ := strconv.ParseFloat(msg.Args[1], 64)
	time.Sleep(time.Duration(float64(time.Second) * d))
	return OKMessage(msg, start), nil
}
