package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"llmdebate/internal/debate"
	appws "llmdebate/internal/ws"
)

const defaultListenAddr = "127.0.0.1:8080"

func main() {
	addr := env("ADDR", defaultListenAddr)
	streamer, provider, err := selectStreamer(streamerOptions{
		Provider:        env("AI_PROVIDER", ""),
		OpenAIAPIKey:    os.Getenv("OPENAI_API_KEY"),
		OpenAIModel:     env("OPENAI_MODEL", "gpt-5.5"),
		CursorAPIKey:    os.Getenv("CURSOR_API_KEY"),
		CursorBaseURL:   os.Getenv("CURSOR_BASE_URL"),
		CursorModel:     os.Getenv("CURSOR_MODEL"),
		AnthropicAPIKey: os.Getenv("ANTHROPIC_API_KEY"),
		AnthropicModel:  os.Getenv("ANTHROPIC_MODEL"),
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("using %s streamer", provider)

	hub := appws.NewHub()
	runner := debate.Runner{Streamer: streamer}
	mux := http.NewServeMux()
	mux.Handle("/ws", appws.Handler{Hub: hub, Runner: runner})
	frontendDist := filepath.Join("frontend", "dist")
	mux.Handle("/", spaHandler(frontendDist))
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func spaHandler(dir string) http.Handler {
	fs := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(dir, filepath.Clean(r.URL.Path))
		if _, err := os.Stat(path); err != nil {
			if filepath.Ext(r.URL.Path) != "" {
				http.NotFound(w, r)
				return
			}
			http.ServeFile(w, r, filepath.Join(dir, "index.html"))
			return
		}
		fs.ServeHTTP(w, r)
	})
}
