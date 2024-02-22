package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

type Record struct {
	Count             int
	OldestMessageTime time.Time
	IsMuted           bool
}

type List struct {
	records []Record
}

var list map[string]Record
var insults = []string{"يا مغفل", "يا ثرثار", "يا مزعج", "يا متطفل", "يا ضعيف الإرادة", "يا أحمق", "يا متعجرف"}
var MAX_MESSAGES = 10

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	token := os.Getenv("BOT_TOKEN")

	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal(err)
	}
	discord.AddHandler(messageCreate)
	discord.AddHandler(ready)
	discord.Identify.Intents = discordgo.IntentsGuildMessages
	err = discord.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	list = make(map[string]Record)

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	discord.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.Author.ID == "679348712472051715" {
		if st := strings.Split(m.Content, " "); st[0] == "!max" {
			i, err := strconv.Atoi(st[1])
			if err == nil {
				MAX_MESSAGES = i
				s.ChannelMessageSendReply(m.ChannelID, fmt.Sprintf("تم تغيير الحد الأقصى للرسائل إلى %d", i), m.Reference())
			}
		}
	}
	if r, ok := list[m.Author.ID]; !ok {
		list[m.Author.ID] = Record{Count: 1, OldestMessageTime: m.Timestamp}
		return
	} else {
		if r.OldestMessageTime.Before(time.Now().Add(-10 * time.Minute)) { // reset
			list[m.Author.ID] = Record{Count: 1, OldestMessageTime: m.Timestamp}
		} else {
			list[m.Author.ID] = Record{Count: list[m.Author.ID].Count + 1, OldestMessageTime: r.OldestMessageTime}
		}
	}
	list[m.Author.ID] = Record{Count: list[m.Author.ID].Count + 1, OldestMessageTime: m.Timestamp}
	if list[m.Author.ID].Count > 10 {
		t := time.Now().Add(10 * time.Minute)
		s.GuildMemberTimeout(m.GuildID, m.Author.ID, &t)
		s.ChannelMessageSend(m.ChannelID,
			fmt.Sprintf("<@%s>\nلقد تجاوزت الحد الأقصى من الرسائل خلال 10 دقائق, %s", m.Author.ID, insults[rand.Intn(len(insults))]))
		list[m.Author.ID] = Record{Count: 1, OldestMessageTime: m.Timestamp}
	}
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	fmt.Println(fmt.Sprintf("Connected as %s#%s", event.User.Username, event.User.Discriminator))
}
