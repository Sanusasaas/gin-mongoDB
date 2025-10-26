package helpers

import (
	"errors"
	"log"
	"github.com/gin-gonic/gin"
)

func CheckUserType(c *gin.Context, role string) (err error) {
	userType := c.GetString("user_type")
	err = nil
	if userType != role {
		err = errors.New("unathorized to access this resource")
		return err
	}
	return err
}

func MatchUserTypeToUid(c *gin.Context, userId string) (err error) {
    userType := c.GetString("user_type")
    uid := c.GetString("uid")
	log.Printf("uid: %s, userId: %s, userType: %s", uid, userId, userType)

    if userType != "ADMIN" && uid != userId {
        return errors.New("unauthorized to access this resource")
    }

    return CheckUserType(c, userType)
}