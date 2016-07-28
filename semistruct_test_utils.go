package semistruct

import (
	"fmt"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"reflect"
	"strings"
	"unicode"
)

func mkAttrStr(m map[string]string, quoted bool) string {
	var acc []string
	for k, v := range m {
		var st string
		if quoted {
			st = fmt.Sprintf("%s=\"%s\"", k, v)
		} else {
			st = fmt.Sprintf("%s=%s", k, v)
		}
		acc = append(acc, st)
	}
	return fmt.Sprint("{ ", strings.Join(acc[:], " "), " }")
}

func mkTagStr(t []string) string {
	tags := strings.Join(t[:], ":")
	return fmt.Sprint("[", tags, "]")
}

func mkLogLine(priority int16, tags string, attrs string) string {
	return fmt.Sprintf("!< %d %s %s >!", priority, tags, attrs)
}

// TODO: their primary example of combining the rune generators is
// with the Frequency function but I don't actually want to use that
// in order to compose all of these rune generators. CombineGens
// doesn't like these RuneRanges so I think I have to put each one
// into a separate "gen"...
func SymbolChar() gopter.Gen {
	return gen.Frequency(map[int]gopter.Gen{
		0: gen.RuneRange('#', '/'),
		1: gen.RuneRange(':', '?'),
		2: gen.RuneRange('{', '~'),
		3: gen.RuneRange('[', '_'),
		4: gen.RuneRange('@', '@'),
		5: gen.RuneRange('`', '`'),
	})
}

func AlphaNumSymbol() gopter.Gen {
	return gopter.CombineGens(
		gen.SliceOf(gen.AlphaNumChar()),
		gen.SliceOf(SymbolChar()),
	).Map(func(values []interface{}) string {
		alpha := values[0].([]rune)
		symb := values[1].([]rune)
		return string(append(alpha, symb...))
	}).SuchThat(func(str string) bool {
		for _, ch := range str {
			if !unicode.IsLetter(ch) &&
				!unicode.IsDigit(ch) &&
				!unicode.Is(unicode.Punct, ch) &&
				!unicode.Is(unicode.Symbol, ch) &&
				!unicode.Is(unicode.Mark, ch) &&
				!unicode.Is(unicode.Other, ch) &&
				!unicode.Is(unicode.Space, ch) {

				return false
			}
		}
		return true
	}).WithShrinker(gen.StringShrinker)
}

// Because this isn't exported by gopter...grrr
func genString(runeGen gopter.Gen, runeSieve func(ch rune) bool) gopter.Gen {
	return gen.SliceOf(runeGen).Map(runesToString).SuchThat(func(v string) bool {
		for _, ch := range v {
			if !runeSieve(ch) {
				return false
			}
		}
		return true
	}).WithShrinker(gen.StringShrinker)
}

// Because this isn't exported by gopter...grrr
func runesToString(v []rune) string {
	return string(v)
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
