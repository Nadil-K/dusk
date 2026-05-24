// Example: chi app with dusk deprecation middleware.
// Run: cd examples/chi && go run chi_app.go
package main

import (
	"encoding/json"
	"log"
	"net/http"

	gochi "github.com/go-chi/chi/v5"
	dusk "github.com/Nadil-K/dusk/packages/go/adapters/chi"
)

func main() {
	r := gochi.NewRouter()
	r.Use(dusk.New("../../dusk.yaml"))

	r.Get("/api/v1/users", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"users": []any{},
			"note":  "deprecated — use /api/v2/users",
		})
	})

	r.Get("/api/v2/users", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"users": []any{},
			"page":  1,
			"total": 0,
		})
	})

	r.Get("/api/v1/orders/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"order_id": gochi.URLParam(r, "id"),
		})
	})

	r.Get("/api/v2/orders/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"order_id": gochi.URLParam(r, "id"),
		})
	})

	log.Println("chi server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
