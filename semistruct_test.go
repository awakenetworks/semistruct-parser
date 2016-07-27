package semistruct

import (
	"fmt"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"reflect"
	"strings"
	"testing"
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
	{`!< 0 [cl2] { unstruct_msg="some random debugging spam" } >!`},
	{`!< 6 [cl2:filestore:notify] { file="cas%20AV_!AA%20nvvpa.jpg" } >!`},
	{`!< 0 [cl2:filestore:notify] { file2="some blah.jpg" } >!`},
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

func TestParserProperties(t *testing.T) {
	p := semistruct_parser()

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 1000

	properties := gopter.NewProperties(parameters)

	properties.Property("log level priority indicator", prop.ForAll(
		func(priority int16) bool {
			l := mkLogLine(
				priority,
				[]string{"tag1", "tag2"},
				map[string]string{
					"proto": "icmp",
					"uuid":  "3e0a5ae3-7c2f-454e-828c-d838a18d5d8e",
				},
			)
			res, _ := p.ParseString(l)
			return res != nil
		},
		gen.Int16Range(0, 9).WithLabel("priority"),
	))

	// TODO: provide a shrinker
	properties.Property("generate arbitrary tags", prop.ForAll(
		func(tags []string) bool {
			l := mkLogLine(
				4,
				tags,
				map[string]string{
					"proto": "icmp",
					"uuid":  "3e0a5ae3-7c2f-454e-828c-d838a18d5d8e",
				},
			)

			res, _ := p.ParseString(l)
			return res != nil
		},
		gen.SliceOf(gen.Identifier()),
	))

	// TODO: provide a shrinker
	properties.Property("generate arbitrary attributes", prop.ForAll(
		func(attrs map[string]string) bool {
			l := mkLogLine(
				4,
				[]string{"tag1", "tag2"},
				attrs,
			)

			res, _ := p.ParseString(l)
			return res != nil
		},
		MapOfN(4, gen.Identifier(), gen.AlphaString()),
	))

	properties.TestingRun(t)
}

func mkAttrStr(m map[string]string) string {
	var acc []string
	for k, v := range m {
		acc = append(acc, fmt.Sprintf("%s=\"%s\"", k, v))
	}
	return fmt.Sprint("{", strings.Join(acc[:], " "), "}")
}

func mkTagStr(t []string) string {
	tags := strings.Join(t[:], ":")
	return fmt.Sprint("[", tags, "]")
}

func mkLogLine(priority int16, tags []string, attrs map[string]string) string {
	atr := mkAttrStr(attrs)
	return fmt.Sprintf("!< %d %s %s >!", priority, mkTagStr(tags), atr)
}

func MapOf(keyGen gopter.Gen, valGen gopter.Gen) gopter.Gen {
	return func(genParams *gopter.GenParameters) *gopter.GenResult {
		len := 0
		if genParams.Size > 0 {
			len = genParams.Rng.Intn(genParams.Size)
		}

		//result, keySieve, valSieve, keyShrinker, valShrinker := genMap(keyGen, valGen, genParams, len)
		result, _, _, _, _ := genMap(keyGen, valGen, genParams, len)

		genResult := gopter.NewGenResult(result.Interface(), gopter.NoShrinker)
		// if elementSieve != nil {
		// 	genResult.Sieve = forAllSieve(elementSieve)
		// }
		return genResult
	}
}

func MapOfN(len int, keyGen gopter.Gen, valGen gopter.Gen) gopter.Gen {
	return func(genParams *gopter.GenParameters) *gopter.GenResult {
		//result, keySieve, valSieve, keyShrinker, valShrinker := genMap(keyGen, valGen, genParams, len)
		result, _, _, _, _ := genMap(keyGen, valGen, genParams, len)

		genResult := gopter.NewGenResult(result.Interface(), gopter.NoShrinker)
		// if elementSieve != nil {
		// 	genResult.Sieve = forAllSieve(elementSieve)
		// }
		return genResult
	}
}

// Generate a Map of the type specified by the keyGen and ValueGen
// functions. Reflection is used heavily here in order to dynamically
// create a map based on the types of the reflected key and value
// generator types.
//
// This was ripped off from gopter's stock genSlice function which
// looks very similar.
func genMap(keyGen gopter.Gen, valGen gopter.Gen, genParams *gopter.GenParameters, len int) (reflect.Value, func(interface{}) bool, func(interface{}) bool, gopter.Shrinker, gopter.Shrinker) {
	key := keyGen(genParams)
	val := valGen(genParams)

	keySieve := key.Sieve
	valSieve := val.Sieve

	keyShrinker := key.Shrinker
	valShrinker := val.Shrinker

	result := reflect.MakeMap(reflect.MapOf(key.ResultType, val.ResultType))

	for i := 0; i < len; i++ {
		keyV, kok := key.Retrieve()
		valV, vok := val.Retrieve()

		if kok && vok {
			result.SetMapIndex(reflect.ValueOf(keyV), reflect.ValueOf(valV))
		}

		key = keyGen(genParams)
		val = valGen(genParams)
	}

	return result, keySieve, valSieve, keyShrinker, valShrinker
}
