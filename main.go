/*

	auth0 tenant domain -> dev-mj2rvpog.auth0.com
	API name -> go-react-test-api
	identifier -> https://goreacttestapi.com

*/

package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "ares"
	dbname   = "Jokes"
)

type Response struct {
	Message string `json:"message"`
}

type Jwks struct {
	Keys []JSONWebKeys `json:"keys"`
}

type JSONWebKeys struct {
	Kty string   `json:"keys"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

type Joke struct {
	ID   int    `json:"id"`
	Joke string `json:"joke"`
}

var jwtMiddleWare *jwtmiddleware.JWTMiddleware

var db *sql.DB

func init() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	tmpDB, err := sql.Open("postgres", psqlInfo)

	if err != nil {
		log.Fatal(err)
	}
	db = tmpDB
}

func main() {

	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			aud := os.Getenv("AUTH0_API_AUDIENCE")
			checkAudience := token.Claims.(jwt.MapClaims).VerifyAudience(aud, false)

			if !checkAudience {
				return token, errors.New("Invalid audience.")
			}

			// verify iss claim
			iss := os.Getenv("AUTH0_DOMAIN")
			checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(iss, false)
			if !checkIss {
				return token, errors.New("Invalid user.")
			}

			cert, err := getPemCert(token)
			if err != nil {
				fmt.Printf("Could not get cert: %+v", err)
			}

			result, _ := jwt.ParseECPublicKeyFromPEM([]byte(cert))
			return result, nil
		},
		SigningMethod: jwt.SigningMethodRS256,
	})

	jwtMiddleWare = jwtMiddleware

	router := gin.Default()
	router.Use(cors.Default())
	router.Use(static.Serve("/", static.LocalFile("./web", true)))

	api := router.Group("/api")

	api.GET("/jokes", GetJokes)
	api.POST("/joke", CreateJoke)
	api.PUT("/joke", UpdateJoke)
	api.DELETE("/joke", DeleteJoke)

	router.Run(":8080")
}

func getPemCert(token *jwt.Token) (string, error) {
	cert := ""
	resp, err := http.Get(os.Getenv("AUTH0_DOMAIN") + ".well-known/jwts.json")

	if err != nil {
		return cert, err
	}

	defer resp.Body.Close()

	var jwks = Jwks{}
	err = json.NewDecoder(resp.Body).Decode(&jwks)

	if err != nil {
		return cert, err
	}

	x5c := jwks.Keys[0].X5c
	for k, v := range x5c {
		if token.Header["kid"] == jwks.Keys[k].Kid {
			cert = "-----BEGIN CERTIFICATE-----\n" + v + "\n-----END CERTIFICATE-----"
		}
	}

	if cert == "" {
		return cert, errors.New("unable to find appropriate key")
	}

	return cert, nil
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println(c.Request)
		err := jwtMiddleWare.CheckJWT(c.Writer, c.Request)
		if err != nil {
			fmt.Println(err)
			c.Abort()
			c.Writer.WriteHeader(http.StatusUnauthorized)
			c.Writer.Write([]byte("Unauthorized"))
			return
		}
	}
}

func GetJokes(c *gin.Context) {
	jokeArray, err := allJokes()

	if err != nil {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, err)
	} else {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, jokeArray)
	}

	return
}

func CreateJoke(c *gin.Context) {
	var jokeData Joke
	err := c.BindJSON(&jokeData)

	if err != nil {
		log.Fatal(err)
	}

	jokeID, err := createJoke(jokeData.Joke)

	_ = jokeID

	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
	} else {
		jokes, err1 := allJokes()

		if err1 != nil {
			c.AbortWithStatus(http.StatusNotFound)
		} else {
			c.JSON(http.StatusOK, &jokes)
		}
	}
}

func UpdateJoke(c *gin.Context) {
	var jokeData Joke
	err := c.BindJSON(&jokeData)

	if err != nil {
		log.Fatal(err)
	}

	updatedID, err := updateJoke(jokeData.ID, jokeData.Joke)

	_ = updatedID

	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
	} else {
		jokes, err1 := allJokes()

		if err1 != nil {
			c.AbortWithStatus(http.StatusNotFound)
		} else {
			c.JSON(http.StatusOK, &jokes)
		}
	}
}

func DeleteJoke(c *gin.Context) {
	var jokeData Joke
	err := c.BindJSON(&jokeData)

	if err != nil {
		log.Fatal(err)
	}

	deleteID, err := deleteJoke(jokeData.ID)

	_ = deleteID

	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
	} else {
		jokes, err1 := allJokes()

		if err1 != nil {
			c.AbortWithStatus(http.StatusNotFound)
		} else {
			c.JSON(http.StatusOK, &jokes)
		}
	}
}
