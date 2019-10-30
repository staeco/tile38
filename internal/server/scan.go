package server

import (
	"bytes"
	"errors"
	"time"

	"github.com/tidwall/geojson"
	"github.com/tidwall/resp"
	"github.com/tidwall/tile38/internal/glob"
)

func (s *Server) cmdScanArgs(vs []string) (
	ls liveFenceSwitches, err error,
) {
	var t searchScanBaseTokens
	vs, t, err = s.parseSearchScanBaseTokens("scan", t, vs)
	if err != nil {
		return
	}
	ls.searchScanBaseTokens = t
	if len(vs) != 0 {
		err = errInvalidNumberOfArguments
		return
	}
	return
}

func (s *Server) cmdScan(msg *Message) (res resp.Value, err error) {
	start := time.Now()
	vs := msg.Args[1:]

	args, err := s.cmdScanArgs(vs)
	if args.usingLua() {
		defer args.Close()
		defer func() {
			if r := recover(); r != nil {
				res = NOMessage
				err = errors.New(r.(string))
				return
			}
		}()
	}
	if err != nil {
		return NOMessage, err
	}
	wr := &bytes.Buffer{}
	sw, err := s.newScanWriter(
		wr, msg, args.key, args.output, args.precision, args.glob, false,
		args.cursor, args.limit, args.wheres, args.whereins, args.whereevals,
		args.nofields)
	if err != nil {
		return NOMessage, err
	}
	if msg.OutputType == JSON {
		wr.WriteString(`{"ok":true`)
	}
	sw.writeHead()
	if sw.col != nil {
		if sw.output == outputCount && len(sw.wheres) == 0 &&
			len(sw.whereins) == 0 && sw.globEverything == true {
			count := sw.col.Count() - int(args.cursor)
			if count < 0 {
				count = 0
			}
			sw.count = uint64(count)
		} else {
			g := glob.Parse(sw.globPattern, args.desc)
			if g.Limits[0] == "" && g.Limits[1] == "" {
				sw.col.Scan(args.desc, sw,
					msg.Deadline,
					func(id string, o geojson.Object, fields []float64) bool {
						return sw.writeObject(ScanWriterParams{
							id:     id,
							o:      o,
							fields: fields,
						})
					},
				)
			} else {
				sw.col.ScanRange(g.Limits[0], g.Limits[1], args.desc, sw,
					msg.Deadline,
					func(id string, o geojson.Object, fields []float64) bool {
						return sw.writeObject(ScanWriterParams{
							id:     id,
							o:      o,
							fields: fields,
						})
					},
				)
			}
		}
	}
	sw.writeFoot()
	if msg.OutputType == JSON {
		wr.WriteString(`,"elapsed":"` + time.Now().Sub(start).String() + "\"}")
		return resp.BytesValue(wr.Bytes()), nil
	}
	return sw.respOut, nil
}
