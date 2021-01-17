// Package iworkout joins Cold Brew Crew's Discord server and listens for emoji reactions in the #iworkout channel.
// Use this link to make the bot join your server: https://discord.com/api/oauth2/authorize?client_id=709797518032240751&scope=bot
// Optionally add a &permissions=ABC parameter, where ABC is a bitwise flag as so: https://discord.com/developers/docs/topics/permissions#permissions-bitwise-permission-flags
package iworkout

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

const startMsg = "Starting Discord handlers..."
const logPrefix = "Discord:"

const iworkoutChannelID = "492391546411417620"

// init is called automatically on boot.
func init() {
	botToken, ok := os.LookupEnv("DISCORD_TOKEN")
	if !ok {
		panic("You must set DISCORD_TOKEN")
	}

	discord, err := discordgo.New(fmt.Sprintf("Bot %s", botToken))
	if err != nil {
		log.Fatalf("%s can't start: %s.\n", startMsg, err)
		return
	}

	discord.AddHandler(handleReactionAdd)
	discord.AddHandler(handleReactionRemove)

	if err := discord.Open(); err != nil {
		log.Fatalf("%s can't open: %s.\n", startMsg, err)
		return
	}

	fmt.Printf("%s ready.\n", startMsg)

	go buildMessagesState(discord, iworkoutChannelID)
}

// handleReactionAdd is called by discordgo when an emoji reaction is added to a message.
func handleReactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	u := fetchUser(s, r.UserID)
	m := fetchMessage(s, r.ChannelID, r.MessageID)

	if _, ok := reactions[m.ID]; !ok {
		reactions[m.ID] = map[string]struct{}{}
	}

	reactions[m.ID][u.ID] = struct{}{}

	fmt.Printf("%s reaction added: \"%s\" had a %s added by %v\n", logPrefix, m.Content, r.Emoji.Name, u.Username)
}

// handleReactionRemove is called by discordgo when an emoji reaction is removed from a message.
func handleReactionRemove(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
	u := fetchUser(s, r.UserID)
	m := fetchMessage(s, r.ChannelID, r.MessageID)

	delete(reactions[m.ID], u.ID)

	fmt.Printf("%s reaction removed: \"%s\" had a %s removed by %v\n", logPrefix, m.Content, r.Emoji.Name, u.Username)
}

// fetchUser gets a user through the Discord API by ID and caches it, if it is not already stored in the local cache.
func fetchUser(s *discordgo.Session, id string) discordgo.User {
	if u, ok := users[id]; ok {
		return u
	}

	u, err := s.User(id)
	if err != nil {
		log.Printf("Couldn't fetch user %s: %s", id, err)
		return discordgo.User{}
	}

	users[id] = *u

	return *u
}

// fetchMessage gets a message through the Discord API by ID and caches it, if it is not already stored in the local cache.
func fetchMessage(s *discordgo.Session, channelID, messageID string) discordgo.Message {
	if m, ok := messages[channelID][messageID]; ok {
		return m
	}

	m, err := s.ChannelMessage(channelID, messageID)
	if err != nil {
		log.Printf("Couldn't fetch message %s/%s: %s", channelID, messageID, err)
		return discordgo.Message{}
	}

	if messages[channelID] == nil {
		messages[channelID] = map[string]discordgo.Message{}
	}
	messages[channelID][messageID] = *m

	return *m
}

func buildMessagesState(s *discordgo.Session, channelID string) {
	const limit = 100
	var earliest, latest discordgo.Message
	fmt.Println("Building message state...")

	for {
		time.Sleep(600 * time.Millisecond)
		m, err := s.ChannelMessages(channelID, limit, earliest.ID, "", "")
		if err != nil {
			log.Printf("Couldn't fetch messages: %s", err)
		}

		if len(m) == 0 {
			fmt.Printf("Done cataloguing messages.\n\tEarliest: %s by %v\n\tLatest:   %s by %v\n", earliest.Content, earliest.Author, latest.Content, latest.Author)
			break
		}

		for _, msg := range m {
			if _, ok := messages[msg.ChannelID]; !ok {
				messages[msg.ChannelID] = map[string]discordgo.Message{}
			}

			messages[msg.ChannelID][msg.ID] = *msg
			users[msg.Author.ID] = *msg.Author

			if earliest.ID == "" {
				earliest = *msg
			}
			if latest.ID == "" {
				latest = *msg
			}

			t, err := msg.Timestamp.Parse()
			if err != nil {
				log.Printf("Can't parse returned message timestamp: %s", err)
			}

			et, err := earliest.Timestamp.Parse()
			if err != nil {
				log.Printf("Can't parse earliest message timestamp: %s", err)
			}

			if t.Before(et) {
				earliest = *msg
			}
			if t.After(et) {
				latest = *msg
			}
		}
		fmt.Printf("Catalogued %d messages (to %s)\n", len(m), earliest.Timestamp)
	}

	buildReactionsState(s)
}

func buildReactionsState(s *discordgo.Session) {
	const limit = 100
	fmt.Println("Building reaction state...")

	for channelID, msgs := range messages {
		for _, msg := range msgs {
			for _, reaction := range msg.Reactions {
				time.Sleep(600 * time.Millisecond)
				users, err := s.MessageReactions(channelID, msg.ID, reaction.Emoji.APIName(), limit)
				if err != nil {
					log.Printf("Can't get users who reacted with %s to %s: %s\n", reaction.Emoji.Name, msg.Content, err)
				}

				for _, user := range users {
					if _, ok := reactions[msg.ID]; !ok {
						reactions[msg.ID] = map[string]struct{}{}
					}
					reactions[msg.ID][user.ID] = struct{}{}
				}
				fmt.Printf("Catalogued %d reactions (%d emoji) to %s by %s\n", len(users), len(msg.Reactions), strings.Replace(msg.Content, "\n", " ", -1), msg.Author)
			}
		}
	}

	fmt.Printf("Done cataloguing reactions.\n")
}
