package analysis

import (
	_ "embed"

	"github.com/dop251/goja"
)

//go:embed change_api.js
var changeAPISource string

//go:embed function_models.js
var functionModelsSource string

//go:embed import_models.js
var importModelsSource string

//go:embed comment_models.js
var commentModelsSource string

var changeAPI = "(" + changeAPISource + ")(" + functionModelsSource + "," + importModelsSource + "," + commentModelsSource + ")"

var changeProgram = goja.MustCompile("guardrail/change", changeAPI, true)
var syntaxProgram = goja.MustCompile("guardrail/ast", "("+syntaxAPI+")", true)
