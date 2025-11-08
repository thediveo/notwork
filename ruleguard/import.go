package ruleguard

import (
	"github.com/quasilyte/go-ruleguard/dsl"
	"github.com/thediveo/notwork/ruleguard/notworkrules"
)

func init() {
	dsl.ImportRules("notwork", notworkrules.Bundle)
}
