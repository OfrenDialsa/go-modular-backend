package utils

func IsValidEmail(email string) bool {
	if len(email) < 3 {
		return false
	}

	atIndex := -1
	for i, c := range email {
		if c == '@' {
			atIndex = i
			break
		}
	}

	if atIndex <= 0 || atIndex >= len(email)-1 {
		return false
	}

	return true
}
