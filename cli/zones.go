package cli

import (
	"strings"
)

func (c *Client) resolveZoneId(handle string) (string, error) {
	if handle == "" {
		return "", nil
	}
	if strings.HasPrefix(handle, "zon-") {
		return handle, nil
	}
	zone, err := c.ZoneByHandle(handle)
	if err != nil {
		return "", err
	}
	return zone.Id, nil
}
