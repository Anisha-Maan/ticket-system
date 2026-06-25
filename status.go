package main

func IsValidStatus(status string) bool {
	switch status {
	case "open", "in_progress", "closed":
		return true
	default:
		return false
	}
}

func CanTransition(current, next string) bool {
	switch current {

	case "open":
		return next == "in_progress"

	case "in_progress":
		return next == "closed"

	case "closed":
		return false
	}

	return false
}