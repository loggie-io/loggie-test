package unit

import (
	"fmt"
	"strconv"
)

func BytesToMiB(val float64) float64 {
	mib := val / 1024 / 1024
	if ret, err := strconv.ParseFloat(fmt.Sprintf("%.2f", mib), 64); err == nil {
		return ret
	}
	return mib
}
