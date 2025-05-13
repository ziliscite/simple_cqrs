package command

import (
	"fmt"
	"strings"
)

type Errs map[string]error

func (e Errs) Error() string {
	var errsStr []string
	for k, v := range e {
		errsStr = append(errsStr, fmt.Sprintf("%s: %s", k, v.Error()))
	}
	return strings.Join(errsStr, "; ")
}
