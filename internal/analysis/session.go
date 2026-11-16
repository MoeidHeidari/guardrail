package analysis

import (
	"errors"

	"github.com/dop251/goja"
)

type jsSession struct {
	validate    goja.Callable
	api         goja.Value
	buildChange goja.Callable
}

var validatorProgram = goja.MustCompile("guardrail/validate", `(() => {
	const stringify = JSON.stringify, isArray = Array.isArray;
	return value => {
		if (typeof value === "boolean") return value;
		if (!isArray(value)) throw Error("rule must return a boolean or an array of findings");
		if (value.length > 1000) throw Error("rule exceeded 1000 findings per file");
		const result = stringify(value);
		if (result === "null" || typeof result !== "string") throw Error("invalid findings");
		return result;
	};
})()`, false)

func newJSSession(runtime *goja.Runtime, content, syntaxJSON string, snapshot syntaxSnapshot) (*jsSession, error) {
	validator, err := runtime.RunProgram(validatorProgram)
	if err != nil {
		return nil, err
	}
	validate, ok := goja.AssertFunction(validator)
	if !ok {
		return nil, errors.New("cannot initialize result validation")
	}
	factory, err := runtime.RunProgram(syntaxProgram)
	if err != nil {
		return nil, err
	}
	buildAPI, ok := goja.AssertFunction(factory)
	if !ok {
		return nil, errors.New("cannot initialize syntax API")
	}
	api, err := buildAPI(goja.Undefined(), runtime.ToValue(syntaxJSON), runtime.ToValue(content),
		runtime.ToValue(snapshot.Language), runtime.ToValue(snapshot.Parsed), runtime.ToValue(snapshot.Error))
	if err != nil {
		return nil, err
	}
	changeFactory, err := runtime.RunProgram(changeProgram)
	if err != nil {
		return nil, err
	}
	buildChange, ok := goja.AssertFunction(changeFactory)
	if !ok {
		return nil, errors.New("cannot initialize change API")
	}
	return &jsSession{validate: validate, api: api, buildChange: buildChange}, nil
}
