package utils

import "time"

func GetTimeUs() int64 {
	return time.Now().UnixMicro()
}
