package assinatura

import (
	"errors"
	"math"
	"strings"
	"unicode/utf8"
)

func (a *Assinatura) Validar() error {
	a.Nome = strings.TrimSpace(a.Nome)
	a.Categoria = strings.TrimSpace(a.Categoria)
	if !utf8.ValidString(a.Nome) || a.Nome == "" || utf8.RuneCountInString(a.Nome) > 100 {
		return errors.New("O nome deve ter entre 1 e 100 caracteres.")
	}
	if !utf8.ValidString(a.Categoria) || utf8.RuneCountInString(a.Categoria) > 50 {
		return errors.New("A categoria deve ter até 50 caracteres.")
	}
	if math.IsNaN(a.Valor) || math.IsInf(a.Valor, 0) || a.Valor < 0 || a.Valor > 99999999.99 {
		return errors.New("O valor deve estar entre 0 e 99.999.999,99 reais.")
	}
	centavos := a.Valor * 100
	if math.Abs(centavos-math.Round(centavos)) > 0.000001 {
		return errors.New("O valor deve ter no máximo duas casas decimais.")
	}
	if a.DiaCobranca < 1 || a.DiaCobranca > 31 {
		return errors.New("O dia da cobrança deve estar entre 1 e 31.")
	}
	return nil
}
