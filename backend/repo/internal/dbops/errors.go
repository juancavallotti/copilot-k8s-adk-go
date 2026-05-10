package dbops

import "errors"

// ErrRecipeNotFound is returned when no recipe exists for the given id.
var ErrRecipeNotFound = errors.New("dbops: recipe not found")

// ErrInvalidID is returned when an id is not a valid UUID.
var ErrInvalidID = errors.New("dbops: invalid recipe id")

// ErrParseIngredient is returned when an ingredient line cannot be parsed.
var ErrParseIngredient = errors.New("dbops: ingredient parse error")
