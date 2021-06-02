package main

import (
	"context"
	bot "grpBot/internal"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	godotenv.Load()

	dbURI := os.Getenv("DB_URL")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dbURI))

	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	dbName := os.Getenv("DB_NAME")
	dbCol := os.Getenv("DB_COL")
	collection := client.Database(dbName).Collection(dbCol)

	username := os.Getenv("TWITCH_USERNAME")
	oauth := os.Getenv("TWITCH_OAUTH")
	chatBot := bot.ChatBot{}
	chatBot.InitBot(username, oauth, collection)
	go func() {
		chatBot.Monitor()
	}()

	eventSub := bot.EventSub{}
	eventSub.InitEventSub(collection)
	eventSub.Start()
}
