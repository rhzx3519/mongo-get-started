package examples

import (
    "context"
    "fmt"
    "github.com/joho/godotenv"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "log"
    "os"
    "sort"
    "sync"
)

var (
    URI = ""
)

func init() {
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found")
    }
    URI = os.Getenv("MONGODB_URI")
}

// start-restaurant-struct
type Restaurant struct {
    Name         string
    RestaurantId string        `bson:"restaurant_id,omitempty"`
    Cuisine      string        `bson:"cuisine,omitempty"`
    Address      interface{}   `bson:"address,omitempty"`
    Borough      string        `bson:"borough,omitempty"`
    Grades       []interface{} `bson:"grades,omitempty"`
}

func ExampleInsert() {
    client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(URI))
    if err != nil {
        panic(err)
    }
    defer func() {
        if err := client.Disconnect(context.TODO()); err != nil {
            panic(err)
        }
    }()

    // Inserts a sample document describing a restaurant into the collection
    // begin insertOne
    coll := client.Database("sample_restaurants").Collection("restaurants")
    newRestaurant := Restaurant{Name: "8282", Cuisine: "Korean"}

    _, err = coll.InsertOne(context.TODO(), newRestaurant)
    if err != nil {
        panic(err)
    }
    // end insertOne

    // Retrieves the first matching document
    var r Restaurant
    filter := bson.D{{"name", "8282"}}
    err = coll.FindOne(context.TODO(), filter).Decode(&r)
    // Prints a message if no documents are matched or if any
    // other errors occur during the operation
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return
        }
        panic(err)
    }

    defer func() {
        // Deletes all documents that have a "runtime" value greater than 800
        _, err := coll.DeleteMany(context.TODO(), filter)
        if err != nil {
            panic(err)
        }
    }()

    fmt.Println(r.Name, r.Cuisine)
    // Output:
    // 8282 Korean
}

func ExampleConcurrentInsert() {
    client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(URI))
    if err != nil {
        panic(err)
    }
    defer func() {
        if err := client.Disconnect(context.TODO()); err != nil {
            panic(err)
        }
    }()

    // Inserts a sample document describing a restaurant into the collection
    // begin insertOne
    coll := client.Database("sample_restaurants").Collection("restaurants")

    var wg sync.WaitGroup
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(i int) {
            defer wg.Done()
            newRestaurant := Restaurant{Name: fmt.Sprintf("%v", i), Cuisine: "Korean"}
            _, err = coll.InsertOne(context.TODO(), newRestaurant)
            if err != nil {
                panic(err)
            }
        }(i)
    }

    wg.Wait()

    filter := bson.D{{"cuisine", "Korean"}}
    cursor, err := coll.Find(context.TODO(), filter)
    if err != nil {
        panic(err)
    }

    var results []Restaurant
    if err = cursor.All(context.TODO(), &results); err != nil {
        panic(err)
    }

    // Prints the results of the find operation as structs
    sort.Slice(results, func(i, j int) bool {
        return results[i].Name < results[j].Name
    })
    for _, result := range results {
        cursor.Decode(&result)
        fmt.Printf("%s\n", result.Name)
    }

    defer func() {
        // Deletes all documents that have a "runtime" value greater than 800
        _, err := coll.DeleteMany(context.TODO(), filter)
        if err != nil {
            panic(err)
        }
    }()

    // Output:
    // 0
    // 1
    // 2
    // 3
    // 4
}
