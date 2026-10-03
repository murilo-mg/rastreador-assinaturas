// Ponto de entrada: API e painel usam a mesma origem.
package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
	_ "time/tzdata"

	"rastreador-assinaturas/internal/assinatura"
	"rastreador-assinaturas/internal/database"
	"rastreador-assinaturas/internal/httpjson"
	"rastreador-assinaturas/internal/relatorio"
	"rastreador-assinaturas/internal/web"
)

func criarHandler(banco *sql.DB) http.Handler {
	mux := http.NewServeMux()
	web.RegistrarRotas(mux)
	mux.HandleFunc("GET /saude", func(w http.ResponseWriter, r *http.Request) {
		if err := banco.PingContext(r.Context()); err != nil {
			httpjson.Erro(w, http.StatusServiceUnavailable, "Banco de dados indisponível.")
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("ok"))
	})
	assinatura.NovoHandler(assinatura.NovoRepositorio(banco)).RegistrarRotas(mux)
	relatorio.NovoHandler(relatorio.NovoRepositorio(banco)).RegistrarRotas(mux)
	return proteger(mux)
}

func proteger(proximo http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self'; connect-src 'self'; object-src 'none'; base-uri 'none'; form-action 'self'; frame-ancestors 'none'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Cache-Control", "no-store")
		// Rejeita escritas de outras páginas; clientes locais como curl não enviam Origin.
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			if origem := r.Header.Get("Origin"); origem != "" {
				u, err := url.Parse(origem)
				if err != nil || u.Host != r.Host || (u.Scheme != "http" && u.Scheme != "https") {
					httpjson.Erro(w, http.StatusForbidden, "Origem não permitida.")
					return
				}
			}
		}
		ctx, cancelar := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancelar()
		proximo.ServeHTTP(w, r.WithContext(ctx))
	})
}

func main() {
	banco, err := database.Conectar()
	if err != nil {
		log.Fatalf("não foi possível conectar ao banco: %v", err)
	}
	defer banco.Close()
	endereco := os.Getenv("SERVER_ADDR")
	if endereco == "" {
		endereco = "127.0.0.1:8080"
	}
	servidor := &http.Server{
		Addr: endereco, Handler: criarHandler(banco),
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second,
	}
	log.Printf("rastreador disponível em %s", endereco)
	log.Fatal(servidor.ListenAndServe())
}
