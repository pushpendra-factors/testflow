package util

func UseQueue(token string, queueAllowedTokens []string) bool {
	// allow all for wildcard(asterisk).
	if len(queueAllowedTokens) == 1 && queueAllowedTokens[0] == "*" {
		return true
	}

	for _, allowedToken := range queueAllowedTokens {
		if token == allowedToken {
			return true
		}
	}

	return false
}
