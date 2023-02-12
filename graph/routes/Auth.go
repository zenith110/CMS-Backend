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
func CreateDefaultAdmin() {
	client := ConnectToMongo()
	dbrole := client.Database("users").Collection("Admin")
	dbnormal := client.Database("blog").Collection("Users")
	var roleUserlookup model.User
	var userlookup model.User
	email := os.Getenv("ADMINEMAIL")
	// Looks up the user
	dbroleerr := dbrole.FindOne(context.TODO(), bson.M{"email": email}).Decode(&roleUserlookup)
	dbnormalerr := dbrole.FindOne(context.TODO(), bson.M{"email": email, "role": "Admin"}).Decode(&userlookup)
	if dbroleerr != nil && dbnormalerr != nil {
		var projects model.Projects
		password := os.Getenv("ADMINPASSWORD")
		hashedPassword := hashAndSalt([]byte(password))
		frontendUri := os.Getenv("CMSFRONTENDURI")
		user := model.User{Email: email, HashedPassword: hashedPassword, ProfilePicture: "", ProfileLInk: fmt.Sprintf("%s/%s", frontendUri, email), Role: "Admin", Projects: &projects}
		_, dbroleInserterr := dbrole.InsertOne(context.TODO(), user)
		_, dbnormalInserterr := dbnormal.InsertOne(context.TODO(), user)
		if dbroleInserterr != nil || dbnormalInserterr != nil {
			fmt.Printf("error is %v", dbnormalInserterr)
		}
	} else {
		fmt.Printf("%s has an account associated already!!", email)
	}
}
func CreateUser(input *model.UserCreation) (*model.User, error) {
	if input.Jwt == "" {
		panic("JWT is invalid!")
	}
	client := ConnectToMongo()
	dbrole := client.Database("users").Collection(*&input.Role)
	dbnormal := client.Database("blog").Collection("Users")
	var roleUserlookup model.User
	var userlookup model.User
	email := input.Email
	// Looks up the user
	dbroleerr := dbrole.FindOne(context.TODO(), bson.M{"email": email}).Decode(&roleUserlookup)
	dbnormalerr := dbrole.FindOne(context.TODO(), bson.M{"email": email}).Decode(&userlookup)
	if dbroleerr != nil && dbnormalerr != nil {
		var projects model.Projects
		password := input.Password
		hashedPassword := hashAndSalt([]byte(password))
		frontendUri := os.Getenv("CMSFRONTENDURI")
		user := model.User{Email: email, HashedPassword: hashedPassword, ProfilePicture: "", ProfileLInk: fmt.Sprintf("%s/%s", frontendUri, email), Role: input.Role, Projects: &projects}
		_, dbroleInserterr := dbrole.InsertOne(context.TODO(), user)
		_, dbnormalInserterr := dbnormal.InsertOne(context.TODO(), user)
		if dbroleInserterr != nil || dbnormalInserterr != nil {
			fmt.Printf("error is %v", dbnormalInserterr)
		}
		defer CloseClientDB()
		return &user, dbnormalInserterr
	}
	var projects model.Projects
	return &model.User{Email: "", HashedPassword: "", Role: "", ProfilePicture: "", ProfileLInk: "", Projects: &projects}, nil
}

func AuthenticateNonReaders(email string, password string, jwt string, role string) model.User {
	if jwt == "" {
		panic("JWT is not valid!")
	}
	client := ConnectToMongo()
	collection := client.Database("users").Collection(role)
	var user model.User
	hashedPassword := hashAndSalt([]byte(password))
	//Passing the bson.D{{}} as the filter matches documents in the collection
	err := collection.FindOne(context.TODO(), bson.M{"email": email, role: role}).Decode(&user)
	if err != nil {
		log.Fatal(err)
	}
	if user.HashedPassword == hashedPassword {
		return user
	}
	defer CloseClientDB()
	return user
}
func AuthenticateReaders(email string, password string) model.User {
	client := ConnectToMongo()
	collection := client.Database("user").Collection("Reader")
	var user model.User
	hashedPassword := hashAndSalt([]byte(password))
	//Passing the bson.D{{}} as the filter matches documents in the collection
	err := collection.FindOne(context.TODO(), bson.M{"email": email, "role": "Reader"}).Decode(&user)
	if err != nil {
		log.Fatal(err)
	}
	if user.HashedPassword == hashedPassword {
		return user
	}
	defer CloseClientDB()
	return user
}
func Login(email string, password string) (string, error) {
	var sampleSecretKey = []byte(os.Getenv("SECRETKEY"))
	token := jwt.New(jwt.SigningMethodEdDSA)
	tokenString, err := token.SignedString(sampleSecretKey)
	if err != nil {
		return "", err
	}
	client := ConnectToMongo()
	collection := client.Database("blog").Collection("Users")
	var user model.User
	hashedPassword := hashAndSalt([]byte(password))
	//Passing the bson.D{{}} as the filter matches documents in the collection
	findErr := collection.FindOne(context.TODO(), bson.M{"email": email}).Decode(&user)
	if findErr != nil {
		log.Fatal(err)
	}
	if user.HashedPassword == hashedPassword {
		defer CloseClientDB()
		return tokenString, nil
	}
	defer CloseClientDB()
	return tokenString, nil
}
func JWTValidityCheck(jwtToken string) (string, error) {
	token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodECDSA)
		if !ok {
			return "Unauthroized!", nil
		}
		return "", nil
	})
	if err != nil {
		return "Unauthroized!", nil
	}
	if token.Valid {
		return "Authorized", nil
	} else {
		return "Unauthorized!", nil
	}
}
