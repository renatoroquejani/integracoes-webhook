package models

// MetaAdsRequest representa a solicitação para consulta de dados do Meta Ads
type MetaAdsRequest struct {
	Token string `json:"token" binding:"required"`
}

// OAuthTokenResponse estrutura para decodificar a resposta JSON do token de acesso OAuth
type OAuthTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// MetaAdsData contém as métricas do Meta Ads
type MetaAdsData struct {
	ID              string  `json:"id,omitempty"`              // ID da campanha ou conta
	Nome            string  `json:"nome,omitempty"`            // Nome da campanha ou conta
	CTR             float64 `json:"ctr"`                       // Click-Through Rate
	CAC             float64 `json:"cac"`                       // Custo de Aquisição por Cliente
	InvestimentoTotal float64 `json:"investimento_total"`       // Investimento Total
	NumeroVendas    int     `json:"numero_vendas"`             // Número de Vendas
}

// MetaAdsResponse representa a resposta com dados do Meta Ads
type MetaAdsResponse struct {
	Success bool        `json:"success"`           // Indica se a operação foi bem-sucedida
	Message string      `json:"message"`           // Mensagem descritiva
	Data    MetaAdsData `json:"data,omitempty"`   // Dados do Meta Ads
	Error   *ErrorInfo  `json:"error,omitempty"` // Informações de erro, se houver
}

// ErrorInfo representa informações detalhadas sobre um erro
type ErrorInfo struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Type    string `json:"type,omitempty"`
	Details string `json:"details,omitempty"`
}
