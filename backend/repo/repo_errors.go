package repo

import (
	"juancavallotti.com/recipes-repo/internal/dbops"
	"juancavallotti.com/recipes-repo/internal/service"
)

// Sentinel errors re-exported for API layers outside internal/.
var (
	ErrRecipeNotFound  = dbops.ErrRecipeNotFound
	ErrPhotoNotFound   = dbops.ErrPhotoNotFound
	ErrInvalidID       = dbops.ErrInvalidID
	ErrParseIngredient = dbops.ErrParseIngredient
	ErrInvalidRecipe   = service.ErrInvalidRecipe
	ErrInvalidRecipeID = service.ErrInvalidRecipeID

	ErrEventNotFound = dbops.ErrEventNotFound

	ErrSkillNotFound           = dbops.ErrSkillNotFound
	ErrSkillNameTaken          = dbops.ErrSkillNameTaken
	ErrInvalidSkillID          = service.ErrInvalidSkillID
	ErrInvalidSkillName        = service.ErrInvalidSkillName
	ErrInvalidSkillDescription = service.ErrInvalidSkillDescription
	ErrInvalidSkillContent     = service.ErrInvalidSkillContent
)
