package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"llmdebate/internal/ai"
	"llmdebate/internal/debate"
	appws "llmdebate/internal/ws"
)

func main() {
	addr := env("ADDR", ":8080")
	streamer := ai.Streamer(ai.NewMockStreamer())
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		streamer = ai.NewOpenAIStreamer(ai.OpenAIConfig{
			APIKey: key,
			Model:  env("OPENAI_MODEL", "gpt-5.5"),
		})
		log.Printf("using OpenAI streamer")
	} else {
		log.Printf("OPENAI_API_KEY not set; using mock streamer")
	}
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
