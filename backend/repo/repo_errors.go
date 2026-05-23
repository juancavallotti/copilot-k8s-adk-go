package repo

import (
	recipeops "juancavallotti.com/recipes-repo/internal/dbops/recipes"
	skillops "juancavallotti.com/recipes-repo/internal/dbops/skills"
	traceops "juancavallotti.com/recipes-repo/internal/dbops/traces"
	recipesvc "juancavallotti.com/recipes-repo/internal/service/recipes"
	skillsvc "juancavallotti.com/recipes-repo/internal/service/skills"
)

// Sentinel errors re-exported for API layers outside internal/.
var (
	ErrRecipeNotFound  = recipeops.ErrRecipeNotFound
	ErrPhotoNotFound   = recipeops.ErrPhotoNotFound
	ErrInvalidID       = recipeops.ErrInvalidID
	ErrParseIngredient = recipeops.ErrParseIngredient
	ErrInvalidRecipe   = recipesvc.ErrInvalidRecipe
	ErrInvalidRecipeID = recipesvc.ErrInvalidRecipeID

	ErrEventNotFound = traceops.ErrEventNotFound

	ErrSkillNotFound           = skillops.ErrSkillNotFound
	ErrSkillNameTaken          = skillops.ErrSkillNameTaken
	ErrInvalidSkillID          = skillsvc.ErrInvalidSkillID
	ErrInvalidSkillName        = skillsvc.ErrInvalidSkillName
	ErrInvalidSkillDescription = skillsvc.ErrInvalidSkillDescription
	ErrInvalidSkillContent     = skillsvc.ErrInvalidSkillContent
)
