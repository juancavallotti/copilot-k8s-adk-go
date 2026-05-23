package commands

import types "juancavallotti.com/recipe-types"

// recipePatch is the CLI equivalent of the API PATCH payload.
type recipePatch struct {
	Name         *string        `json:"name"`
	Description  *string        `json:"description"`
	Ingredients  *[]string      `json:"ingredients"`
	Instructions *[]string      `json:"instructions"`
	Category     *string        `json:"category"`
	Image        *string        `json:"image"`
	Photos       *[]types.Photo `json:"photos"`
}

func (p recipePatch) anySet() bool {
	return p.Name != nil || p.Description != nil || p.Ingredients != nil || p.Instructions != nil ||
		p.Category != nil || p.Image != nil || p.Photos != nil
}

func (p recipePatch) setFields() []string {
	var out []string
	if p.Name != nil {
		out = append(out, "name")
	}
	if p.Description != nil {
		out = append(out, "description")
	}
	if p.Ingredients != nil {
		out = append(out, "ingredients")
	}
	if p.Instructions != nil {
		out = append(out, "instructions")
	}
	if p.Category != nil {
		out = append(out, "category")
	}
	if p.Image != nil {
		out = append(out, "image")
	}
	if p.Photos != nil {
		out = append(out, "photos")
	}
	return out
}

func mergeRecipePatch(cur types.Recipe, p recipePatch) types.Recipe {
	out := cur
	if p.Name != nil {
		out.Name = *p.Name
	}
	if p.Description != nil {
		out.Description = *p.Description
	}
	if p.Ingredients != nil {
		out.Ingredients = *p.Ingredients
	}
	if p.Instructions != nil {
		out.Instructions = *p.Instructions
	}
	if p.Category != nil {
		out.Category = *p.Category
	}
	if p.Image != nil {
		out.Image = *p.Image
	}
	if p.Photos != nil {
		out.Photos = *p.Photos
	}
	return out
}
