package services

import (
	"errors"
	"math"
	"math/rand"
	"time"

	"poc-integracoes-onm/models"
)

// GoogleAdsService implementa o serviço para integração com o Google Ads
type GoogleAdsService struct {
	// Configurações do serviço, se necessário
}

// NewGoogleAdsService cria uma nova instância do serviço Google Ads
func NewGoogleAdsService() *GoogleAdsService {
	return &GoogleAdsService{}
}

// GetMetricas obtém as métricas principais do Google Ads usando as credenciais fornecidas
func (s *GoogleAdsService) GetMetricas(clientID, clientSecret, refreshToken, managerID string) (*models.GoogleAdsData, error) {
	if clientID == "" || clientSecret == "" || refreshToken == "" {
		return nil, errors.New("credenciais incompletas fornecidas")
	}

	// Para a POC, vamos retornar dados simulados
	return s.getMockData(managerID), nil
}

// GetCampaignInsights obtém insights detalhados de uma campanha específica
func (s *GoogleAdsService) GetCampaignInsights(clientID, clientSecret, refreshToken, campaignID string) (*models.GoogleAdsData, error) {
	if clientID == "" || clientSecret == "" || refreshToken == "" {
		return nil, errors.New("credenciais incompletas fornecidas")
	}

	if campaignID == "" {
		return nil, errors.New("ID da campanha não fornecido")
	}

	// Para a POC, vamos retornar dados simulados
	return s.getMockCampaignData(campaignID), nil
}

// GetAccountInsights obtém insights da conta de anúncios
func (s *GoogleAdsService) GetAccountInsights(clientID, clientSecret, refreshToken, accountID string) (*models.GoogleAdsData, error) {
	if clientID == "" || clientSecret == "" || refreshToken == "" {
		return nil, errors.New("credenciais incompletas fornecidas")
	}

	if accountID == "" {
		return nil, errors.New("ID da conta não fornecido")
	}

	// Para a POC, vamos retornar dados simulados
	return s.getMockAccountData(accountID), nil
}

// ListCampaigns lista as campanhas disponíveis para a conta
func (s *GoogleAdsService) ListCampaigns(clientID, clientSecret, refreshToken, accountID string) ([]models.GoogleAdsData, error) {
	if clientID == "" || clientSecret == "" || refreshToken == "" {
		return nil, errors.New("credenciais incompletas fornecidas")
	}

	if accountID == "" {
		return nil, errors.New("ID da conta não fornecido")
	}

	// Para a POC, vamos retornar dados simulados
	return s.getMockCampaignsList(accountID), nil
}

// getMockData retorna dados simulados para a conta
func (s *GoogleAdsService) getMockData(accountID string) *models.GoogleAdsData {
	// Inicializar o gerador de números aleatórios
	rand.Seed(time.Now().UnixNano())

	// Gerar dados simulados
	impressions := rand.Intn(10000) + 5000
	clicks := rand.Intn(1000) + 100
	conversoes := rand.Intn(50) + 10
	custo := float64(rand.Intn(500000) + 100000) / 100.0 // Custo em reais

	// Calcular métricas
	ctr := float64(clicks) / float64(impressions) * 100.0
	cpc := custo / float64(clicks)
	taxaConversao := float64(conversoes) / float64(clicks) * 100.0
	custoConversao := custo / float64(conversoes)

	return &models.GoogleAdsData{
		ID:                accountID,
		Nome:              "Conta " + accountID,
		CTR:               roundFloat(ctr, 2),
		CPC:               roundFloat(cpc, 2),
		Conversoes:        conversoes,
		TaxaConversao:     roundFloat(taxaConversao, 2),
		CustoConversao:    roundFloat(custoConversao, 2),
		InvestimentoTotal: roundFloat(custo, 2),
		Impressions:       impressions,
		Clicks:            clicks,
	}
}

// getMockCampaignData retorna dados simulados para uma campanha específica
func (s *GoogleAdsService) getMockCampaignData(campaignID string) *models.GoogleAdsData {
	// Inicializar o gerador de números aleatórios
	rand.Seed(time.Now().UnixNano())

	// Gerar dados simulados
	impressions := rand.Intn(5000) + 1000
	clicks := rand.Intn(500) + 50
	conversoes := rand.Intn(30) + 5
	custo := float64(rand.Intn(200000) + 50000) / 100.0 // Custo em reais

	// Calcular métricas
	ctr := float64(clicks) / float64(impressions) * 100.0
	cpc := custo / float64(clicks)
	taxaConversao := float64(conversoes) / float64(clicks) * 100.0
	custoConversao := custo / float64(conversoes)

	// Gerar nome da campanha baseado no ID
	nomeCampanha := "Campanha "
	switch campaignID {
	case "1":
		nomeCampanha += "Busca - Palavras-chave"
	case "2":
		nomeCampanha += "Display - Interesses"
	case "3":
		nomeCampanha += "YouTube - Vídeos"
	default:
		nomeCampanha += campaignID
	}

	return &models.GoogleAdsData{
		ID:                campaignID,
		Nome:              nomeCampanha,
		CTR:               roundFloat(ctr, 2),
		CPC:               roundFloat(cpc, 2),
		Conversoes:        conversoes,
		TaxaConversao:     roundFloat(taxaConversao, 2),
		CustoConversao:    roundFloat(custoConversao, 2),
		InvestimentoTotal: roundFloat(custo, 2),
		Impressions:       impressions,
		Clicks:            clicks,
	}
}

// getMockAccountData retorna dados simulados para uma conta específica
func (s *GoogleAdsService) getMockAccountData(accountID string) *models.GoogleAdsData {
	// Inicializar o gerador de números aleatórios
	rand.Seed(time.Now().UnixNano())

	// Gerar dados simulados
	impressions := rand.Intn(20000) + 10000
	clicks := rand.Intn(2000) + 500
	conversoes := rand.Intn(100) + 20
	custo := float64(rand.Intn(1000000) + 200000) / 100.0 // Custo em reais

	// Calcular métricas
	ctr := float64(clicks) / float64(impressions) * 100.0
	cpc := custo / float64(clicks)
	taxaConversao := float64(conversoes) / float64(clicks) * 100.0
	custoConversao := custo / float64(conversoes)

	return &models.GoogleAdsData{
		ID:                accountID,
		Nome:              "Conta " + accountID,
		CTR:               roundFloat(ctr, 2),
		CPC:               roundFloat(cpc, 2),
		Conversoes:        conversoes,
		TaxaConversao:     roundFloat(taxaConversao, 2),
		CustoConversao:    roundFloat(custoConversao, 2),
		InvestimentoTotal: roundFloat(custo, 2),
		Impressions:       impressions,
		Clicks:            clicks,
	}
}

// getMockCampaignsList retorna uma lista simulada de campanhas
func (s *GoogleAdsService) getMockCampaignsList(accountID string) []models.GoogleAdsData {
	// Criar uma lista de campanhas simuladas
	campaigns := []models.GoogleAdsData{
		*s.getMockCampaignData("1"),
		*s.getMockCampaignData("2"),
		*s.getMockCampaignData("3"),
	}

	return campaigns
}

// FallbackToMockData retorna dados simulados quando a API real falha
func (s *GoogleAdsService) FallbackToMockData(err error) (*models.GoogleAdsResponse, error) {
	// Extrair informações do erro
	errorInfo := extractErrorInfoFromGoogleAds(err)

	// Criar resposta com dados simulados
	response := &models.GoogleAdsResponse{
		Success: false,
		Message: "Erro ao acessar a API do Google Ads. Usando dados simulados.",
		Data:    *s.getMockData("123456789"),
		Error:   errorInfo,
	}

	return response, nil
}

// extractErrorInfoFromGoogleAds extrai informações detalhadas de um erro da API do Google Ads
func extractErrorInfoFromGoogleAds(err error) *models.ErrorInfo {
	if err == nil {
		return nil
	}

	errorInfo := &models.ErrorInfo{
		Message: err.Error(),
		Type:    "GoogleAdsApiError",
	}

	return errorInfo
}

// roundFloat arredonda um float para o número especificado de casas decimais
func roundFloat(val float64, precision int) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}
