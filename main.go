package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"

	//	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

type Food struct {
	ID    primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name  string             `json:"name,omitempty" bson:"name,omitempty"`
	Tribe string             `json:"tribe,omitempty" bson:"tribe,omitempty"`
}

func TestFood(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	resp, err := http.Get("http://gobyexample.com")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	response.WriteHeader(http.StatusAccepted)
	//fmt.Println("Response status:", resp.Status)
	response.Write([]byte(`{ "reponseClient": "` + resp.Status + `" }`))

	scanner := bufio.NewScanner(resp.Body)
	for i := 0; scanner.Scan() && i < 5; i++ {
		fmt.Println(scanner.Text())
		response.Write([]byte(`{ "messageClient": "` + scanner.Text() + `" }`))
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

}

func AddFood(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var food Food
	_ = json.NewDecoder(request.Body).Decode(&food)
	fmt.Println(food.Name)
	collection := client.Database("myfood").Collection("foods")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	result, _ := collection.InsertOne(ctx, food)
	json.NewEncoder(response).Encode(result)
}

func GetFood(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	params := mux.Vars(request)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var food Food
	collection := client.Database("myfood").Collection("foods")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	err := collection.FindOne(ctx, Food{ID: id}).Decode(&food)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(food)
}

func GetFoods(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var foods []Food
	collection := client.Database("myfood").Collection("foods")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var food Food
		cursor.Decode(&food)
		foods = append(foods, food)
	}
	if err := cursor.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}

	json.NewEncoder(response).Encode(foods)

}

func main() {
	fmt.Println("Starting the application on port 8080")

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	//clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	clientOptions := options.Client().ApplyURI("mongodb://mongodb:27017")
	client, _ = mongo.Connect(ctx, clientOptions)
	// Connect to MongoDB
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	err = client.Ping(context.TODO(), nil)

	//	if err != nil {
	//  		log.Fatal(err)

	//	}
	fmt.Println("Connectd to MondoDB")

	router := mux.NewRouter()
	router.HandleFunc("/food", AddFood).Methods("POST")
	router.HandleFunc("/test", TestFood).Methods("POST")
	router.HandleFunc("/food", GetFoods).Methods("GET")
	router.HandleFunc("/food/{id}", GetFood).Methods("GET")
	http.ListenAndServe(":8080", router)
}
