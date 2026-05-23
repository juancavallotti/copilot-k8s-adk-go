package recipes

import "errors"

// ErrRecipeNotFound is returned when no recipe exists for the given id.
var ErrRecipeNotFound = errors.New("dbops: recipe not found")

// ErrPhotoNotFound is returned when no photo link exists for the given recipe/photo ids.
var ErrPhotoNotFound = errors.New("dbops: recipe photo not found")

// ErrInvalidID is returned when an id is not a valid UUID.
var ErrInvalidID = errors.New("dbops: invalid recipe id")

// ErrParseIngredient is returned when an ingredient line cannot be parsed.
var ErrParseIngredient = errors.New("dbops: ingredient parse error")
