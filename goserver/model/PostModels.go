package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Post struct {
	ID     primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	Name   string             `json:"name,omitempty"`
	Prompt string             `json:"prompt,omitempty"`
	Photo  string             `json:"photo,omitempty"`
}
