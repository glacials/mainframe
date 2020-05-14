package iworkout

// reactions is a map of message ID to a "set" of user IDs who reacted to that message (with any reaction).
var reactions = map[string]map[string]struct{}{}

// Reactions returns all reactions we've catalogued as a map of message ID to a "set" of user IDs who have reacted to that message (with any emoji).
func Reactions() map[string]map[string]struct{} {
	return reactions
}
