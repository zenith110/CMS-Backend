package routes

import (
	"context"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt"
	log "github.com/sirupsen/logrus"
	"github.com/zenith110/CMS-Backend/graph/model"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

func hashAndSalt(password []byte) string {

	// Use GenerateFromPassword to hash & salt pwd.
	// MinCost is just an integer constant provided by the bcrypt
	// package along with DefaultCost & MaxCost.
	// The cost can be any value you want provided it isn't lower
	// than the MinCost (4)
	hash, err := bcrypt.GenerateFromPassword(password, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	} // GenerateFromPassword returns a byte slice so we need to
	// convert the bytes to a string and return it
	return string(hash)
}
func CreateUser(input *model.UserCreation) (*model.User, error) {
	client := ConnectToMongo()
	db := client.Database("blog").Collection("Users")
	var userlookup model.User
	email := input.Email
	// Looks up the user
	err := db.FindOne(context.TODO(), bson.M{"email": email}).Decode(&userlookup)
	if err != nil {
		var projects model.Projects
		password := input.Password
		hashedPassword := hashAndSalt([]byte(password))
		frontendUri := os.Getenv("CMSFRONTENDURI")
		user := model.User{Email: email, HashedPassword: hashedPassword, ProfilePicture: "", ProfileLInk: fmt.Sprintf("%s/%s", frontendUri, email), Role: "User", Projects: &projects}
		_, err := db.InsertOne(context.TODO(), user)
		if err != nil {
			fmt.Printf("error is %v", err)
		}
		return &user, err
	}
	var projects model.Projects
	return &model.User{Email: "", HashedPassword: "", Role: "", ProfilePicture: "", ProfileLInk: "", Projects: &projects}, err
}

func Authenticate(email string, password string, jwt string) model.User {
	if jwt == "" {
		panic("JWT is not valid!")
	}
	client := ConnectToMongo()
	collection := client.Database("blog").Collection("Users")
	var user model.User
	hashedPassword := hashAndSalt([]byte(password))
	//Passing the bson.D{{}} as the filter matches documents in the collection
	err := collection.FindOne(context.TODO(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		log.Fatal(err)
	}
	if user.HashedPassword == hashedPassword {
		return user
	}
	return user
}
func Login() string {
	var sampleSecretKey = []byte(os.Getenv("SECRETKEY"))
	token := jwt.New(jwt.SigningMethodEdDSA)
	tokenString, err := token.SignedString(sampleSecretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
