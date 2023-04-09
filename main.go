package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// func postAlbums(c *gin.Context) {
// 	var newAlbum album

// 	// Call BindJSON to bind the received JSON to
// 	// newAlbum.
// 	if err := c.BindJSON(&newAlbum); err != nil {
// 		return
// 	}

// 	// Add the new album to the slice.
// 	albums = append(albums, newAlbum)
// 	c.IndentedJSON(http.StatusCreated, newAlbum)
// }

// func getAlbumByID(c *gin.Context) {
// 	id := c.Param("id")

// 	for _, a := range albums {
// 		if a.ID == id {
// 			c.IndentedJSON(http.StatusOK, a)
// 			return
// 		}
// 	}

// 	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
// }

const baseURL = string("https://api.themoviedb.org/3/") // Note: needs trailing slash
var apiKey string                                       // API v3 auth key
const logName = "movigo.log"

type ContentType int

const (
	Movie ContentType = iota + 1
	TVShow
	Person
)

var ContentTypeMap = map[string]ContentType{
	"movie":  Movie,
	"tv":     TVShow,
	"person": Person,
}

// Get the string representation of the content type
func (contentType ContentType) String() string {
	names := [...]string{
		"movie",
		"tv",
		"person",
	}

	if contentType < Movie || contentType > Person {
		return "Unknown"
	}

	return names[contentType-1]
}

func (contentType ContentType) Validate() bool {
	return contentType >= Movie && contentType <= Person
}

func ParseContentType(content string) (ContentType, error) {
	contentType := ContentTypeMap[content]
	if !contentType.Validate() {
		return 0, errors.New("invalid content type")
	}

	return ContentTypeMap[content], nil
}

type MediaRequest struct {
	ID          int
	ContentType ContentType
}

func NewMediaRequest(id int, contentType ContentType) *MediaRequest {
	// Can't have negative IDs
	if id <= 0 {
		return nil
	}

	return &MediaRequest{
		ID:          id,
		ContentType: contentType,
	}
}

func getContentByType(c *gin.Context) {
	contentType, err := ParseContentType(c.Param("contentType"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid id"})
		return
	}

	fmt.Println(contentType, id)

	c.IndentedJSON(http.StatusOK, MediaRequest{ID: id, ContentType: contentType})
}

var logger *os.File

func main() {
	var err error
	logger = initLogger(logName) // Initialize the log file

	// Set the API key
	apiKey = os.Getenv("TMDB_API_V3_KEY")
	if apiKey == "" {
		log.Fatal("TMDB_API_V3_KEY is unavailable in the current environment")
	}

	db, err = initDB("./client-access.db") // Initialize the database connection
	defer db.Close()                       // Close the database connection
	defer logger.Close()                   // Close the log file

	testToken := GenerateToken()
	err = testToken.Write(db)
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()

	// Set up CORS middleware
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://127.0.0.1:3000"}
	router.Use(cors.New(config))
	// router.Use(cors.Default())

	router.GET("/:contentType/:id", getContentDetails)
	// router.GET("/albums/:id", getAlbumByID)
	// router.POST("/albums", postAlbums)

	// mediaReq := MediaRequest{ID: 95557, ContentType: TVShow}

	// getTMDB(&mediaReq) // 95557,tv: Invincible

	router.Run("localhost:8080")
}

func initLogger(logName string) *os.File {
	f, err := os.OpenFile(logName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(f)

	return f
}

func getTMDB(mediaReq *MediaRequest) {

	reqURL, err := url.Parse(baseURL)
	if err != nil {
		fmt.Println("Malformed URL: ", err.Error())
		return
	}

	reqURL.Path += mediaReq.ContentType.String() + "/" + strconv.Itoa(mediaReq.ID)
	params := url.Values{}
	params.Add("api_key", apiKey)
	params.Add("language", "en-US")

	reqURL.RawQuery = params.Encode()

	fmt.Println("Request URL: ", reqURL.String())

	res, err := http.DefaultClient.Get(reqURL.String())
	if err != nil {
		fmt.Println("Error: ", err.Error())
		return
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error: ", err.Error())
		return
	}

	var payload TV
	json.Unmarshal(body, &payload)

	fmt.Println(payload.Name, payload.ProductionCompanies[0].Name)
}
