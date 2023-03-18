package routes

import (
	"context"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/zenith110/CMS-Backend/graph/model"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

func hashAndSalt(password []byte) string {
	hash, err := bcrypt.GenerateFromPassword(password, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}
	return string(hash)
}

func comparePasswords(hashedPwd string, plainPwd []byte) bool {
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPwd)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func CreateDefaultAdmin() {
	client := ConnectToMongo()
	adminuuid := uuid.New()
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
		user := model.User{Email: email, HashedPassword: hashedPassword, ProfilePicture: "", ProfileLink: fmt.Sprintf("%s/%s", frontendUri, email), Role: "Admin", Projects: &projects, Username: username, UUID: adminuuid.String()}
		_, dbroleInserterr := dbrole.InsertOne(context.TODO(), user)
		_, dbnormalInserterr := dbnormal.InsertOne(context.TODO(), user)
		if dbroleInserterr != nil || dbnormalInserterr != nil {
			fmt.Printf("error is %v", dbnormalInserterr)
		}
		CreateZincUser(username, password, email)
	} else {
		fmt.Printf("%s has an account associated already!!\n", os.Getenv("ADMINUSER"))
	}
}
func CreateUser(input *model.UserCreation) (*model.User, error) {
	redisClient := RedisClientInstation()
	redisData := RedisUserInfo(input.Jwt, redisClient)
	if input.Jwt == "" || redisData["role"] != "Admin" {
		panic("Error occured. Credentials are not valid!")
	}
	username := input.Username
	client := ConnectToMongo()
	dbrole := client.Database("users").Collection(*&input.Role)
	dbnormal := client.Database("blog").Collection("Users")
	var roleUserlookup model.User
	var userlookup model.User
	email := input.Email
	// Looks up the user
	dbroleerr := dbrole.FindOne(context.TODO(), bson.M{"username": username}).Decode(&roleUserlookup)
	dbnormalerr := dbnormal.FindOne(context.TODO(), bson.M{"username": username}).Decode(&userlookup)
	if dbroleerr != nil || dbnormalerr != nil {
		var projects model.Projects
		password := input.Password
		hashedPassword := hashAndSalt([]byte(password))
		frontendUri := os.Getenv("CMSFRONTENDURI")
		profilePic := UploadAvatarImageCreation(input)
		user := model.User{Email: email, HashedPassword: hashedPassword, ProfilePicture: profilePic, ProfileLink: fmt.Sprintf("%s/%s", frontendUri, username), Role: input.Role, Projects: &projects, Username: username, UUID: input.UUID}
		_, dbroleInserterr := dbrole.InsertOne(context.TODO(), user)
		_, dbnormalInserterr := dbnormal.InsertOne(context.TODO(), user)
		if dbroleInserterr != nil || dbnormalInserterr != nil {
			fmt.Printf("error is %v", dbnormalInserterr)
		}
		defer CloseClientDB()
		return &user, dbnormalInserterr
	}
	var projects model.Projects
	return &model.User{Email: "", HashedPassword: "", Role: "", ProfilePicture: "", ProfileLink: "", Projects: &projects, Username: ""}, nil
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

	//Passing the bson.D{{}} as the filter matches documents in the collection
	err := collection.FindOne(context.TODO(), bson.M{"username": username, role: role}).Decode(&user)
	if err != nil {
		log.Fatal(err)
	}
	passwordCheck := comparePasswords(user.HashedPassword, []byte(password))
	if passwordCheck == true {
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

	//Passing the bson.D{{}} as the filter matches documents in the collection
	err := collection.FindOne(context.TODO(), bson.M{"username": username, "role": "Reader"}).Decode(&user)
	if err != nil {
		log.Fatal(err)
	}
	passwordCheck := comparePasswords(user.HashedPassword, []byte(password))
	if passwordCheck == true {
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
func Login(username string, password string) (*model.LoginData, error) {
	var sampleSecretKey = []byte(os.Getenv("SECRETKEY"))
	token := jwt.New(jwt.SigningMethodHS512)
	tokenString, err := token.SignedString(sampleSecretKey)
	if err != nil {
		log.Error(fmt.Sprintf("%v", err))
		panic(fmt.Sprintf("%v", err))
	}
	redisLoginClient := RedisClientInstation()
	client := ConnectToMongo()
	collection := client.Database("blog").Collection("Users")
	var user model.User
	findErr := collection.FindOne(context.TODO(), bson.M{"username": username}).Decode(&user)

	if findErr != nil {
		fmt.Printf("%v", err)
		log.Fatal(err)
	}
	passwordCheck := comparePasswords(user.HashedPassword, []byte(password))

	if passwordCheck == true {
		userInfo := map[string]string{
			"username": user.Username,
			"password": password,
			"role":     user.Role,
		}
		userData, marshalErr := MarshalBinary(userInfo)
		if marshalErr != nil {
			panic(fmt.Sprintf("error while marshling user data is: %v", err))
		}
		err := redisLoginClient.Set(tokenString, userData, 0).Err()
		if err != nil {
			panic(fmt.Sprintf("error is %v", err))
		}
		redisClient := RedisClientInstation()
		redisData := RedisUserInfo(tokenString, redisClient)
		loginData := model.LoginData{Jwt: tokenString, Role: redisData["role"], Username: redisData["username"]}
		defer CloseClientDB()
		return &loginData, nil
	}
	var loginData model.LoginData
	defer CloseClientDB()
	return &loginData, nil
}

/*
@param - jwtToken
@type - string

@rtype - string, err
@description - Validates the JWT is properly signed within the CMS
*/
func JWTValidityCheck(jwtToken string) (string, error) {
	token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			panic(fmt.Sprint("An error has occured in signing!"))
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

func Logout(jwt string) (string, error) {
	redisClient := RedisClientInstation()
	_, err := redisClient.Del(jwt).Result()
	if err != nil {
		panic(fmt.Sprintf("Error while logging out! %v\n", err))
	}
	return "", err
}
