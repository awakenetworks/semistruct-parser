package semistruct

import (
	"fmt"
	"github.com/leanovate/gopter"
	"reflect"
	"strings"
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
	return fmt.Sprint("{", strings.Join(acc[:], " "), "}")
}

func mkTagStr(t []string) string {
	tags := strings.Join(t[:], ":")
	return fmt.Sprint("[", tags, "]")
}

func mkLogLine(priority int16, tags string, attrs string) string {
	return fmt.Sprintf("!< %d %s %s >!", priority, tags, attrs)
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
