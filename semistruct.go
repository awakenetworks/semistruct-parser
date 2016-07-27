package semistruct

import "strconv"
import p "github.com/andyleap/parser"

type semistruct_log struct {
	priority int64
	tags     []string
	attrs    map[string]string
}

type kv_pair struct {
	key string
	val string
}

// TODO: Make sure all fields in the schema that are optional are
// handled in here with good mempty-like defaults (or is there an
// option type analog in Go?).
//
// TODO: Rename some of these parser combinators.
//
// TODO: Refactor the repetitive Mult calls all over the place if I
// can, there's a lot of repetition I'd like to get rid of on my first
// pass.
//
// TODO: handle failure better - if the string can't parse, what do we
// do?
//
// TODO: are we consuming to the end of the line???

// Parse the log level priority indicator; this can only be an integer
// from 0 to 9
func priorityInt() *p.Grammar {
	o := p.Set("0-9")
	o.Node(func(m p.Match) (p.Match, error) {
		v, err := strconv.ParseInt(p.String(m), 10, 64)
		if err != nil {
			return nil, err
		}
		return v, nil
	})

	return o
}

// Parse zero or more characters that are alpha-numeric (note, \w
// includes the underscore).
func alphaNum() *p.Grammar {
	return p.Mult(
		0, 0, p.Set("\\w\\-"),
	)
}

// Parse zero or more characters that are alpha-numeric, space
// characters, and special characters (note, \w includes the
// underscore).
func alphaSpecial() *p.Grammar {
	return p.Mult(
		0, 0,
		p.Set("\\w\\s\\-~`!@#$%^&*()+=?/><,.{}[]|'\\"),
	)
}

// Parse zero or more characters that are alpha and underscore.
func alpha() *p.Grammar {
	return p.Mult(
		0, 0, p.Set("a-zA-Z_"),
	)
}

// Parse a alphaSpecial string wrapped in double quotes.
func quotedStr() *p.Grammar {
	return p.And(
		p.Ignore(p.Lit("\"")),
		alphaSpecial(),
		p.Ignore(p.Lit("\"")),
	)
}

// Consume the opening sentinel "!<".
func openSentinel() *p.Grammar {
	return p.Ignore(
		p.And(p.Lit("!"), p.Lit("<")),
	)
}

// Consume the ending sentinel ">!".
func endSentinel() *p.Grammar {
	return p.Ignore(
		p.And(p.Lit(">"), p.Lit("!")),
	)
}

// Consume whitespace.
func skipSpace() *p.Grammar {
	return p.Ignore(
		p.Mult(0, 0, p.Set("\\s")),
	)
}

// Parse zero or more tag atoms separated by a colon.
func tag() *p.Grammar {
	o := p.Optional(
		p.And(
			p.Tag("tag", alphaNum()),
			p.Mult(
				0, 0, p.And(
					p.Ignore(p.Lit(":")),
					p.Tag("tag", alphaNum()),
				),
			),
		),
	)

	o.Node(func(m p.Match) (p.Match, error) {
		var tagAccum []string
		for _, op := range p.GetTags(m, "tag") {
			tagAccum = append(tagAccum, p.String(op))
		}
		return tagAccum, nil
	})
	return o
}

func tags() *p.Grammar {
	o := p.Optional(
		p.And(
			p.Ignore(p.Lit("[")),
			p.Tag("tag-elements", tag()),
			p.Ignore(p.Lit("]")),
		),
	)

	o.Node(func(m p.Match) (p.Match, error) {
		elems := p.GetTag(m, "tag-elements")
		if elems == nil {
			elems = make([]string, 0)
		}
		return elems, nil
	})
	return o
}

// NOTE: order within the Or combinator is important here, we want
// to try matching the alphaNum value first before then attempting
// the alphaNum value
func kvpair() *p.Grammar {
	o := p.And(
		p.Tag("key", alphaNum()),
		p.Lit("="),
		p.Tag(
			"value",
			p.Or(quotedStr(), alphaNum()),
		),
	)

	o.Node(func(m p.Match) (p.Match, error) {
		k := p.String(p.GetTag(m, "key"))
		v := p.String(p.GetTag(m, "value"))
		return kv_pair{k, v}, nil
	})
	return o
}

func kvpairs() *p.Grammar {
	o := p.Optional(
		p.And(
			p.Tag("pair", kvpair()),
			p.Mult(0, 0,
				p.And(
					p.Ignore(p.Set("\\s")),
					p.Tag("pair", kvpair()),
				),
			),
		),
	)

	o.Node(func(m p.Match) (p.Match, error) {
		attrmap := make(map[string]string)
		for _, v := range p.GetTags(m, "pair") {
			kvp := v.(kv_pair)
			attrmap[kvp.key] = kvp.val
		}
		return attrmap, nil
	})
	return o
}

func attrs() *p.Grammar {
	o := p.Optional(
		p.And(
			p.Ignore(p.Lit("{")),
			skipSpace(),
			p.Tag("attrmap", kvpairs()),
			skipSpace(),
			p.Ignore(p.Lit("}")),
		),
	)

	o.Node(func(m p.Match) (p.Match, error) {
		attrmap := p.GetTag(m, "attrmap")
		if attrmap == nil {
			attrmap = make(map[string]string)
		}

		return attrmap, nil
	})

	return o
}

// Top-level semi-structured log line parser combinator
func semistruct_parser() *p.Grammar {
	o := p.And(
		// Parse opening sentinel "!<"
		openSentinel(),
		skipSpace(),

		// Parse our log-level priority
		p.Tag("priority", priorityInt()),
		skipSpace(),

		// Parse our tag list "[tag1:tag2:tag3]"
		p.Tag("tags", tags()),
		skipSpace(),

		// Parse our attribute set "{ key=val key2=val2 }"
		p.Tag("attrs", attrs()),
		skipSpace(),

		// Parse the ending sentinel ">!"
		endSentinel(),
	)

	o.Node(func(m p.Match) (p.Match, error) {
		pr := p.GetTag(m, "priority").(int64)
		tg := p.GetTag(m, "tags").([]string)
		at := p.GetTag(m, "attrs").(map[string]string)

		return semistruct_log{pr, tg, at}, nil
	})

	return o
}
