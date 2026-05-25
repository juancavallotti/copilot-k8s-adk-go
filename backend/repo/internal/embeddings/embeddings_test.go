package embeddings

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNoopReturnsDisabled(t *testing.T) {
	t.Parallel()
	_, err := Noop{}.Embed(context.Background(), "hi")
	if !errors.Is(err, ErrDisabled) {
		t.Fatalf("err = %v, want ErrDisabled", err)
	}
}

func TestFormatVector(t *testing.T) {
	t.Parallel()
	got := FormatVector([]float32{0.5, -1, 0})
	want := "[0.5,-1,0]"
	if got != want {
		t.Fatalf("FormatVector = %q, want %q", got, want)
	}
	if FormatVector(nil) != "[]" {
		t.Fatalf("FormatVector(nil) = %q, want []", FormatVector(nil))
	}
}

func TestGeminiClientSuccess(t *testing.T) {
	t.Parallel()
	values := make([]float32, Dimensions)
	for i := range values {
		values[i] = float32(i) / 1000.0
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-goog-api-key") != "test-key" {
			http.Error(w, "missing key", http.StatusUnauthorized)
			return
		}
		var got embedRequest
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if got.OutputDimensionality != Dimensions {
			http.Error(w, "wrong dims", http.StatusBadRequest)
			return
		}
		if len(got.Content.Parts) != 1 || got.Content.Parts[0].Text != "hello" {
			http.Error(w, "wrong content", http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(embedResponse{
			Embedding: struct {
				Values []float32 `json:"values"`
			}{Values: values},
		})
	}))
	defer srv.Close()

	c := NewGeminiClient("test-key")
	c.endpoint = srv.URL

	out, err := c.Embed(context.Background(), "hello")
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != Dimensions {
		t.Fatalf("len = %d, want %d", len(out), Dimensions)
	}
}

func TestGeminiClientRejectsEmpty(t *testing.T) {
	t.Parallel()
	c := NewGeminiClient("k")
	if _, err := c.Embed(context.Background(), "   "); err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestGeminiClientErrorStatus(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":{"message":"quota"}}`))
	}))
	defer srv.Close()
	c := NewGeminiClient("k")
	c.endpoint = srv.URL
	_, err := c.Embed(context.Background(), "hi")
	if err == nil || !strings.Contains(err.Error(), "quota") {
		t.Fatalf("err = %v, want quota error", err)
	}
}

func TestOpenAIClientSuccess(t *testing.T) {
	t.Parallel()
	values := make([]float32, Dimensions)
	for i := range values {
		values[i] = float32(i) / 1000.0
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			http.Error(w, "missing key", http.StatusUnauthorized)
			return
		}
		var got openAIRequest
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if got.Dimensions != Dimensions || got.Model != OpenAIModel || got.Input != "hello" {
			http.Error(w, "wrong request", http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(openAIResponse{
			Data: []struct {
				Embedding []float32 `json:"embedding"`
			}{{Embedding: values}},
		})
	}))
	defer srv.Close()

	c := NewOpenAIClient("test-key")
	c.endpoint = srv.URL

	out, err := c.Embed(context.Background(), "hello")
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != Dimensions {
		t.Fatalf("len = %d, want %d", len(out), Dimensions)
	}
}

func TestOpenAIClientErrorStatus(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":{"message":"rate limited"}}`))
	}))
	defer srv.Close()
	c := NewOpenAIClient("k")
	c.endpoint = srv.URL
	_, err := c.Embed(context.Background(), "hi")
	if err == nil || !strings.Contains(err.Error(), "rate limited") {
		t.Fatalf("err = %v, want rate limited error", err)
	}
}

func TestNewFromEnvSelection(t *testing.T) {
	cases := []struct {
		name     string
		env      map[string]string
		want     Provider
		wantType any
	}{
		{
			name:     "no keys",
			env:      map[string]string{},
			want:     ProviderNoop,
			wantType: Noop{},
		},
		{
			name:     "only gemini",
			env:      map[string]string{"GEMINI_API_KEY": "g"},
			want:     ProviderGemini,
			wantType: (*GeminiClient)(nil),
		},
		{
			name:     "only openai",
			env:      map[string]string{"OPENAI_API_KEY": "o"},
			want:     ProviderOpenAI,
			wantType: (*OpenAIClient)(nil),
		},
		{
			name:     "both keys defaults to gemini",
			env:      map[string]string{"GEMINI_API_KEY": "g", "OPENAI_API_KEY": "o"},
			want:     ProviderGemini,
			wantType: (*GeminiClient)(nil),
		},
		{
			name:     "override forces openai",
			env:      map[string]string{"GEMINI_API_KEY": "g", "OPENAI_API_KEY": "o", "EMBEDDING_PROVIDER": "openai"},
			want:     ProviderOpenAI,
			wantType: (*OpenAIClient)(nil),
		},
		{
			name:     "override gemini without key falls back to noop",
			env:      map[string]string{"OPENAI_API_KEY": "o", "EMBEDDING_PROVIDER": "gemini"},
			want:     ProviderNoop,
			wantType: Noop{},
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			for _, k := range []string{"GEMINI_API_KEY", "OPENAI_API_KEY", "EMBEDDING_PROVIDER"} {
				t.Setenv(k, "")
			}
			for k, v := range tc.env {
				t.Setenv(k, v)
			}
			c, p := NewFromEnv()
			if p != tc.want {
				t.Fatalf("provider = %q, want %q", p, tc.want)
			}
			switch tc.wantType.(type) {
			case Noop:
				if _, ok := c.(Noop); !ok {
					t.Fatalf("client = %T, want Noop", c)
				}
			case *GeminiClient:
				if _, ok := c.(*GeminiClient); !ok {
					t.Fatalf("client = %T, want *GeminiClient", c)
				}
			case *OpenAIClient:
				if _, ok := c.(*OpenAIClient); !ok {
					t.Fatalf("client = %T, want *OpenAIClient", c)
				}
			}
		})
	}
}

func TestGeminiClientDimensionMismatch(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(embedResponse{
			Embedding: struct {
				Values []float32 `json:"values"`
			}{Values: []float32{0.1, 0.2}},
		})
	}))
	defer srv.Close()
	c := NewGeminiClient("k")
	c.endpoint = srv.URL
	_, err := c.Embed(context.Background(), "hi")
	if err == nil || !strings.Contains(err.Error(), "want 768") {
		t.Fatalf("err = %v, want dimension mismatch", err)
	}
}
