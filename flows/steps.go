package flows

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/amezianechayer/corren/core"
	sharia "github.com/amezianechayer/corren-sharia"
	"github.com/amezianechayer/corren/wallets"
)

type stepOutcome struct {
	kind     int
	resumeAt string
}

const (
	outcomeNext = iota
	outcomeStop
	outcomeWait
)

func (e *Engine) execStep(step Step, input map[string]interface{}) (stepOutcome, error) {
	switch step.Type {
	case StepTransfer:
		return stepOutcome{kind: outcomeNext}, e.executeTransfer(step.Config, input)
	case StepSharia:
		return stepOutcome{kind: outcomeNext}, e.executeSharia(step.Config, input)
	case StepCondition:
		pass, err := e.executeCondition(step.Config, input)
		if err != nil {
			return stepOutcome{}, err
		}
		if !pass {
			return stepOutcome{kind: outcomeStop}, nil
		}
		return stepOutcome{kind: outcomeNext}, nil
	case StepWait:
		resumeAt, err := executeWait(step.Config)
		if err != nil {
			return stepOutcome{}, err
		}
		return stepOutcome{kind: outcomeWait, resumeAt: resumeAt}, nil
	default:
		return stepOutcome{}, newError(ErrInvalidStep, "unknown step type: "+step.Type)
	}
}

// executeTransfer runs a FaRl script, substituting $vars from the instance
// input first. The posting goes through the ledger (and the Guard).
func (e *Engine) executeTransfer(cfg json.RawMessage, input map[string]interface{}) error {
	var c transferConfig
	if err := json.Unmarshal(cfg, &c); err != nil || c.Script == "" {
		return newError(ErrInvalidStep, "transfer step needs a script")
	}
	if err := e.led.Execute(core.Script{Plain: substitute(c.Script, input)}); err != nil {
		return newError(ErrStepFailed, "transfer failed: "+err.Error())
	}
	return nil
}

// executeSharia creates or advances a sharia contract. Params/id are templated
// with $vars from the input, so a flow can build a contract from its trigger.
func (e *Engine) executeSharia(cfg json.RawMessage, input map[string]interface{}) error {
	var c shariaConfig
	if err := json.Unmarshal(cfg, &c); err != nil {
		return newError(ErrInvalidStep, "invalid sharia step")
	}
	switch c.Action {
	case "", "create":
		params := json.RawMessage(substitute(string(c.Params), input))
		if _, _, err := e.sharia.Create(sharia.CreateRequest{Type: c.ContractType, ID: substitute(c.ID, input), Params: params}); err != nil {
			return newError(ErrStepFailed, "sharia create failed: "+err.Error())
		}
	case "transition":
		if _, err := e.sharia.Transition(substitute(c.ID, input), c.Transition, sharia.TransitionInput{}); err != nil {
			return newError(ErrStepFailed, "sharia transition failed: "+err.Error())
		}
	default:
		return newError(ErrInvalidStep, "unknown sharia action: "+c.Action)
	}
	return nil
}

// executeCondition evaluates "<field> <op> <int>"; field is an input key (e.g.
// amount) or "balance" (the credited wallet's main balance). A false result
// stops the flow.
func (e *Engine) executeCondition(cfg json.RawMessage, input map[string]interface{}) (bool, error) {
	var c conditionConfig
	if err := json.Unmarshal(cfg, &c); err != nil || c.Expr == "" {
		return false, newError(ErrInvalidStep, "condition step needs an expr")
	}
	fields := strings.Fields(c.Expr)
	if len(fields) != 3 {
		return false, newError(ErrInvalidStep, "condition must be '<field> <op> <int>'")
	}
	right, err := strconv.ParseInt(fields[2], 10, 64)
	if err != nil {
		return false, newError(ErrInvalidStep, "condition right side must be an integer")
	}
	left, err := e.resolveLeft(fields[0], input)
	if err != nil {
		return false, err
	}
	switch fields[1] {
	case ">":
		return left > right, nil
	case ">=":
		return left >= right, nil
	case "<":
		return left < right, nil
	case "<=":
		return left <= right, nil
	case "==":
		return left == right, nil
	case "!=":
		return left != right, nil
	default:
		return false, newError(ErrInvalidStep, "unknown operator: "+fields[1])
	}
}

func (e *Engine) resolveLeft(field string, input map[string]interface{}) (int64, error) {
	if field == "balance" {
		walletID, _ := input["wallet_id"].(string)
		asset, _ := input["asset"].(string)
		if walletID == "" || asset == "" {
			return 0, newError(ErrInvalidStep, "balance condition needs wallet_id and asset in input")
		}
		acct, err := e.led.GetAccount(wallets.MainAccount(walletID))
		if err != nil {
			return 0, newError(ErrStepFailed, "balance lookup failed: "+err.Error())
		}
		return acct.Balances[asset], nil
	}
	v, ok := input[field]
	if !ok {
		return 0, newError(ErrInvalidStep, "condition field not found in input: "+field)
	}
	return toInt64(v), nil
}

func executeWait(cfg json.RawMessage) (string, error) {
	var c waitConfig
	if err := json.Unmarshal(cfg, &c); err != nil || c.Days <= 0 {
		return "", newError(ErrInvalidStep, "wait step needs days > 0")
	}
	return addDaysUTC(c.Days), nil
}

// substitute replaces $key tokens with the input values, in two passes so that
// a flow definition stays valid JSON:
//   1. a whole quoted placeholder "$key" becomes the typed JSON value, so a
//      number stays a number ("amount":"$amount" -> "amount":100000);
//   2. a bare $key inside a string is replaced with its stringified value
//      (e.g. the FaRl script [$amount], or "@client:$wallet_id").
// Keys are applied longest-first so $sender does not clobber $senderId.
func substitute(template string, input map[string]interface{}) string {
	if template == "" || len(input) == 0 {
		return template
	}
	keys := make([]string, 0, len(input))
	for k := range input {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return len(keys[i]) > len(keys[j]) })
	out := template
	for _, k := range keys {
		if b, err := json.Marshal(input[k]); err == nil {
			out = strings.ReplaceAll(out, `"$`+k+`"`, string(b))
		}
	}
	for _, k := range keys {
		out = strings.ReplaceAll(out, "$"+k, stringify(input[k]))
	}
	return out
}

func stringify(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatInt(int64(t), 10)
	case int64:
		return strconv.FormatInt(t, 10)
	case bool:
		return strconv.FormatBool(t)
	default:
		return fmt.Sprintf("%v", t)
	}
}

func toInt64(v interface{}) int64 {
	switch t := v.(type) {
	case float64:
		return int64(t)
	case int64:
		return t
	case string:
		n, _ := strconv.ParseInt(t, 10, 64)
		return n
	default:
		return 0
	}
}
