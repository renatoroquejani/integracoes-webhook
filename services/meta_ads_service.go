package services

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"poc-integracoes-onm/models"

	fb "github.com/huandu/facebook/v2"
)

// Configurar a versão da API do Meta Ads globalmente
func init() {
	fb.Version = "v22.0"
}

// MetaAdsService implementa o serviço para integração com o Meta Ads
type MetaAdsService struct {
	// Configurações do serviço, se necessário
}

// NewMetaAdsService cria uma nova instância do serviço Meta Ads
func NewMetaAdsService() *MetaAdsService {
	return &MetaAdsService{}
}

// GetMetricas obtém as métricas principais do Meta Ads usando o token fornecido
func (s *MetaAdsService) GetMetricas(token string) (*models.MetaAdsData, error) {
	if token == "" {
		return nil, errors.New("token não fornecido")
	}

	// Criar uma sessão do Facebook com o token fornecido
	session := fb.New("", "").Session(token)

	// Verificar se o token é válido fazendo uma chamada simples
	_, err := session.Get("/me", fb.Params{})
	if err != nil {
		return nil, fmt.Errorf("token inválido ou não autorizado: %w", err)
	}

	// Obter as contas de anúncios disponíveis para o token
	params := fb.Params{
		"fields": "account_id,name",
		"limit":  "1", // Apenas a primeira conta para simplificar
	}
	res, err := session.Get("/me/adaccounts", params)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter contas de anúncios: %w", err)
	}

	// Verificar se há contas de anúncios disponíveis
	data, ok := res["data"].([]interface{})
	if !ok || len(data) == 0 {
		return nil, errors.New("nenhuma conta de anúncios encontrada para este token")
	}

	// Obter o ID da primeira conta de anúncios
	account := data[0].(map[string]interface{})
	accountID, ok := account["account_id"].(string)
	if !ok {
		return nil, errors.New("ID da conta de anúncios não encontrado")
	}

	// Obter insights da conta de anúncios
	return s.GetAccountInsights(token, accountID)
}

// GetCampaignInsights obtém insights detalhados de uma campanha específica
func (s *MetaAdsService) GetCampaignInsights(token string, campaignID string) (*models.MetaAdsData, error) {
	if token == "" {
		return nil, errors.New("token não fornecido")
	}

	if campaignID == "" {
		return nil, errors.New("ID da campanha não fornecido")
	}

	// Criar uma sessão do Facebook com o token fornecido
	session := fb.New("", "").Session(token)

	// Obter informações da campanha
	campaignParams := fb.Params{
		"fields": "name",
	}
	campaignRes, err := session.Get("/"+campaignID, campaignParams)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter informações da campanha: %w", err)
	}

	campanhaNome, _ := campaignRes["name"].(string)

	// Obter insights da campanha
	params := fb.Params{
		"fields":      "clicks,impressions,spend,actions,cost_per_action_type",
		"date_preset": "last_30d",
		"level":       "campaign",
	}
	res, err := session.Get("/"+campaignID+"/insights", params)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter insights da campanha: %w", err)
	}

	return processInsightsData(res, campaignID, campanhaNome)
}

// GetAccountInsights obtém insights da conta de anúncios
func (s *MetaAdsService) GetAccountInsights(token string, accountID string) (*models.MetaAdsData, error) {
	if token == "" {
		return nil, errors.New("token não fornecido")
	}

	if accountID == "" {
		return nil, errors.New("ID da conta não fornecido")
	}

	// Criar uma sessão do Facebook com o token fornecido
	session := fb.New("", "").Session(token)

	// Obter insights da conta de anúncios
	params := fb.Params{
		"fields":      "clicks,impressions,spend,actions,cost_per_action_type",
		"date_preset": "last_30d",
		"level":       "account",
	}
	res, err := session.Get("/act_"+accountID+"/insights", params)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter insights da conta: %w", err)
	}

	return processInsightsData(res, "", "Conta "+accountID)
}

// processInsightsData processa os dados de insights retornados pela API do Facebook
func processInsightsData(res fb.Result, id string, nome string) (*models.MetaAdsData, error) {
	data := &models.MetaAdsData{
		ID:   id,
		Nome: nome,
	}

	// Verificar se temos dados de insights
	insights, ok := res["data"].([]interface{})
	if !ok || len(insights) == 0 {
		return data, nil // Retornar dados vazios se não houver insights
	}

	// Obter o primeiro insight (normalmente é o único para o período especificado)
	insight, ok := insights[0].(map[string]interface{})
	if !ok {
		return data, nil
	}

	// Extrair métricas básicas
	clicks, _ := insight["clicks"].(string)
	impressions, _ := insight["impressions"].(string)
	spend, _ := insight["spend"].(string)

	// Converter para números
	clicksNum, _ := strconv.ParseFloat(clicks, 64)
	impressionsNum, _ := strconv.ParseFloat(impressions, 64)
	spendNum, _ := strconv.ParseFloat(spend, 64)

	// Calcular CTR
	var ctr float64
	if impressionsNum > 0 {
		ctr = (clicksNum / impressionsNum) * 100
	}

	// Extrair ações (vendas)
	var vendas float64
	actions, ok := insight["actions"].([]interface{})
	if ok {
		for _, action := range actions {
			actionMap, ok := action.(map[string]interface{})
			if !ok {
				continue
			}

			// Verificar se é uma ação de compra
			actionType, _ := actionMap["action_type"].(string)
			if actionType == "purchase" || actionType == "offsite_conversion.fb_pixel_purchase" {
				actionValue, _ := actionMap["value"].(string)
				if actionValue == "" {
					// Se não tiver valor, apenas contar o número de ações
					vendas++
				} else {
					// Se tiver valor, somar ao total
					valueNum, _ := strconv.ParseFloat(actionValue, 64)
					vendas += valueNum
				}
			}
		}
	}

	// Calcular CAC (Custo de Aquisição de Cliente)
	var cac float64
	if vendas > 0 {
		cac = spendNum / vendas
	}

	// Preencher os dados
	data.CTR = ctr
	data.CAC = cac
	data.InvestimentoTotal = spendNum
	data.NumeroVendas = int(vendas)

	return data, nil
}

// FallbackToMockData retorna dados simulados quando a API real falha
func (s *MetaAdsService) FallbackToMockData(err error) (*models.MetaAdsResponse, error) {
	// Extrair informações detalhadas do erro
	errorInfo := extractErrorInfoFromMain(err)

	// Criar dados simulados para demonstração
	mockData := &models.MetaAdsData{
		ID:                "mock_campaign_123",
		Nome:              "Campanha Simulada",
		CTR:               2.5,
		CAC:               15.75,
		InvestimentoTotal: 1250.50,
		NumeroVendas:      80,
	}

	// Retornar resposta com dados simulados e informações do erro
	return &models.MetaAdsResponse{
		Success: false,
		Message: "Usando dados simulados devido a falha na API. Detalhes do erro abaixo.",
		Data:    *mockData,
		Error:   errorInfo,
	}, nil
}

// extractErrorInfoFromMain extrai informações detalhadas de um erro
func extractErrorInfoFromMain(err error) *models.ErrorInfo {
	errorInfo := &models.ErrorInfo{
		Message: err.Error(),
	}

	// Verificar se é um erro do Facebook SDK
	if fbErr, ok := err.(*fb.Error); ok {
		errorInfo.Code = fbErr.Code
		errorInfo.Type = fbErr.Type
		errorInfo.Message = fbErr.Message
		errorInfo.Details = fbErr.TraceID

		// Tratar especificamente o erro de versão da API obsoleta
		if fbErr.Code == 2635 {
			errorInfo.Type = "DeprecatedAPIVersionError"
			errorInfo.Details = "A versão da API do Meta Ads utilizada está obsoleta. A aplicação será atualizada para usar a versão mais recente (v22.0)."
		}
	} else {
		// Tentar extrair mais informações do erro
		errorStr := err.Error()
		if strings.Contains(errorStr, "token inválido") {
			errorInfo.Type = "AuthenticationError"
			errorInfo.Code = 190
			errorInfo.Details = "O token fornecido é inválido ou expirou"
		} else if strings.Contains(errorStr, "permissão") {
			errorInfo.Type = "PermissionError"
			errorInfo.Code = 200
			errorInfo.Details = "O token não tem permissões suficientes"
		} else if strings.Contains(errorStr, "limite") {
			errorInfo.Type = "RateLimitError"
			errorInfo.Code = 4
			errorInfo.Details = "Limite de requisições excedido"
		} else if strings.Contains(errorStr, "deprecated version") || strings.Contains(errorStr, "2635") {
			errorInfo.Type = "DeprecatedAPIVersionError"
			errorInfo.Code = 2635
			errorInfo.Details = "A versão da API do Meta Ads utilizada está obsoleta. A aplicação será atualizada para usar a versão mais recente (v22.0)."
		}
	}

	return errorInfo
}
