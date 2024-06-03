package settings

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

type stringDecoderFunc func(string) (any, error)

type MetaField struct {
	FlagValueP      any
	FlagName        string
	FlagDesc        string
	EnvKey          string
	ActionInputName string
	ActionType      string
	DefaultValue    string
	stringDecoder   stringDecoderFunc
	cobraFunc       func(*MetaField, *cobra.Command)
}

func (m MetaField) EvaluateString() (string, error) {
	v, err := m.Evaluate()
	if err != nil {
		return "", err
	}
	value, ok := v.(string)
	if !ok {
		err := fmt.Errorf("invalid type: %T expected string", v)
		return "", err
	}
	return value, nil
}

func (m MetaField) MustEvaluateString() string {
	value, err := m.EvaluateString()
	if err != nil {
		panic(err)
	}
	return value
}

func (m MetaField) EvaluateInt() (int, error) {
	v, err := m.Evaluate()
	if err != nil {
		return 0, err
	}
	value, ok := v.(int)
	if !ok {
		err := fmt.Errorf("invalid type: %T expected int", v)
		return 0, err
	}
	return value, nil
}

func (m MetaField) MustEvaluateInt() int {
	value, err := m.EvaluateInt()
	if err != nil {
		panic(err)
	}
	return value
}

func (m MetaField) EvaluateBool() (bool, error) {
	v, err := m.Evaluate()
	if err != nil {
		return false, err
	}
	value, ok := v.(bool)
	if !ok {
		err := fmt.Errorf("invalid type: %T expected bool", v)
		return false, err
	}
	return value, nil
}

func (m MetaField) MustEvaluateBool() bool {
	value, err := m.EvaluateBool()
	if err != nil {
		panic(err)
	}
	return value
}

func (m MetaField) MustEvaluateStringSlice() []string {
	value, err := m.EvaluateStringSlice()
	if err != nil {
		panic(err)
	}
	return value
}

func (m MetaField) EvaluateStringSlice() ([]string, error) {
	v, err := m.Evaluate()
	if err != nil {
		return nil, err
	}
	value, ok := v.([]string)
	if !ok {
		err := fmt.Errorf("invalid type: %T expected []string", v)
		return nil, err
	}
	return value, nil
}

// Evaluate returns the runtime value
//
// Order of Precedence:
// 1. A non default flag value
// 2. A non empty os environment variable
// 3. The default value
func (m MetaField) Evaluate() (any, error) {
	// determines order of evaluation
	evalFuncs := []func() (any, error){
		m.evaluateFlag,
		m.evaluateEnv,
		m.evaluateDefault,
	}

	for _, evalFunc := range evalFuncs {
		value, err := evalFunc()
		if err != nil {
			return nil, err
		}
		if value != nil {
			return value, nil
		}
	}

	return nil, nil
}

func (m MetaField) MustEvaluate() any {
	v, err := m.Evaluate()
	if err != nil {
		panic(err)
	}
	return v
}

func (m MetaField) evaluateFlag() (any, error) {
	if m.FlagValueP == nil {
		return nil, nil
	}
	elem := reflect.ValueOf(m.FlagValueP).Elem()
	if elem.IsZero() {
		return nil, nil
	}

	defaultValue, err := m.stringDecoder(m.DefaultValue)
	if err != nil {
		return nil, err
	}

	if reflect.DeepEqual(defaultValue, elem.Interface()) {
		return nil, err
	}

	return elem.Interface(), nil
}

func (m MetaField) evaluateEnv() (any, error) {
	valueStr, exists := os.LookupEnv(m.EnvKey)
	if !exists {
		return nil, nil
	}
	return m.stringDecoder(valueStr)
}

func (m MetaField) evaluateDefault() (any, error) {
	return m.stringDecoder(m.DefaultValue)
}

func stringToStringDecoder(s string) (interface{}, error) {
	return s, nil
}

func stringToIntDecoder(s string) (interface{}, error) {
	return strconv.Atoi(s)
}

func stringToBoolDecoder(s string) (interface{}, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "y", "yes", "1", "true", "on":
		return true, nil
	case "n", "no", "false", "0", "off":
		return false, nil
	default:
		return nil, errors.New("value is not set")
	}
}
