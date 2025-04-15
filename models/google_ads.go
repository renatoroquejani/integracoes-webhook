package models

// GoogleAdsRequest representa a solicitação para consulta de dados do Google Ads
type GoogleAdsRequest struct {
	ClientID     string `json:"client_id" binding:"required"`     // ID do cliente OAuth
	ClientSecret string `json:"client_secret" binding:"required"` // Secret do cliente OAuth
	RefreshToken string `json:"refresh_token" binding:"required"` // Token de atualização OAuth
	ManagerID    string `json:"manager_id,omitempty"`              // ID da conta gerenciadora (opcional)
}

// GoogleAdsData contém as métricas do Google Ads
type GoogleAdsData struct {
	ID                string  `json:"id,omitempty"`                // ID da campanha ou conta
	Nome              string  `json:"nome,omitempty"`              // Nome da campanha ou conta
	CTR               float64 `json:"ctr"`                         // Click-Through Rate
	CPC               float64 `json:"cpc"`                         // Custo por Clique
	Conversoes        int     `json:"conversoes"`                   // Número de conversões
	TaxaConversao     float64 `json:"taxa_conversao"`              // Taxa de conversão
	CustoConversao    float64 `json:"custo_conversao"`             // Custo por conversão
	InvestimentoTotal float64 `json:"investimento_total"`          // Investimento Total
	Impressions       int     `json:"impressions"`                 // Número de impressões
	Clicks            int     `json:"clicks"`                      // Número de cliques
}

// GoogleAdsResponse representa a resposta com dados do Google Ads
type GoogleAdsResponse struct {
	Success bool          `json:"success"`           // Indica se a operação foi bem-sucedida
	Message string        `json:"message"`           // Mensagem descritiva
	Data    GoogleAdsData `json:"data,omitempty"`   // Dados do Google Ads
	Error   *ErrorInfo    `json:"error,omitempty"` // Informações de erro, se houver
}

// GoogleAdsCampaignListResponse representa a resposta com lista de campanhas
type GoogleAdsCampaignListResponse struct {
	Success bool             `json:"success"`           // Indica se a operação foi bem-sucedida
	Message string           `json:"message"`           // Mensagem descritiva
	Data    []GoogleAdsData  `json:"data,omitempty"`   // Lista de campanhas
	Error   *ErrorInfo       `json:"error,omitempty"` // Informações de erro, se houver
}

// GoogleAdsAccountInfo representa as informações básicas de uma conta do Google Ads
type GoogleAdsAccountInfo struct {
	ID               string `json:"id"`                      // ID da conta
	DescriptiveName  string `json:"descriptive_name"`        // Nome descritivo da conta
	CurrencyCode     string `json:"currency_code"`           // Código da moeda (ex: BRL, USD)
	TimeZone         string `json:"time_zone"`               // Fuso horário da conta
	TrackingTemplate string `json:"tracking_template,omitempty"` // Template de rastreamento
	AutoTagging      bool   `json:"auto_tagging"`             // Se a marcação automática está ativada
	TestAccount      bool   `json:"test_account"`             // Se é uma conta de teste
}

// GoogleAdsAccountInfoResponse representa a resposta com informações básicas da conta
type GoogleAdsAccountInfoResponse struct {
	Success bool                `json:"success"`           // Indica se a operação foi bem-sucedida
	Message string              `json:"message"`           // Mensagem descritiva
	Data    *GoogleAdsAccountInfo `json:"data,omitempty"`   // Informações da conta
	Error   *ErrorInfo          `json:"error,omitempty"` // Informações de erro, se houver
}
