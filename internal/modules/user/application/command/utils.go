package command

func hashJTI(token string) string {
	return token[:min(32, len(token))]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
