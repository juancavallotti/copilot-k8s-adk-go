package main

import (
	"testing"

	"google.golang.org/genai"
)

func TestFirstInlineImageDataReturnsFirstImage(t *testing.T) {
	want := []byte("image-bytes")
	response := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []*genai.Part{
						{Text: "description"},
						{InlineData: &genai.Blob{Data: want, MIMEType: "image/png"}},
					},
				},
			},
		},
	}

	got, err := firstInlineImageData(response)
	if err != nil {
		t.Fatalf("firstInlineImageData() error = %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("firstInlineImageData() = %q, want %q", got, want)
	}
}

func TestFirstInlineImageDataErrorsWithoutImage(t *testing.T) {
	response := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []*genai.Part{{Text: "text only"}},
				},
			},
		},
	}

	if _, err := firstInlineImageData(response); err == nil {
		t.Fatal("firstInlineImageData() error = nil, want error")
	}
}
