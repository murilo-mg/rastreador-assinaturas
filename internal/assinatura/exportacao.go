package assinatura

import (
	"bytes"
	"encoding/csv"
	"net/http"
	"strconv"
	"strings"
	"unicode"

	"rastreador-assinaturas/internal/httpjson"
)

func (h *Handler) exportarCSV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		httpjson.Erro(w, http.StatusMethodNotAllowed, "Método não permitido.")
		return
	}
	assinaturas, err := h.repositorio.Listar(r.Context())
	if err != nil {
		httpjson.Interno(w, err)
		return
	}

	// Monta o arquivo antes de responder para não baixar um CSV parcial em caso de erro.
	var arquivo bytes.Buffer
	arquivo.WriteString("\uFEFF") // BOM UTF-8 para preservar acentos no Excel.
	writer := csv.NewWriter(&arquivo)
	writer.Comma = ';'
	registros := [][]string{{"ID", "Nome", "Valor mensal (R$)", "Categoria", "Dia da cobrança", "Ativa"}}
	for _, a := range assinaturas {
		ativa := "Não"
		if a.Ativa {
			ativa = "Sim"
		}
		valor := strings.Replace(strconv.FormatFloat(a.Valor, 'f', 2, 64), ".", ",", 1)
		registros = append(registros, []string{strconv.Itoa(a.ID), textoCSV(a.Nome), valor,
			textoCSV(a.Categoria), strconv.Itoa(a.DiaCobranca), ativa})
	}
	if err := writer.WriteAll(registros); err != nil {
		httpjson.Interno(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="assinaturas.csv"`)
	w.Header().Set("Cache-Control", "no-store")
	w.Write(arquivo.Bytes())
}

// A tabulação inicial fica dentro de aspas pelo encoding/csv. Reduz a
// interpretação como fórmula no Excel sem alterar o texto salvo no banco.
// Referência: https://community.owasp.org/attacks/CSV_Injection
func textoCSV(texto string) string {
	inicio := strings.TrimLeftFunc(texto, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsControl(r) || unicode.Is(unicode.Cf, r)
	})
	if strings.HasPrefix(inicio, "=") || strings.HasPrefix(inicio, "+") ||
		strings.HasPrefix(inicio, "-") || strings.HasPrefix(inicio, "@") ||
		strings.HasPrefix(inicio, "＝") || strings.HasPrefix(inicio, "＋") ||
		strings.HasPrefix(inicio, "－") || strings.HasPrefix(inicio, "＠") {
		return "\t" + texto
	}
	return texto
}
