package cli

import (
	"strings"
)

func (c *Client) resolveServerTypeId(handle string) (string, error) {
	if handle == "" {
		return "", nil
	}
	if strings.HasPrefix(handle, "typ-") {
		return handle, nil
	}
	stype, err := c.ServerTypeByHandle(handle)
	if err != nil {
		return "", err
	}
	return stype.Id, nil
}
