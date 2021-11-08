package auth

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/twinj/uuid"
	"golang.org/x/crypto/bcrypt"
)

/*
 * PASSWORD HASHING
 */

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

/*
 * TOKEN MEMORY MANAGEMENT
 */

// Look for tokens that have expired and remove from map
func CleanTokens() {
	for {
		time.Sleep(time.Minute * 5)
		for k, v := range ActiveTokens {
			if time.Now().After(v.Timeout) {
				fmt.Printf("Invalidating id: %d\n", v.Id)
				delete(ActiveTokens, k)
			}
		}
	}
}

// Store tokens in memory and attach a timer
func CreateAuth(id uint32, td *TokenDetails) {
	at := time.Unix(td.AtExpires, 0)
	rt := time.Unix(td.RtExpires, 0)

	ActiveTokens[td.AccessUuid] = ActiveToken{id, at}
	ActiveTokens[td.RefreshUuid] = ActiveToken{id, rt}
	fmt.Printf("Assigned new token: %s to user %d\n", td.AccessToken, id)
}

// Remove token from memory
func DeleteAuth(uuid string) (uint32, error) {
	v, ok := ActiveTokens[uuid]
	if !ok {
		return 0, fmt.Errorf("unable to find uuid in active tokens")
	}
	delete(ActiveTokens, uuid)
	return v.Id, nil
}

/*
 * TOKEN CREATE AND DELETE
 */

// Create Access and Refresh tokens
func CreateToken(userid uint32) (*TokenDetails, error) {
	var err error
	td := &TokenDetails{}
	td.AtExpires = time.Now().Add(time.Minute * 15).Unix()
	td.AccessUuid = uuid.NewV4().String()

	td.RtExpires = time.Now().Add(time.Hour * 24 * 7).Unix()
	td.RefreshUuid = uuid.NewV4().String()

	// Create Access Token
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["access_uuid"] = td.AccessUuid
	atClaims["user_id"] = userid
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(ACCESS_SECRET))
	if err != nil {
		return nil, err
	}

	//Create Refresh Token
	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = td.RefreshUuid
	rtClaims["user_id"] = userid
	rtClaims["exp"] = td.RtExpires
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(REFRESH_SECRET))
	if err != nil {
		return nil, err
	}

	return td, nil
}

/*
 * TOKEN VALIDATION AND EXTRACTION
 */

// Get token string from request header
func ExtractToken(r *http.Request) string {
	bearToken := r.Header.Get("Authorization")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

// Ensure Token is signed correctly
func VerifyToken(r *http.Request) (*jwt.Token, error) {
	tokenString := ExtractToken(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(ACCESS_SECRET), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

// Ensure Token is valid
func TokenValid(r *http.Request) error {
	token, err := VerifyToken(r)
	if err != nil {
		return err
	}
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		return err
	}
	return nil
}

// Extract UUID and ID from token
func ExtractTokenMetadata(r *http.Request) (*AccessDetails, error) {
	token, err := VerifyToken(r)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		accessUuid, ok := claims["access_uuid"].(string)
		if !ok {
			return nil, err
		}
		userId, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["user_id"]), 10, 32)
		if err != nil {
			return nil, err
		}
		return &AccessDetails{
			AccessUuid: accessUuid,
			UserId:     uint32(userId),
		}, nil
	}
	return nil, err
}

// Return ID based on UUID in ActiveTokens
func FetchAuth(authD *AccessDetails) (uint32, error) {
	token, ok := ActiveTokens[authD.AccessUuid]
	if !ok {
		return 0, fmt.Errorf("no matching token")
	}
	return token.Id, nil
}
