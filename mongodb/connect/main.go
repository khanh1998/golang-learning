package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Ingredient struct {
	ID   primitive.ObjectID `bson:"_id,omitempty"`
	Name string             `bson:"name,omitempty"`
	Unit string             `bson:"unit,omitempty"`
	V    int                `bson:"__v,omitempty"`
}

func getMongodbURI() string {
	var myEnv map[string]string

	myEnv, err := godotenv.Read()
	if err != nil {
		log.Fatal(err)
	}
	return myEnv["MONGODB_URI"]
}

func getConnection() (*mongo.Client, context.Context) {
	mongodb_uri := getMongodbURI()
	client, err := mongo.NewClient(options.Client().ApplyURI(mongodb_uri))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	return client, ctx
}

func testConnection(client mongo.Client, ctx context.Context) bool {
	err := client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
		return false
	}
	return true
}

func GetDatabaseNames(client mongo.Client, ctx context.Context) []string {
	databases, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	return databases
}

func GetCollectionNames(database mongo.Database, ctx context.Context) []string {
	collectionNames, err := database.ListCollectionNames(ctx, bson.D{}, options.ListCollections())
	if err != nil {
		log.Fatal(err)
	}
	return collectionNames
}

func GetDocuments(collection mongo.Collection, ctx context.Context, filter interface{}) []Ingredient {
	var ingredients []Ingredient
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}
	if err = cursor.All(ctx, &ingredients); err != nil {
		log.Panic(err)
	}
	return ingredients
}
func DeleteOne(collection mongo.Collection, ctx context.Context, id string) bool {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Fatal(err)
	}
	filter := bson.M{
		"_id": objectId,
	}
	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}
	return result.DeletedCount == 1
}
func GetOne(collection mongo.Collection, ctx context.Context, id string) Ingredient {
	var ingredient Ingredient
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Fatal(err)
	}
	if err := collection.FindOne(ctx, bson.M{"_id": objectId}).Decode(&ingredient); err != nil {
		log.Fatal(err)
	}
	return ingredient
}
func UpdateOne(collection mongo.Collection, ctx context.Context, id string, ingredient Ingredient) bool {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Fatal(err)
	}
	data := bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "name", Value: ingredient.Name},
			primitive.E{Key: "unit", Value: ingredient.Unit},
		}},
	}
	res, err := collection.UpdateByID(ctx, objectId, data)
	if err != nil {
		log.Fatal(err)
	}
	return res.ModifiedCount == 1
}
func AddDocument(collection mongo.Collection, ctx context.Context, ingredient Ingredient) string {
	result, err := collection.InsertOne(ctx, ingredient)
	if err != nil {
		log.Panic(err)
	}
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		return oid.Hex()
	} else {
		log.Fatal("we have trouble with getting id of inserted item")
		return ""
	}
}
func main() {
	client, ctx := getConnection()
	defer client.Disconnect(ctx)
	success := testConnection(*client, ctx)
	if success {
		fmt.Println("connect to database successfully")
	}
	databaseNames := GetDatabaseNames(*client, ctx)
	fmt.Println(databaseNames)
	todoList := client.Database("cooking_recipe")
	collectionNames := GetCollectionNames(*todoList, ctx)
	fmt.Print(collectionNames)
	todoItems := todoList.Collection("ingredients")
	ingredients := GetDocuments(*todoItems, ctx, bson.D{})
	fmt.Println(ingredients)
	fish := Ingredient{
		Name: "FISH",
		Unit: "Kilo",
	}
	insertedId := AddDocument(*todoItems, ctx, fish)
	fmt.Println(insertedId)
	success = DeleteOne(*todoItems, ctx, insertedId)
	if success {
		fmt.Println("delete item successfully")
	}
	fmt.Println("Get ingredient by ID")
	ingredient := GetOne(*todoItems, ctx, "60696ba7f08a5d110bcdfb00")
	fmt.Println(ingredient)
	success = UpdateOne(*todoItems, ctx, "60696ba7f08a5d110bcdfb00", fish)
	if success {
		fmt.Println("Update data successfully")
	}
}
