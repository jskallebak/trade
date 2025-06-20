package auth

import ()

func GenerateJWT(userID int32, email string) (string, error) {

	return "", nil
}

func ValidateJWT(tokenString string) (userID int32, email string, err error) {
	return -1, "", nil
}
