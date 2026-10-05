// O painel é incorporado ao binário; somente estes recursos são publicados.
package web

import (
	"bytes"
	"embed"
	"net/http"
	"time"
)

//go:embed arquivos/*
var arquivos embed.FS

func RegistrarRotas(mux *http.ServeMux) {
	for rota, nome := range map[string]string{
		"/{$}": "index.html", "/assets/app.css": "app.css",
		"/assets/app.js": "app.js", "/assets/favicon.svg": "favicon.svg",
		"/assets/demo.mjs": "demo.mjs",
	} {
		conteudo, err := arquivos.ReadFile("arquivos/" + nome)
		if err != nil {
			panic(err)
		}
		mux.HandleFunc("GET "+rota, func(w http.ResponseWriter, r *http.Request) {
			http.ServeContent(w, r, nome, time.Time{}, bytes.NewReader(conteudo))
		})
	}
}
