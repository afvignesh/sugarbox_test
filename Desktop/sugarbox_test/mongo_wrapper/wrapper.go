package mongo_wrapper

import (
    "fmt"
	"os"
	// "github.com/joho/godotenv"
    "go.mongodb.org/mongo-driver/mongo"
    // "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"context"

	c "../constants"
	
)

var db *mongo.Database


func init() {
    ctx := context.Background()
    mongoURI := fmt.Sprintf("mongodb+srv://" + os.Getenv("mongo_username")+ ":" + os.Getenv("mongo_pass") +"@cluster0-yqtrj.mongodb.net/test?retryWrites=true&w=majority")

    // Set client options and connect
    clientOptions := options.Client().ApplyURI(mongoURI)
    client, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
	}

    db = client.Database(c.DB_NAME);

}
//returns a handle to the DB object
func GetDB() *mongo.Database {
	return db
}

func MongoFindOneRecord(collectionName string, condition interface{}, decodeTo interface{}) error{
   err := GetDB().Collection(collectionName).FindOne(context.TODO(), condition).Decode(decodeTo)
   return err
}

func MongoFindMultipleRecord(collectionName string, condition interface{}) (*mongo.Cursor, error){
    cur, err := GetDB().Collection(collectionName).Find(context.TODO(), condition)
    // return "nice", err
    return cur, err
}


func MongoInsertOneRecord(collectionName string, body interface{}) (*mongo.InsertOneResult, error) {
    result, err := GetDB().Collection(collectionName).InsertOne(context.TODO(), body)
    return result, err
}

func MongoUpdateOneRecord(collectionName string, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	result, err := GetDB().Collection(collectionName).UpdateOne(context.TODO(), filter, update)
    return result, err
}

