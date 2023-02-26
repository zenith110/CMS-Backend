package routes

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v4"
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
	username := os.Getenv("ADMINUSER")
	// Looks up the user
	dbroleerr := dbrole.FindOne(context.TODO(), bson.M{"username": username}).Decode(&roleUserlookup)
	dbnormalerr := dbrole.FindOne(context.TODO(), bson.M{"username": username, "role": "Admin"}).Decode(&userlookup)
	if dbroleerr != nil && dbnormalerr != nil {
		var projects model.Projects
		password := os.Getenv("ADMINPASSWORD")
		hashedPassword := hashAndSalt([]byte(password))
		frontendUri := os.Getenv("CMSFRONTENDURI")
		email := os.Getenv("ADMINEMAIL")
		user := model.User{Email: email, HashedPassword: hashedPassword, ProfilePicture: "", ProfileLInk: fmt.Sprintf("%s/%s", frontendUri, email), Role: "Admin", Projects: &projects, Username: username}
		_, dbroleInserterr := dbrole.InsertOne(context.TODO(), user)
		_, dbnormalInserterr := dbnormal.InsertOne(context.TODO(), user)
		if dbroleInserterr != nil || dbnormalInserterr != nil {
			fmt.Printf("error is %v", dbnormalInserterr)
		}
		zincUsername := os.Getenv("ZINC_FIRST_ADMIN_USER")
		zincPassword := os.Getenv("ZINC_FIRST_ADMIN_PASSWORD")
		zincBaseUrl := os.Getenv("ZINCBASE")
		zincData := fmt.Sprintf(`{
			"_id": "%s",
			"name": "%s",
			"role": "Admin",
			"password": "%s"
		}`, username, email, password)
		zincDocumentUrl := fmt.Sprintf("%s/api/user", zincBaseUrl)
		req, err := http.NewRequest("POST", zincDocumentUrl, strings.NewReader(zincData))
		if err != nil {
			log.Fatal(fmt.Errorf("error has occured when sending data! %v", err))
		}

		req.SetBasicAuth(zincUsername, zincPassword)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", os.Getenv("USERAGENT"))

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Fatal(fmt.Errorf("error has occured while grabbing data! %v", err))
		}
		defer resp.Body.Close()

		_, userCreationErr := io.ReadAll(resp.Body)
		if userCreationErr != nil {
			log.Fatal(fmt.Errorf("error occured while reading the data! %v", err))
		}

	} else {
		fmt.Printf("%s has an account associated already!!\n", os.Getenv("ADMINUSER"))
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
	dbroleerr := dbrole.FindOne(context.TODO(), bson.M{"username": input.Username}).Decode(&roleUserlookup)
	dbnormalerr := dbrole.FindOne(context.TODO(), bson.M{"username": input.Username}).Decode(&userlookup)
	if dbroleerr != nil && dbnormalerr != nil {
		var projects model.Projects
		password := input.Password
		hashedPassword := hashAndSalt([]byte(password))
		frontendUri := os.Getenv("CMSFRONTENDURI")
		user := model.User{Email: email, HashedPassword: hashedPassword, ProfilePicture: "", ProfileLInk: fmt.Sprintf("%s/%s", frontendUri, email), Role: input.Role, Projects: &projects, Username: input.Username}
		_, dbroleInserterr := dbrole.InsertOne(context.TODO(), user)
		_, dbnormalInserterr := dbnormal.InsertOne(context.TODO(), user)
		if dbroleInserterr != nil || dbnormalInserterr != nil {
			fmt.Printf("error is %v", dbnormalInserterr)
		}
		defer CloseClientDB()
		zincUsername := os.Getenv("ZINC_FIRST_ADMIN_USER")
		zincPassword := os.Getenv("ZINC_FIRST_ADMIN_PASSWORD")
		zincBaseUrl := os.Getenv("ZINCBASE")
		zincData := fmt.Sprintf(`{
			"_id": "%s",
			"name": "%s",
			"role": "Admin",
			"password": "%s"
		}`, input.Username, input.Email, input.Password)
		zincDocumentUrl := fmt.Sprintf("%s/api/user", zincBaseUrl)
		req, err := http.NewRequest("POST", zincDocumentUrl, strings.NewReader(zincData))
		if err != nil {
			log.Fatal(fmt.Errorf("error has occured when sending data! %v", err))
		}

		req.SetBasicAuth(zincUsername, zincPassword)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", os.Getenv("USERAGENT"))

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Fatal(fmt.Errorf("error has occured while grabbing data! %v", err))
		}
		defer resp.Body.Close()

		_, userCreationErr := io.ReadAll(resp.Body)
		if userCreationErr != nil {
			log.Fatal(fmt.Errorf("error occured while reading the data! %v", err))
		}
		return &user, dbnormalInserterr
	}
	var projects model.Projects
	return &model.User{Email: "", HashedPassword: "", Role: "", ProfilePicture: "", ProfileLInk: "", Projects: &projects, Username: ""}, nil
}

/*
@param - username
@type - string

@param - password
@type - string

@param - role
@type - string
@description - Role a user has

@rtype - model.User
@description - Authenticates the user based off a non Reader role.
*/
func AuthenticateNonReaders(username string, password string, jwt string, role string) model.User {
	if jwt == "" {
		panic("JWT is not valid!")
	}
	client := ConnectToMongo()
	collection := client.Database("users").Collection(role)
	var user model.User
	hashedPassword := hashAndSalt([]byte(password))
	//Passing the bson.D{{}} as the filter matches documents in the collection
	err := collection.FindOne(context.TODO(), bson.M{"username": username, role: role}).Decode(&user)
	if err != nil {
		log.Fatal(err)
	}
	if user.HashedPassword == hashedPassword {
		return user
	}
	defer CloseClientDB()
	return user
}

/*
@param - email
@type - string

@param - password
@type - string
@rtype - model.User
@description - Authenticates only Reader users(examples include public facing sites)
*/
func AuthenticateReaders(username string, password string) model.User {
	client := ConnectToMongo()
	collection := client.Database("user").Collection("Reader")
	var user model.User
	hashedPassword := hashAndSalt([]byte(password))
	//Passing the bson.D{{}} as the filter matches documents in the collection
	err := collection.FindOne(context.TODO(), bson.M{"username": username, "role": "Reader"}).Decode(&user)
	if err != nil {
		log.Fatal(err)
	}
	if user.HashedPassword == hashedPassword {
		return user
	}
	defer CloseClientDB()
	return user
}

/*
@param - email
@type - string

@param - password
@type - string
@rtype - string, err
@description - Authenticates the user, and returns a JWT.
*/
func Login(username string, password string) (string, error) {
	var sampleSecretKey = []byte(os.Getenv("SECRETKEY"))
	token := jwt.New(jwt.SigningMethodHS512)
	tokenString, err := token.SignedString(sampleSecretKey)
	if err != nil {
		log.Error(fmt.Sprintf("%v", err))
		return "", err
	}
	client := ConnectToMongo()
	collection := client.Database("blog").Collection("Users")
	var user model.User
	hashedPassword := hashAndSalt([]byte(password))
	findErr := collection.FindOne(context.TODO(), bson.M{"username": username}).Decode(&user)

	if findErr != nil {
		fmt.Printf("%v", err)
		log.Fatal(err)
	}
	if user.HashedPassword == hashedPassword {
		log.Error("User was found!!!\n")
		defer CloseClientDB()
		return tokenString, nil
	}
	defer CloseClientDB()
	return tokenString, nil
}

/*
@param - jwtToken
@type - string

@rtype - string, err
@description - Validates the JWT is properly signed.
*/
func JWTValidityCheck(jwtToken string) (string, error) {
	token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
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
