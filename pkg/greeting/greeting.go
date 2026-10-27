package greeting

import (
	"fmt"
	"strings"
)

func Build(name string, shout bool) string {
	msg := fmt.Sprintf("Hello, %s! 👋", name)
	if shout {
		msg = strings.ToUpper(msg)
	}
	return msg
}
