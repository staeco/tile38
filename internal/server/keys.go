package server

import (
	"bytes"
	"strings"
	"time"

	"github.com/tidwall/resp"
	"github.com/tidwall/tile38/internal/glob"
)

func (c *Server) cmdKeys(msg *Message) (res resp.Value, err error) {
	var start = time.Now()
	vs := msg.Args[1:]

	var pattern string
	var ok bool
	if vs, pattern, ok = tokenval(vs); !ok || pattern == "" {
		return NOMessage, errInvalidNumberOfArguments
	}
	if len(vs) != 0 {
		return NOMessage, errInvalidNumberOfArguments
	}

	var wr = &bytes.Buffer{}
	var once bool
	if msg.OutputType == JSON {
		wr.WriteString(`{"ok":true,"keys":[`)
	}
	var everything bool
	var greater bool
	var greaterPivot string
	var vals []resp.Value

	iterator := func(key string, value interface{}) bool {
		var match bool
		if everything {
			match = true
		} else if greater {
			if !strings.HasPrefix(key, greaterPivot) {
				return false
			}
			match = true
		} else {
			match, _ = glob.Match(pattern, key)
		}
		if match {
			if once {
				if msg.OutputType == JSON {
					wr.WriteByte(',')
				}
			} else {
				once = true
			}
			switch msg.OutputType {
			case JSON:
				wr.WriteString(jsonString(key))
			case RESP:
				vals = append(vals, resp.StringValue(key))
			}
		}
		return true
	}
	if pattern == "*" {
		everything = true
		c.cols.Scan(iterator)
	} else {
		if strings.HasSuffix(pattern, "*") {
			greaterPivot = pattern[:len(pattern)-1]
			if glob.IsGlob(greaterPivot) {
				greater = false
				c.cols.Scan(iterator)
			} else {
				greater = true
				c.cols.Ascend(greaterPivot, iterator)
			}
		} else if glob.IsGlob(pattern) {
			greater = false
			c.cols.Scan(iterator)
		} else {
			greater = true
			greaterPivot = pattern
			c.cols.Ascend(greaterPivot, iterator)
		}
	}
	if msg.OutputType == JSON {
		wr.WriteString(`],"elapsed":"` + time.Now().Sub(start).String() + "\"}")
		return resp.StringValue(wr.String()), nil
	}
	return resp.ArrayValue(vals), nil
}
