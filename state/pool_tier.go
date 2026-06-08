package state

func PoolTier(cpuMillis, memoryMiB int64) string {
	switch {
	case cpuMillis >= 30000 || memoryMiB >= 200000:
		return "xl"
	case cpuMillis >= 14000 || memoryMiB >= 100000:
		return "l"
	case cpuMillis >= 3000 || memoryMiB >= 24000:
		return "m"
	default:
		return "s"
	}
}
