package promhttp

import (
	"sort"
	"strconv"
	"strings"
)

type Accept struct {
	Type, SubType string
	Q             float64
	Params        map[string]string
}

// For internal use, so that we can use the sort interface
type accept_slice []Accept

func (accept accept_slice) Len() int {
	slice := []Accept(accept)
	return len(slice)
}

func (accept accept_slice) Less(i, j int) bool {
	slice := []Accept(accept)
	ai, aj := slice[i], slice[j]
	if ai.Q > aj.Q {
		return true
	}
	if ai.Type != "*" && aj.Type == "*" {
		return true
	}
	if ai.SubType != "*" && aj.SubType == "*" {
		return true
	}
	return false
}

func (accept accept_slice) Swap(i, j int) {
	slice := []Accept(accept)
	slice[i], slice[j] = slice[j], slice[i]
}

// Parse an Accept Header string returning a sorted list
// of clauses
func ParseAccept(header string) (accept []Accept) {
	parts := strings.Split(header, ",")
	accept = make([]Accept, 0, len(parts))
	for _, part := range parts {
		part := strings.Trim(part, " ")

		a := Accept{}
		a.Params = make(map[string]string)
		a.Q = 1.0

		mrp := strings.Split(part, ";")

		media_range := mrp[0]
		sp := strings.Split(media_range, "/")
		a.Type = strings.Trim(sp[0], " ")

		switch {
		case len(sp) == 1 && a.Type == "*":
			a.SubType = "*"
		case len(sp) == 2:
			a.SubType = strings.Trim(sp[1], " ")
		default:
			continue
		}

		if len(mrp) == 1 {
			accept = append(accept, a)
			continue
		}

		for _, param := range mrp[1:] {
			sp := strings.SplitN(param, "=", 2)
			if len(sp) != 2 {
				continue
			}
			token := strings.Trim(sp[0], " ")
			if token == "q" {
				a.Q, _ = strconv.ParseFloat(sp[1], 32)
			} else {
				a.Params[token] = strings.Trim(sp[1], " ")
			}
		}

		accept = append(accept, a)
	}

	slice := accept_slice(accept)
	sort.Sort(slice)

	return
}
