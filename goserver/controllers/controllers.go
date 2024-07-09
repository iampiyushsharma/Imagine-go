package controllers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"github.com/cloudinary/cloudinary-go"
	"github.com/cloudinary/cloudinary-go/api/uploader"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const dbName = "test"
const colName = "posts"

// MOST IMPORTANT
var collection *mongo.Collection

type PromptRequest struct {
	Prompt string `json:"prompt"`
}

type PostRequest struct {
	Prompt string `json:"prompt"`
	Name   string `json:"name"`
	Photo  string `json:"photo"`
}

type Post struct {
	Name   string `json:"name"`
	Prompt string `json:"prompt"`
	Photo  string `json:"photo"`
}

type Response struct {
	Success bool   `json:"success"`
	Data    []Post `json:"data"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// Read implements io.Reader.
func (PostRequest) Read(p []byte) (n int, err error) {
	panic("unimplemented")
}


func init() {

	godotenv.Load()
	connectionString := os.Getenv("MONGO_URI")
	clientOption := options.Client().ApplyURI(connectionString)

	//connect to mongodb
	client, err := mongo.Connect(context.TODO(), clientOption)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("MongoDB connection success")

	collection = client.Database(dbName).Collection(colName)

	fmt.Println("Collection instance is ready")
}

// helper
func getAllPosts() []primitive.M {
	cur, err := collection.Find(context.Background(), bson.D{{}})
	if err != nil {
		log.Fatal(err)
	}

	var posts []primitive.M

	for cur.Next(context.Background()) {
		var movie bson.M
		err := cur.Decode(&movie)
		if err != nil {
			log.Fatal(err)
		}
		posts = append(posts, movie)
	}

	defer cur.Close(context.Background())
	return posts
}

// controllers
func GetMyAllPosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/x-www-form-urlencode")
	allPosts := getAllPosts()
	json.NewEncoder(w).Encode(allPosts)
}

func GenerateImageFromGetImg(w http.ResponseWriter, r *http.Request) {
	godotenv.Load()
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(http.StatusOK)
		return
	}

	url := "https://api.getimg.ai/v1/essential-v2/text-to-image"
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Read the request body
	var promptRequest PromptRequest
	err := json.NewDecoder(r.Body).Decode(&promptRequest)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	payload := strings.NewReader(fmt.Sprintf(`{"style":"photorealism","prompt":"%s","aspect_ratio":"4:5","output_format":"jpeg","response_format":"b64"}`, promptRequest.Prompt))

	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("authorization", "Bearer "+ os.Getenv("TOKEN_GETIMG"))

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	var responseData map[string]interface{}

	if err := json.Unmarshal(body, &responseData); err != nil {
		http.Error(w, "Failed to decode response", http.StatusInternalServerError)
		return
	}

	imageBase64 := responseData["image"].(string)

	json.NewEncoder(w).Encode(map[string]string{"photo": imageBase64})

}

func PostPhoto(w http.ResponseWriter, r *http.Request) {
	godotenv.Load()
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(http.StatusOK)
		return
	}

	var postRequest PostRequest
	err := json.NewDecoder(r.Body).Decode(&postRequest)
	if err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	imageData, err := base64.StdEncoding.DecodeString(postRequest.Photo)
	if err != nil {
		http.Error(w, "Failed to decode base64 image string", http.StatusBadRequest)
		return
	}

	tempFile, err := os.CreateTemp("", "generated_image_*.jpeg")
	if err != nil {
		http.Error(w, "Failed to create temp file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.Write(imageData); err != nil {
		http.Error(w, "Failed to write image data to temp file", http.StatusInternalServerError)
		return
	}

	cld, err := cloudinary.NewFromURL("cloudinary://"+os.Getenv("CLOUD_API_KEY")+":"+os.Getenv("CLOUD_API_SECRET")+"@"+os.Getenv("CLOUD_NAME"))
	if err != nil {
		http.Error(w, "Failed to create Cloudinary instance", http.StatusInternalServerError)
		return
	}

	var ctx = context.Background()
	uploadResp, err := cld.Upload.Upload(ctx, tempFile.Name(), uploader.UploadParams{})
	if err != nil {
		http.Error(w, "Failed to upload image to Cloudinary", http.StatusInternalServerError)
		return
	}

	var post Post
	post.Name = postRequest.Name
	post.Prompt = postRequest.Prompt
	post.Photo = uploadResp.SecureURL

	inserted, err := collection.InsertOne(context.Background(), post)
	if err != nil {
		http.Error(w, "Failed to insert post into database", http.StatusInternalServerError)
		return
	}

	fmt.Println("Inserted 1 post in db with id: ", inserted.InsertedID)
	fmt.Println("url ", uploadResp.SecureURL)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}
