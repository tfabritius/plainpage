package model

import "errors"

var ErrNotFound = errors.New("not found")
var ErrParentFolderNotFound = errors.New("parent folder not found")
var ErrPageOrFolderExistsAlready = errors.New("page or folder exists already")
var ErrFolderNotEmpty = errors.New("folder is not empty")
var ErrInvalidUsername = errors.New("invalid username")
var ErrUserExistsAlready = errors.New("user already exists")
