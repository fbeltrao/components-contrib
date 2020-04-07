package bindings

import (
	"fmt"
	"strconv"
	"time"
)

const (
	ttlMetadataKey = "ttl"
)

// TryGetTTL tries to get the ttl (in seconds) value for a binding
func TryGetTTL(props map[string]string) (time.Duration, bool, error) {
	if val, ok := props[ttlMetadataKey]; ok && val != "" {
		valInt, err := strconv.Atoi(val)
		if err != nil {
			return 0, false, err
		}

		if valInt <= 0 {
			return 0, false, fmt.Errorf("%s value must be higher than zero: actual is %d", ttlMetadataKey, valInt)
		}

		return time.Duration(valInt) * time.Second, true, nil
	}

	return 0, false, nil
}
