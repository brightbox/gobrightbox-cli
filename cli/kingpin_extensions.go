package cli

import (
	"fmt"
	"strconv"
)

type pStringValue struct {
	Target **string
}

func (ps *pStringValue) Set(s string) error {
	*ps.Target = &s
	return nil
}

func (ps *pStringValue) String() string {
	if ps.Target != nil && *ps.Target != nil {
		return **ps.Target
	} else {
		return ""
	}
}

type pBoolValue struct {
	Target **bool
}

func (ps *pBoolValue) Set(s string) error {
	v, err := strconv.ParseBool(s)
	*ps.Target = &v
	return err
}

func (ps *pBoolValue) String() string {
	if ps.Target != nil && *ps.Target != nil {
		return fmt.Sprintf("%v", **ps.Target)
	} else {
		return fmt.Sprintf("%v", false)
	}
}
