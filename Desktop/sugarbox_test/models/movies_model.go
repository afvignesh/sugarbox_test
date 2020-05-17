package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Comment struct{
	UserName string `json:"user_name"`
	Comment string `json:"comment"`
}

type Rating struct{
	UserName string `json:"user_name"`
	Rating int `json:"rating"`
}

type MovieInfo struct{
	ID primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"` 
	Name string `json:"name"`
	Comments []Comment `json:"comment"`
	Ratings []Rating `json:"ratings"`
	AvgRating float64 `json:"avg_rating,omitempty"`
	RatingCount int `json:"rating_count,omitempty"`
}

type UserActivity struct{
	UserName string `json:"user_name"`
	Movies []string `json:"movies"`
}

