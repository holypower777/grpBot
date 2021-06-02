package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type EventSub struct {
	collection *mongo.Collection
	Router     *mux.Router
	DocId      primitive.ObjectID
}

type Response struct {
	Challenge    string `json:"challenge,omitempty"`
	Subscription struct {
		Id        string `json:"id,omitempty"`
		Status    string `json:"status,omitempty"`
		Type      string `json:"type,omitempty"`
		Version   string `json:"version,omitempty"`
		Cost      int    `json:"cost,omitempty"`
		Condition struct {
			BroadcasterId string `json:"broadcaster_user_id,omitempty"`
		} `json:"condition,omitempty"`
		Transport struct {
			Method   string `json:"method,omitempty"`
			Callback string `json:"callback,omitempty"`
		} `json:"transport,omitempty"`
		CreatedAt string `json:"created_at,omitempty"`
	} `json:"subscription,omitempty"`
	Event struct {
		UserId               string `json:"user_id,omitempty"`
		UserLogin            string `json:"user_login,omitempty"`
		UserName             string `json:"user_name,omitempty"`
		BroadcasterId        string `json:"broadcaster_user_id,omitempty"`
		BroadcasterUserLogin string `json:"broadcaster_user_login,omitempty"`
		BroadcasterUserName  string `json:"broadcaster_user_name,omitempty"`
	} `json:"event,omitempty"`
}

type Follower struct {
	UserId    string `bson:"user_id"`
	UserLogin string `bson:"user_login"`
	UserName  string `bson:"user_name"`
}

type Followers struct {
	Followers []*Follower `bson:"followers"`
}

func (f *Followers) GetSlice(from, to string) ([]*Follower, error) {
	var startIndex int
	var endIndex int
	startFound, endFound := false, false
	for k, v := range f.Followers {
		if v.UserLogin == from {
			startIndex = k
			startFound = true
		}
		if v.UserLogin == to {
			endIndex = k + 1
			endFound = true
			break
		}
	}

	if !startFound {
		return nil, &ErrUserNotFound{from}
	}

	if !endFound {
		return nil, &ErrUserNotFound{to}
	}

	return f.Followers[startIndex:endIndex], nil
}

func (e *EventSub) HandlePing(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Pong")
}

func (e *EventSub) HandleWebhookCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST request is applicable", http.StatusNotFound)
		return
	}

	resp := new(Response)
	err := json.NewDecoder(r.Body).Decode(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	//TODO: make signature verification

	if resp.Challenge != "" {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, resp.Challenge)
		return
	}

	who := bson.M{"_id": e.DocId}
	pushToArray := bson.M{"$addToSet": bson.M{"followers": bson.M{
		"user_id":    resp.Event.UserId,
		"user_login": resp.Event.UserLogin,
		"user_name":  resp.Event.UserName,
	}}}

	_, err = e.collection.UpdateOne(context.Background(), who, pushToArray)

	if err != nil {
		log.Print(err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (e *EventSub) InitEventSub(collection *mongo.Collection) {
	id := os.Getenv("DOC_ID")

	e.collection = collection
	e.DocId, _ = primitive.ObjectIDFromHex(id)
	e.Router = mux.NewRouter()
	e.Router.HandleFunc("/webhook/callback", e.HandleWebhookCallback)
	e.Router.HandleFunc("/", e.HandlePing)
}

func (e *EventSub) Start() {
	port := os.Getenv("PORT")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), e.Router))
}
