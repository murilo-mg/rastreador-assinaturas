// Define a estrutura de dados de uma assinatura.
package assinatura

type Assinatura struct {
	ID          int     `json:"id"`
	Nome        string  `json:"nome"`
	Valor       float64 `json:"valor"`
	Categoria   string  `json:"categoria"`
	DiaCobranca int     `json:"dia_cobranca"`
	Ativa       bool    `json:"ativa"`
}