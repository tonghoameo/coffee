package controller

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/binbomb/storeCoffee/database"
	helper "github.com/binbomb/storeCoffee/helpers"
	"github.com/binbomb/storeCoffee/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		// c.Query = recordPerPage=2&page=2
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}
		page, err := strconv.Atoi(c.Query("page"))
		if err != nil || page < 1 {
			page = 1
		}
		startIdx := (page - 1) * recordPerPage
		startIdx, err = strconv.Atoi(c.Query("startIdx"))
		matchStage := bson.D{{"$match", bson.D{}}}
		projectStage := bson.D{
			{
				"$project", bson.D{
					{"_id", 0},
					{"total_Count", 1},
					{"user_items", bson.D{{"$slice", []interface{}{"$data", startIdx, recordPerPage}}}},
				}}}

		rs, err := userCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage, projectStage,
		})
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error when aggrate users"})
		}
		count, err := userCollection.CountDocuments(ctx, bson.M{})
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		fmt.Printf("number user in database %d \n", count)
		var allUsers []bson.M
		if err = rs.All(ctx, &allUsers); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allUsers)

	}
}
func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		userId := c.Param("user_id")
		var user models.User
		var err = userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error orcued while fetcing data find user_id"})
		}
		c.JSON(http.StatusOK, user)

	}
}
func Signup() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		fmt.Println("here")
		validatorErr := validate.Struct(user)
		if validatorErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validatorErr.Error()})
			return
		}
		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		hashedPassword := HashPassword(*user.Password)
		user.Password = &hashedPassword
		count, err = userCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if count > 1 {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "email or phone is used"})
			return
		}
		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()

		token, refreshToken, _ := helper.GeneralAllToken(*user.Email, *user.First_name, *user.Last_name, user.User_id)

		user.Token = &token
		user.Refresh_Token = &refreshToken
		rs, err := userCollection.InsertOne(ctx, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, rs)
	}
}
func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		var foundUser models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		fmt.Printf("%s \n %s \n", *user.Password, *foundUser.Password)
		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)

		defer cancel()
		if passwordIsValid == false {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		token, refreshToken, _ := helper.GeneralAllToken(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, foundUser.User_id)
		helper.UpdateAllTokens(token, refreshToken, foundUser.User_id)

		c.JSON(http.StatusOK, foundUser)
	}
}

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Fatal(err)
	}

	return string(bytes)
}

// password like hashedpassword, and providePassword like raw text
func VerifyPassword(password string, hashedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	check := true
	msg := "success"
	if err != nil {
		check = false
		msg = "passwod wrong"
	}
	return check, msg
}
