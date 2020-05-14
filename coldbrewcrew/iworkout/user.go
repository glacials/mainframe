package iworkout

import "github.com/bwmarrin/discordgo"

// users is a map of user ID to user.
var users = map[string]discordgo.User{}

// Users returns all users we've catalogued as a map of user ID to user.
func Users() map[string]discordgo.User {
	return users
}
