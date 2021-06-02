package bot

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gempir/go-twitch-irc/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ChatBot struct {
	client     *twitch.Client
	moderators map[string]struct{}
	channel    string
	collection *mongo.Collection
	DocId      primitive.ObjectID
}

func (b *ChatBot) InitBot(username, oauth string, collection *mongo.Collection) {
	id := os.Getenv("DOC_ID")
	channel := os.Getenv("CHANNEL")

	b.client = twitch.NewClient(username, oauth)
	b.DocId, _ = primitive.ObjectIDFromHex(id)
	b.collection = collection
	b.moderators = map[string]struct{}{
		"169909190": {}, // Id1Syda
		"67643748":  {}, // kosti4eg
		"101478730": {}, // lev_majorov
		"37583788":  {}, // no_ice_coffee
		"156468036": {}, // RACCOONPOTASKUN
		"141955769": {}, //Hamstersher
		"49514558":  {}, // JuggernautsKiller
		"115141884": {}, // GRPZDC
	}
	b.channel = channel

	b.client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		if b.isModerator(message.User.ID) {
			msg := strings.Fields(message.Message)

			// probably useless check
			if len(msg) == 0 {
				return
			}

			command := msg[0]
			if command == "!ban" {
				if len(msg) < 3 {
					b.client.Say(b.channel, fmt.Sprint(ErrNotEnoughArgs))
					return
				}

				from := msg[1]
				to := msg[2]
				followers, err := b.getFollowersSlice(from, to)
				if err != nil {
					b.client.Say(b.channel, fmt.Sprint(err.Error()))
					return
				}

				for _, v := range followers {
					b.client.Say(b.channel, fmt.Sprintf("/ban %s", v.UserLogin))
				}
			}
		}
	})
}

func (b *ChatBot) Monitor() {
	b.client.Join(b.channel)
	b.client.Connect()
}

func (b *ChatBot) isModerator(id string) bool {
	_, ok := b.moderators[id]
	return ok
}

func (b *ChatBot) getFollowersSlice(from, to string) ([]*Follower, error) {
	who := bson.M{"_id": b.DocId}
	resp := new(Followers)

	b.collection.FindOne(context.Background(), who).Decode(resp)
	sl, err := resp.GetSlice(from, to)
	return sl, err
}
