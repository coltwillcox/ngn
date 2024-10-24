package conductor

import (
	"fmt"
)

func fmtCmd[T any](cmd T) (cmdStr string) {
	if stringer, ok := any(cmd).(fmt.Stringer); ok {
		cmdStr = stringer.String()
	} else {
		cmdStr = fmt.Sprintf("%v", cmd)
	}

	return
}
