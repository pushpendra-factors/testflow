package model

import (
	"time"
)

// A model to store the following data:
// {"aud":"tyttxndI2AVVHGt9AKJ24TRzFAbhwxj1","email":"muktesharyan@factors.ai","email_verified":true,"exp":1644169122,"family_name":"Uppala","given_name":"Muktesharyan","iat":1644133122,"iss":"https://dev-ustbxpch.us.auth0.com/","locale":"en","name":"Muktesharyan Uppala","nickname":"muktesharyan","picture":"https://lh3.googleusercontent.com/a/AATXAJydMJS90JkbsYo6GoJnqVkFkBAEYqrInauifen-=s96-c","sub":"google-oauth2|105189179907788069961","updated_at":"2022-02-06T07:38:40.292Z"}

type Auth0Profile struct {
	Email           string    `json:"email"`
	IsEmailVerified bool      `json:"email_verified"`
	FirstName       string    `json:"given_name"`
	LastName        string    `json:"family_name"`
	Subject         string    `json:"sub"`
	IssuedAt        uint64    `json:"iat"`
	ExpiresAt       uint64    `json:"exp"`
	UpdatedAt       time.Time `json:"updated_at"`
}
