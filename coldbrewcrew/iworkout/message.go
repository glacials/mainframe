package iworkout

import "github.com/bwmarrin/discordgo"

// messages is a two-dimensional map of channel ID to message ID to message.
var messages = map[string]map[string]discordgo.Message{}

// Messages returns all messages we've catalogued as a map of message ID to message.
func Messages() map[string]discordgo.Message {
	return messages[iworkoutChannelID]
}
