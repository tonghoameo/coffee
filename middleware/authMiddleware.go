package middleware

import (
	"fmt"
	"net/http"

	helper "github.com/binbomb/storeCoffee/helpers"
	"github.com/gin-gonic/gin"
)

func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientToken := c.Request.Header.Get("token")
		if clientToken == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "authien cate err"})
			c.Abort()
			return
		}
		//claim is type SignedDetails
		claim, msg := helper.ValidToken(clientToken)
		if msg != "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			c.Abort()
			return
		}
		fmt.Println(claim)
		c.Set("email", claim.Email)
		c.Set("first_name", claim.First_name)
		c.Set("last_name", claim.Last_name)
		c.Set("uid", claim.Uid)
		c.Next()
	}
}
