package semistruct

import (
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"testing"
)

type testPair struct {
	logline string
}

// Hand-written unit tests for common cases this human can think of.
var tests = []testPair{
	{`!< 2 [cl7323:featstore:sess_fun] { ONE=two DOS="wah=hh-77" } >!`},
	{`!< 7 [local:pktproc:parsePkt] { PROTO=icmp ERROR="unassigned_type" ACTION=dropped } >!`},
	{`!< 4 [cl610:featstore:sess_fun] { UNKNOWN_SESS=3e0a5ae3-7c2f-454e-828c-d838a18d5d8e } >!`},
	{`!< 0 [cl2] { UNSTRUCT_MSG="some random debugging spam" } >!`},
	{`!< 6 [cl2:filestore:notify] { FILE="cas%20AV_!AA%20nvvpa.jpg" } >!`},
	{`!< 0 [cl2:filestore:notify] { FILE2="some blah.jpg" } >!`},
	{`!< 2 [cl2:filestore:notify] { FILE2="some blah.jpg" } >!`},
	{`!< 2 [cl2:filestore:notify] { FILE2="some bl\ah.jpg" } >!`},
	{`!< 2 [cl2:filestore:notify] { FILE2="some bl|ah.jpg" } >! ** (process:4707): WARNING **: send_infos_cb: No such interface 'org.gtk.vfs.Enumerator' on object at path /org/gtk/vfs/client/enumerator/385 (g-dbus-error-quark, 19)`},
	{`!< 2 { FILE2="some blah.jpg" } >!`},
	{`!< 2 [cl2:filestore:notify] >!`},
	{`!< 2 >!`},
	{"!< 3 [blah] { FLFANHB2X6UBMERR=\"kL]_:;\" } >!"},
}

// Iterate over the hand-written tests and attempt to parse each line.
func TestParser(t *testing.T) {
	p := NewParseSemistruct()

	for _, pair := range tests {
		res, err := p.ParseString(pair.logline)
		if err != nil || res == nil {
			t.Error(
				"Parser failed miseraby on this log line: ", pair.logline,
			)
		}
	}
}

// Property tests for a whole semistructured log line and for each
// field.
func TestParserProperties(t *testing.T) {
	p := NewParseSemistruct()
	parameters := gopter.DefaultTestParameters()

	// Instantiate a configuration for the *whole* semistructured log
	// line parser.
	parameters.MinSuccessfulTests = 100
	wholeProperty := gopter.NewProperties(parameters)

	// Test the whole log line parsing behavior, each property is
	// generated and composed together then a parse is attempted.
	wholeProperty.Property("arbitrary log line", prop.ForAll(
		func(priority int16, tags []string, attrs map[string]string) bool {
			l := mkLogLine(
				priority,
				mkTagStr(tags),
				mkAttrStr(attrs, true),
			)
			res, err := p.ParseString(l)
			return err == nil && res != nil
		},
		gen.Int16Range(0, 9).WithLabel("priority"),
		gen.SliceOf(gen.Identifier()).WithLabel("tags"),
		MapOf(UpperIdentifier(), gen.AlphaString()).WithLabel("attributes"),
	))

	// Instantiate a configuration for individual property tests of
	// each field.
	parameters.MinSuccessfulTests = 5000
	fieldProperties := gopter.NewProperties(parameters)

	// Test parsing a log line arbitrarily generating just a single
	// property of each "piece" of a log line (the priority, tags, or
	// attributes).
	fieldProperties.Property("arbitrary log level priority indicator", prop.ForAll(
		func(priority int16) bool {
			l := mkLogLine(
				priority,
				mkTagStr([]string{"tag1", "tag2"}),
				mkAttrStr(map[string]string{
					"PROTO": "icmp",
					"UUID":  "3e0a5ae3-7c2f-454e-828c-d838a18d5d8e",
				}, true),
			)
			res, err := p.ParseString(l)
			return err == nil && res != nil
		},
		gen.Int16Range(0, 9).WithLabel("priority"),
	))

	fieldProperties.Property("arbitrary tags", prop.ForAll(
		func(tags []string) bool {
			l := mkLogLine(
				4,
				mkTagStr(tags),
				mkAttrStr(map[string]string{
					"PROTO": "icmp",
					"UUID":  "3e0a5ae3-7c2f-454e-828c-d838a18d5d8e",
				}, true),
			)

			res, err := p.ParseString(l)
			return err == nil && res != nil
		},
		gen.SliceOf(gen.Identifier()).WithLabel("tags"),
	))

	// TODO: provide a shrinker
	fieldProperties.Property("arbitrary attributes with quoted values", prop.ForAll(
		func(attrs map[string]string) bool {
			l := mkLogLine(
				4,
				mkTagStr([]string{"tag1", "tag2"}),
				mkAttrStr(attrs, true),
			)

			res, err := p.ParseString(l)
			return err == nil && res != nil
		},
		MapOf(UpperIdentifier(), AlphaNumSpecialString()).WithLabel("attributes"),
	))

	// TODO: provide a shrinker
	fieldProperties.Property("arbitrary attributes with unquoted values", prop.ForAll(
		func(attrs map[string]string) bool {
			l := mkLogLine(
				4,
				mkTagStr([]string{"tag1", "tag2"}),
				mkAttrStr(attrs, false),
			)

			res, err := p.ParseString(l)
			return err == nil && res != nil
		},
		MapOf(UpperIdentifier(), gen.Identifier()).WithLabel("attributes"),
	))

	// Run the configured property tests!!
	wholeProperty.TestingRun(t)
	fieldProperties.TestingRun(t)
}
