package router

import (
	"fmt"
)

type node struct {
	pattern   string
	children  []*node
	subroutes []string
	endpoint  Handler
}

func (n *node) addChild(pattern string, endpoint Handler) *node {
	child := node{
		pattern:  pattern,
		endpoint: endpoint,
	}
	n.children = append(n.children, &child)
	n.subroutes = append(n.subroutes, pattern)
	return &child
}

// Returns handler for first found matching route
func (n *node) route(ctx *Context, data string) Handler {
	if ctx == nil {
		ctx = NewRouteContext()
	}
	for i, p := range n.subroutes {
		if params := PrefixParse(p, data); params != nil {
			for i := 0; i < len(params.Names); i++ {
				ctx.AddParam(params.Names[i], params.Values[i])
			}
			ctx.ParentPath = ctx.RoutePath
			ctx.RoutePath = ctx.RoutePath + params.Match
			if n.endpoint != nil {
				ctx.Router = n.endpoint.(Router)
			}
			if len(n.children[i].subroutes) == 0 {
				return n.children[i].endpoint
			}
			return n.children[i].route(ctx, data[len(params.Match):])
		}
	}
	return nil
}

//func walk(n *node, parentPattern string) {
//	for _, c := range n.children {
//		pattern := fmt.Sprintf("%s%s", n.pattern, c.pattern)
//		if c.isLeaf() {
//			fmt.Println(pattern)
//		}
//		walk(c, pattern)
//	}
//}

//func (n *node) isLeaf() bool {
//	return len(n.subroutes) == 0
//}

type Param struct {
	Match  string
	Names  []string
	Values []string
}

type parserState uint8

const (
	psAcceptingRawText   parserState = iota << 2
	psAcceptingSpecifier             // accepting {name}
	psAcceptingValue
	psCompleted
)

func shift(runes []rune, buf []rune, r rune) ([]rune, []rune, rune) {
	if len(runes) > 0 {
		return runes[1:], append(buf, r), runes[0]
	}
	return runes, append(buf, r), 0
}

// /article/2022-12-15/
// /article/{year}-{month}-{day}/
func PrefixParse(pattern, input string) *Param {
	if len(pattern) == 0 {
		return &Param{
			Match: input,
		}
	}
	if len(input) == 0 {
		return nil
	}
	p := []rune(pattern)
	p, bufP, p0 := shift(p[1:], nil, p[0])
	i := []rune(input)
	i, bufI, i0 := shift(i[1:], nil, i[0])

	res := Param{
		Match: string(bufI),
	}
	ps := psAcceptingRawText

	for {

		if (i0 == 0 && p0 == 0) || ps == psCompleted {
			break
		}

		switch {
		// reached pattern end
		case ps == psAcceptingRawText && p0 == 0:
			ps = psCompleted
		// read pattern specifier
		case ps == psAcceptingRawText && p0 == '{':
			bufP = nil
			p, _, p0 = shift(p, bufP, p0) // skip {
			for p0 != '}' && p0 != 0 {
				p, bufP, p0 = shift(p, bufP, p0)
			}
			p, _, p0 = shift(p, bufP, p0) // skip }
			res.Names = append(res.Names, string(bufP))
			ps = psAcceptingValue
		// read input value for specifier
		case ps == psAcceptingValue:
			bufI = nil
			for i0 != p0 && i0 != 0 {
				i, bufI, i0 = shift(i, bufI, i0)
			}
			res.Match = fmt.Sprintf("%s%s", res.Match, string(bufI))
			res.Values = append(res.Values, string(bufI))
			ps = psAcceptingRawText
		default:
			if i0 != p0 {
				return nil
			}
			if i0 != 0 {
				res.Match = fmt.Sprintf("%s%s", res.Match, string(i0))
				i, bufI, i0 = shift(i, bufI, i0)
			}
			if p0 != 0 {
				p, bufP, p0 = shift(p, bufP, p0)
			}
		}

	}
	if len(res.Names) > 0 && len(res.Names) == len(res.Values) {
		return &res
	}
	if len(res.Names) == 0 && string(bufP) == string(bufI) {
		return &res
	}
	return nil
}
