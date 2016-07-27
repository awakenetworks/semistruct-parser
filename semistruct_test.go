package semistruct

import (
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"testing"
	"fmt"
)

type testpair struct {
	logline string
}

// TODO: add property tests for the individual combinators for the
// combinator as a whole.

// Set of hand-written lines
var tests = []testpair{
	{`!< 2 [cl7323:featstore:sess_fun] { one=two dos="wah=hh-77" } >!`},
	{`!< 7 [local:pktproc:parsePkt] { proto=icmp error="unassigned_type" action=dropped } >!`},
	{`!< 4 [cl610:featstore:sess_fun] { unknown_sess=3e0a5ae3-7c2f-454e-828c-d838a18d5d8e } >!`},
	{`!< 7 [cl2] { unstruct_msg="some random debugging spam" } >!`},
	{`!< 6 [cl2:filestore:notify] { file="cas%20AV_!AA%20nvvpa.jpg" } >!`},
}

func TestParser(t *testing.T) {
	p := semistruct_parser()

	for _, pair := range tests {
		res, _ := p.ParseString(pair.logline)
		if res == nil {
			t.Error(
				"Parser failed miseraby on this log line: ", pair.logline,
			)
		}
	}
}

func makeLogLine(p int, tags []string, map[string]string) string {
	
	return fmt.Sprintf("!< %d >!")
}

func PropertyTestParser(t *testing.T) {

	parameters := gopter.DefaultTestParameters()

	//parameters.Rng.Seed(1234) // Just for this example to generate reproducable results
	parameters.MinSuccessfulTests = 1000

	properties := gopter.NewProperties(parameters)

	properties.Property("log level priority indicator", prop.ForAll(
		func(p int) string {
			result := spookyCalculation(a, b)
			if result < 0 {
				return "negative result"
			}
			if result%2 == 0 {
				return "even result"
			}
			return ""
		},
		gen.Int().WithLabel("a"),
		gen.Int().WithLabel("b"),
	))

	properties.TestingRun(t)
}
