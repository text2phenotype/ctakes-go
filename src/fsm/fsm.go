package fsm

import (
	"text2phenotype.com/fdl/types"
	"errors"
	"fmt"
)

type MachineRule struct {
	Dst  string
	Cond Condition
}

type Machine map[string][]MachineRule

func (fsm Machine) Input(token types.HasSpan, currentState string) string {
	rules, isOk := fsm[currentState]
	if !isOk {
		errTxt := fmt.Sprintf("Wrong rule: there is no transitions from '%s' state", currentState)
		panic(errors.New(errTxt))
	}

	for _, rule := range rules {
		if rule.Cond(token) {
			return rule.Dst
		}
	}

	return currentState
}
