// Code generated by github.com/spachava753/fibergql, DO NOT EDIT.

package model

import (
	"github.com/spachava753/fibergql/graphql"
)

// The `File` type, represents the response of uploading a file.
type File struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Content     string `json:"content"`
	ContentType string `json:"contentType"`
}

// The `UploadFile` type, represents the request for uploading a file with certain payload.
type UploadFile struct {
	ID   int            `json:"id"`
	File graphql.Upload `json:"file"`
}
