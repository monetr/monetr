package zoneinfo

import "time"

var (
	aliases = map[string]string{}
)

// Timezone will return a time location but will handle aliases or updated time
// zone names. If a timezone name specified has been changed to instead be a new
// name, this will automatically use the new name instead.
func Timezone(timezone string) (*time.Location, error) {
	newName, ok := aliases[timezone]
	if !ok {
		return time.LoadLocation(timezone)
	}

	return time.LoadLocation(newName)
}
