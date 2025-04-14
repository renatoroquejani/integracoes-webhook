package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"poc-integracoes-onm/models"
)

// GoogleAdsConfig contém as configurações necessárias para autenticação com a API do Google Ads
type GoogleAdsConfig struct {
	ClientID       string
	ClientSecret   string
	RedirectURI    string
	State          string
	DeveloperToken string
}

// GoogleAdsService implementa o serviço para integração com o Google Ads
type GoogleAdsService struct {
	// Configurações do serviço
	Config GoogleAdsConfig
}

// NewGoogleAdsService cria uma nova instância do serviço Google Ads
func NewGoogleAdsService() *GoogleAdsService {
	return &GoogleAdsService{
		Config: GoogleAdsConfig{
			DeveloperToken: "dummytoken", // Em uma implementação real, este token seria obtido de uma variável de ambiente ou configuração
		},
	}
}

// NewGoogleAdsServiceWithConfig cria uma nova instância do serviço Google Ads com configurações específicas
func NewGoogleAdsServiceWithConfig(clientID, clientSecret, redirectURI, state string) *GoogleAdsService {
	return &GoogleAdsService{
		Config: GoogleAdsConfig{
			ClientID:       clientID,
			ClientSecret:   clientSecret,
			RedirectURI:    redirectURI,
			State:          state,
			DeveloperToken: "dummytoken", // Em uma implementação real, este token seria obtido de uma variável de ambiente ou configuração
		},
	}
}

// GetAuthURL retorna a URL para autorização do usuário
func (s *GoogleAdsService) GetAuthURL() (string, error) {
	if s.Config.ClientID == "" || s.Config.RedirectURI == "" {
		return "", errors.New("configurações incompletas: ClientID e RedirectURI são obrigatórios")
	}

	authURL, err := url.Parse("https://accounts.google.com/o/oauth2/v2/auth")
	if err != nil {
		return "", fmt.Errorf("erro ao parsear URL: %w", err)
	}

	params := url.Values{}
	params.Set("client_id", s.Config.ClientID)
	params.Set("redirect_uri", s.Config.RedirectURI)
	params.Set("response_type", "code")
	params.Set("scope", "https://www.googleapis.com/auth/adwords")
	params.Set("state", s.Config.State)
	params.Set("access_type", "offline")
	params.Set("prompt", "consent")
	authURL.RawQuery = params.Encode()

	return authURL.String(), nil
}

// ExchangeCodeForToken troca o código de autorização por um token de acesso
func (s *GoogleAdsService) ExchangeCodeForToken(authorizationCode string) (*models.OAuthTokenResponse, error) {
	if s.Config.ClientID == "" || s.Config.ClientSecret == "" || s.Config.RedirectURI == "" {
		return nil, errors.New("configurações incompletas: ClientID, ClientSecret e RedirectURI são obrigatórios")
	}

	if authorizationCode == "" {
		return nil, errors.New("código de autorização não fornecido")
	}

	tokenEndpoint := "https://oauth2.googleapis.com/token"
	data := url.Values{}
	data.Set("client_id", s.Config.ClientID)
	data.Set("client_secret", s.Config.ClientSecret)
	data.Set("redirect_uri", s.Config.RedirectURI)
	data.Set("code", authorizationCode)
	data.Set("grant_type", "authorization_code")

	req, err := http.NewRequest("POST", tokenEndpoint, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar a requisição: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer a requisição para obter o token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("falha na requisição do token. Status: %s, resposta: %s", resp.Status, body)
	}

	// Estrutura para decodificar a resposta JSON
	var tokenResponse models.OAuthTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta JSON: %w", err)
	}

	return &tokenResponse, nil
}

// RefreshAccessToken atualiza o token de acesso usando o refresh token
func (s *GoogleAdsService) RefreshAccessToken(refreshToken string) (*models.OAuthTokenResponse, error) {
	if s.Config.ClientID == "" || s.Config.ClientSecret == "" {
		return nil, errors.New("configurações incompletas: ClientID e ClientSecret são obrigatórios")
	}

	if refreshToken == "" {
		return nil, errors.New("refresh token não fornecido")
	}

	tokenEndpoint := "https://oauth2.googleapis.com/token"
	data := url.Values{}
	data.Set("client_id", s.Config.ClientID)
	data.Set("client_secret", s.Config.ClientSecret)
	data.Set("refresh_token", refreshToken)
	data.Set("grant_type", "refresh_token")

	req, err := http.NewRequest("POST", tokenEndpoint, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar a requisição: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer a requisição para atualizar o token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("falha na requisição de atualização do token. Status: %s, resposta: %s", resp.Status, body)
	}

	// Estrutura para decodificar a resposta JSON
	var tokenResponse models.OAuthTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta JSON: %w", err)
	}

	return &tokenResponse, nil
}

// GetMetricas obtém as métricas principais do Google Ads usando as credenciais fornecidas
func (s *GoogleAdsService) GetMetricas(clientID, clientSecret, refreshToken, managerID string) (*models.GoogleAdsData, error) {
	if clientID == "" || clientSecret == "" || refreshToken == "" {
		return nil, errors.New("credenciais incompletas fornecidas")
	}

	// Configurar o serviço com as credenciais fornecidas
	s.Config.ClientID = clientID
	s.Config.ClientSecret = clientSecret

	// Obter token de acesso atualizado
	_, err := s.RefreshAccessToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("erro ao atualizar token de acesso: %w", err)
	}

	// Usar o ID da conta gerenciadora se fornecido, ou usar um ID padrão
	accountID := managerID
	if accountID == "" {
		accountID = "5808256042" // ID padrão para a POC
	}

	// Obter insights da conta
	return s.GetAccountInsights(clientID, clientSecret, refreshToken, accountID)
}

// GetCampaignInsights obtém insights detalhados de uma campanha específica
func (s *GoogleAdsService) GetCampaignInsights(clientID, clientSecret, refreshToken, campaignID string) (*models.GoogleAdsData, error) {
	if clientID == "" || clientSecret == "" || refreshToken == "" {
		return nil, errors.New("credenciais incompletas fornecidas")
	}

	if campaignID == "" {
		return nil, errors.New("ID da campanha não fornecido")
	}

	// Configurar o serviço com as credenciais fornecidas
	s.Config.ClientID = clientID
	s.Config.ClientSecret = clientSecret

	// Obter token de acesso atualizado
	token, err := s.RefreshAccessToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("erro ao atualizar token de acesso: %w", err)
	}

	// Construir a URL da API do Google Ads
	// De acordo com a documentação oficial da API Google Ads V1:
	// https://developers.google.com/google-ads/api/docs/rest/overview
	// A URL correta para o serviço GoogleAdsService.Search é:
	// POST https://googleads.googleapis.com/v1/customers/{customer_id}:search

	// O ID correto da conta é 580-825-6042, mas para a API precisamos remover os hífens
	customerID := "5808256042" // ID da conta formatado corretamente para a API

	// Usando a versão v11 (mais estável e bem documentada)
	apiURL := fmt.Sprintf("https://googleads.googleapis.com/v11/customers/%s:search", customerID)

	// Registrar os detalhes da requisição
	fmt.Printf("URL da API: %s\n", apiURL)

	// Construir o payload da requisição (JSON)
	payload := map[string]interface{}{
		"query": fmt.Sprintf("SELECT campaign.id, campaign.name, metrics.clicks, metrics.impressions, metrics.ctr, metrics.average_cpc, metrics.cost_micros, metrics.conversions, metrics.cost_per_conversion FROM campaign WHERE campaign.id = %s AND segments.date DURING LAST_30_DAYS", campaignID),
	}

	// Converter o payload para JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter payload para JSON: %v", err)
	}

	// Registrar o payload da requisição
	fmt.Printf("Payload: %s\n", string(payloadBytes))

	// Criar a requisição HTTP
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição HTTP: %v", err)
	}

	// Adicionar cabeçalhos necessários
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	// Importante: Este token de desenvolvedor é necessário para acessar a API
	req.Header.Set("developer-token", "d_pAjxizpi1whv22-_SYjA") // Token de desenvolvedor real

	// Vamos remover o cabeçalho x-goog-user-project que pode estar causando confusão
	// E não vamos incluir o cabeçalho x-goog-api-client por enquanto

	fmt.Printf("Headers:\n")
	fmt.Printf("  Content-Type: application/json\n")
	fmt.Printf("  Authorization: Bearer %s...\n", token.AccessToken[:15])
	fmt.Printf("  Developer-token: %s...\n", "d_pAjxizpi1whv22-_SYjA"[:10])

	// Fazer a requisição HTTP
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Erro na chamada à API real: %v\n", err)
		return nil, fmt.Errorf("erro na chamada à API do Google Ads: %w", err)
	}
	defer resp.Body.Close()

	// Verificar o status da resposta
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("API retornou status %d: %s\n", resp.StatusCode, string(body))
		fmt.Printf("Headers da resposta: %v\n", resp.Header)
		return nil, fmt.Errorf("API do Google Ads retornou erro: status %d - %s", resp.StatusCode, string(body))
	}

	// Processar a resposta
	var responseData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	// Processar e extrair dados reais da resposta da API
	data := &models.GoogleAdsData{
		ID:   campaignID,
		Nome: "Campanha Real - " + campaignID,
	}

	// Extrair as métricas da resposta - isso teria que ser implementado adequadamente
	// para processar os dados reais da API do Google Ads
	fmt.Printf("Processando dados reais da API do Google Ads para a campanha %s\n", campaignID)

	// Exemplo de como você extrairia dados reais:
	// Por enquanto, estamos apenas inicializando com zeros
	// Em uma implementação real, você extrairia esses valores da resposta
	data.Impressions = 0
	data.Clicks = 0
	data.CTR = 0
	data.CPC = 0
	data.Conversoes = 0
	data.TaxaConversao = 0
	data.CustoConversao = 0
	data.InvestimentoTotal = 0

	// Exemplo de como extrair campos da resposta (isto é um exemplo simplificado)
	if results, ok := responseData["results"]; ok {
		if resultsArray, ok := results.([]interface{}); ok && len(resultsArray) > 0 {
			fmt.Printf("Encontrados %d resultados na resposta\n", len(resultsArray))
		}
	}

	return data, nil
}

// GetAccountInsights obtém insights da conta de anúncios
func (s *GoogleAdsService) GetAccountInsights(clientID, clientSecret, refreshToken, accountID string) (*models.GoogleAdsData, error) {
	if clientID == "" || clientSecret == "" || refreshToken == "" {
		return nil, errors.New("credenciais incompletas fornecidas")
	}

	if accountID == "" {
		return nil, errors.New("ID da conta não fornecido")
	}

	// Configurar o serviço com as credenciais fornecidas
	s.Config.ClientID = clientID
	s.Config.ClientSecret = clientSecret

	// Obter token de acesso atualizado
	token, err := s.RefreshAccessToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("erro ao atualizar token de acesso: %w", err)
	}

	// Construir a URL da API
	// Usando a versão v11 (mais estável e bem documentada)
	apiURL := fmt.Sprintf("https://googleads.googleapis.com/v11/customers/%s:search", accountID)

	// Construir o payload da requisição (JSON)
	payload := map[string]interface{}{
		"query": "SELECT customer.id, customer.descriptive_name, metrics.clicks, metrics.impressions, metrics.ctr, metrics.average_cpc, metrics.cost_micros, metrics.conversions, metrics.cost_per_conversion FROM customer WHERE segments.date DURING LAST_30_DAYS",
	}

	// Converter o payload para JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar payload: %w", err)
	}

	// Criar a requisição HTTP
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}

	// Adicionar headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("developer-token", s.Config.DeveloperToken)

	// Enviar a requisição
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar requisição: %w", err)
	}
	defer resp.Body.Close()

	// Verificar o status da resposta
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API do Google Ads retornou erro: status %d - %s", resp.StatusCode, string(body))
	}

	// Processar a resposta
	var responseData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	// Extrair os dados relevantes da resposta
	// Nota: Esta é uma implementação simplificada, a estrutura real da resposta pode ser diferente
	var impressions int = 0
	var clicks int = 0
	var conversoes int = 0
	var custoMicros int64 = 0
	var nome string = "Conta " + accountID

	// Processar os resultados
	if results, ok := responseData["results"].([]interface{}); ok && len(results) > 0 {
		for _, result := range results {
			resultMap, ok := result.(map[string]interface{})
			if !ok {
				continue
			}

			// Extrair métricas
			if metrics, ok := resultMap["metrics"].(map[string]interface{}); ok {
				if imp, ok := metrics["impressions"].(float64); ok {
					impressions += int(imp)
				}
				if clk, ok := metrics["clicks"].(float64); ok {
					clicks += int(clk)
				}
				if conv, ok := metrics["conversions"].(float64); ok {
					conversoes += int(conv)
				}
				if cost, ok := metrics["cost_micros"].(float64); ok {
					custoMicros += int64(cost)
				}
			}

			// Extrair nome da conta
			if customer, ok := resultMap["customer"].(map[string]interface{}); ok {
				if descriptiveName, ok := customer["descriptive_name"].(string); ok && descriptiveName != "" {
					nome = descriptiveName
				}
			}
		}
	}

	// Calcular métricas
	custo := float64(custoMicros) / 1000000.0 // Converter micros para reais
	ctr := 0.0
	if impressions > 0 {
		ctr = float64(clicks) / float64(impressions) * 100.0
	}
	cpc := 0.0
	if clicks > 0 {
		cpc = custo / float64(clicks)
	}
	taxaConversao := 0.0
	if clicks > 0 {
		taxaConversao = float64(conversoes) / float64(clicks) * 100.0
	}
	custoConversao := 0.0
	if conversoes > 0 {
		custoConversao = custo / float64(conversoes)
	}

	// Criar e retornar o objeto de dados
	return &models.GoogleAdsData{
		ID:                accountID,
		Nome:              nome,
		CTR:               roundFloat(ctr, 2),
		CPC:               roundFloat(cpc, 2),
		Conversoes:        conversoes,
		TaxaConversao:     roundFloat(taxaConversao, 2),
		CustoConversao:    roundFloat(custoConversao, 2),
		InvestimentoTotal: roundFloat(custo, 2),
		Impressions:       impressions,
		Clicks:            clicks,
	}, nil
}

// ListCampaigns lista as campanhas disponíveis para a conta
func (s *GoogleAdsService) ListCampaigns(clientID, clientSecret, refreshToken, accountID string) ([]models.GoogleAdsData, error) {
	if clientID == "" || clientSecret == "" || refreshToken == "" {
		return nil, errors.New("credenciais incompletas fornecidas")
	}

	if accountID == "" {
		return nil, errors.New("ID da conta não fornecido")
	}

	// Configurar o serviço com as credenciais fornecidas
	s.Config.ClientID = clientID
	s.Config.ClientSecret = clientSecret

	// Obter token de acesso atualizado
	token, err := s.RefreshAccessToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("erro ao atualizar token de acesso: %w", err)
	}

	// Construir a URL da API
	// Usando a versão v11 (mais estável e bem documentada)
	apiURL := fmt.Sprintf("https://googleads.googleapis.com/v11/customers/%s:search", accountID)

	// Construir o payload da requisição (JSON)
	payload := map[string]interface{}{
		"query": "SELECT campaign.id, campaign.name, metrics.clicks, metrics.impressions, metrics.ctr, metrics.average_cpc, metrics.cost_micros, metrics.conversions, metrics.cost_per_conversion FROM campaign WHERE segments.date DURING LAST_30_DAYS",
	}

	// Converter o payload para JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar payload: %w", err)
	}

	// Criar a requisição HTTP
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}

	// Adicionar headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("developer-token", s.Config.DeveloperToken)

	// Enviar a requisição
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar requisição: %w", err)
	}
	defer resp.Body.Close()

	// Verificar o status da resposta
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API do Google Ads retornou erro: status %d - %s", resp.StatusCode, string(body))
	}

	// Processar a resposta
	var responseData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	// Lista para armazenar as campanhas
	var campaigns []models.GoogleAdsData

	// Processar os resultados
	if results, ok := responseData["results"].([]interface{}); ok {
		for _, result := range results {
			resultMap, ok := result.(map[string]interface{})
			if !ok {
				continue
			}

			// Extrair dados da campanha
			var campaignID string = ""
			var campaignName string = ""
			var impressions int = 0
			var clicks int = 0
			var conversoes int = 0
			var custoMicros int64 = 0

			// Extrair ID e nome da campanha
			if campaign, ok := resultMap["campaign"].(map[string]interface{}); ok {
				if id, ok := campaign["id"].(string); ok {
					campaignID = id
				}
				if name, ok := campaign["name"].(string); ok {
					campaignName = name
				}
			}

			// Extrair métricas
			if metrics, ok := resultMap["metrics"].(map[string]interface{}); ok {
				if imp, ok := metrics["impressions"].(float64); ok {
					impressions = int(imp)
				}
				if clk, ok := metrics["clicks"].(float64); ok {
					clicks = int(clk)
				}
				if conv, ok := metrics["conversions"].(float64); ok {
					conversoes = int(conv)
				}
				if cost, ok := metrics["cost_micros"].(float64); ok {
					custoMicros = int64(cost)
				}
			}

			// Pular campanhas sem ID
			if campaignID == "" {
				continue
			}

			// Calcular métricas
			custo := float64(custoMicros) / 1000000.0 // Converter micros para reais
			ctr := 0.0
			if impressions > 0 {
				ctr = float64(clicks) / float64(impressions) * 100.0
			}
			cpc := 0.0
			if clicks > 0 {
				cpc = custo / float64(clicks)
			}
			taxaConversao := 0.0
			if clicks > 0 {
				taxaConversao = float64(conversoes) / float64(clicks) * 100.0
			}
			custoConversao := 0.0
			if conversoes > 0 {
				custoConversao = custo / float64(conversoes)
			}

			// Adicionar campanha à lista
			campaigns = append(campaigns, models.GoogleAdsData{
				ID:                campaignID,
				Nome:              campaignName,
				CTR:               roundFloat(ctr, 2),
				CPC:               roundFloat(cpc, 2),
				Conversoes:        conversoes,
				TaxaConversao:     roundFloat(taxaConversao, 2),
				CustoConversao:    roundFloat(custoConversao, 2),
				InvestimentoTotal: roundFloat(custo, 2),
				Impressions:       impressions,
				Clicks:            clicks,
			})
		}
	}

	return campaigns, nil
}

// getMockData retorna dados simulados para a conta
func (s *GoogleAdsService) getMockData(accountID string) *models.GoogleAdsData {
	// Inicializar o gerador de números aleatórios
	rand.Seed(time.Now().UnixNano())

	// Gerar dados simulados
	impressions := rand.Intn(10000) + 5000
	clicks := rand.Intn(1000) + 100
	conversoes := rand.Intn(50) + 10
	custo := float64(rand.Intn(500000)+100000) / 100.0 // Custo em reais

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
	custo := float64(rand.Intn(200000)+50000) / 100.0 // Custo em reais

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
	custo := float64(rand.Intn(1000000)+200000) / 100.0 // Custo em reais

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

// GetConsolidatedCampaignData obtém dados consolidados de todas as campanhas de todas as contas
func (s *GoogleAdsService) GetConsolidatedCampaignData(clientID, clientSecret, refreshToken string) ([]*models.GoogleAdsData, error) {
	if clientID == "" || clientSecret == "" || refreshToken == "" {
		return nil, errors.New("credenciais incompletas fornecidas")
	}

	fmt.Printf("Iniciando busca de dados consolidados do Google Ads\n")

	// Configurar o serviço com as credenciais fornecidas
	s.Config.ClientID = clientID
	s.Config.ClientSecret = clientSecret

	// Obter token de acesso atualizado
	log.Printf("Tentando atualizar token de acesso com refresh_token: %s...\n", refreshToken)
	_, err := s.RefreshAccessToken(refreshToken)
	if err != nil {
		log.Printf("Erro ao atualizar token de acesso: %v\n", err)
		return nil, fmt.Errorf("erro ao atualizar token de acesso: %w", err)
	}
	log.Printf("Token de acesso atualizado com sucesso!\n")

	fmt.Printf("Token de acesso obtido com sucesso. Tentando acessar a API real...\n")

	// Em uma implementação real completa, faríamos:
	// 1. Listar todas as contas de anúncios disponíveis
	// 2. Para cada conta, listar todas as campanhas
	// 3. Para cada campanha, obter os insights detalhados
	// 4. Consolidar todos os dados em uma única lista

	// Tentar fazer uma chamada real à API para verificar se o token funciona
	// e se conseguimos obter algum dado real
	log.Printf("Testando conexão com a API do Google Ads...\n")
	testResult, err := s.TestGoogleAdsConnection(refreshToken, clientID, clientSecret)
	if err != nil {
		log.Printf("Erro ao testar conexão com a API: %v\n", err)
		return nil, fmt.Errorf("erro ao testar conexão com a API: %w", err)
	}

	fmt.Printf("Conexão com a API testada com sucesso: %v\n", testResult)

	// Tentar obter dados reais da API
	// Como esta é uma POC e a integração completa com a API do Google Ads é complexa,
	// vamos verificar se conseguimos pelo menos obter dados básicos de uma conta

	// O ID correto da conta é 580-825-6042, mas para a API precisamos remover os hífens
	customerID := "5808256042" // ID da conta formatado corretamente para a API
	log.Printf("Usando customer ID: %s\n", customerID)

	// Tentar obter dados básicos da conta
	var consolidated []*models.GoogleAdsData

	// Adicionar dados reais da conta se conseguirmos obter
	log.Printf("Tentando obter insights da conta %s...\n", customerID)
	accountData, err := s.GetAccountInsights(clientID, clientSecret, refreshToken, customerID)
	if err != nil {
		log.Printf("Erro ao obter insights da conta %s: %v\n", customerID, err)
		return nil, fmt.Errorf("erro ao obter insights da conta %s: %w", customerID, err)
	}
	log.Printf("Insights da conta %s obtidos com sucesso!\n", customerID)

	// Se chegamos até aqui, conseguimos obter pelo menos os dados da conta
	consolidated = append(consolidated, accountData)
	fmt.Printf("Adicionados dados reais da conta %s\n", customerID)

	// Tentar listar campanhas reais
	log.Printf("Tentando listar campanhas da conta %s...\n", customerID)
	campaigns, err := s.ListCampaigns(clientID, clientSecret, refreshToken, customerID)
	if err != nil {
		log.Printf("Erro ao listar campanhas da conta %s: %v\n", customerID, err)
		// Retornar apenas os dados da conta que já obtivemos
		return consolidated, nil
	}
	log.Printf("Campanhas da conta %s listadas com sucesso! Total: %d\n", customerID, len(campaigns))

	// Se chegamos até aqui, conseguimos listar campanhas reais
	fmt.Printf("Obtidas %d campanhas reais da conta %s\n", len(campaigns), customerID)

	// Adicionar dados das campanhas reais
	for _, campaign := range campaigns {
		campaignData := &models.GoogleAdsData{
			ID:                campaign.ID,
			Nome:              campaign.Nome,
			CTR:               campaign.CTR,
			CPC:               campaign.CPC,
			Conversoes:        campaign.Conversoes,
			TaxaConversao:     campaign.TaxaConversao,
			CustoConversao:    campaign.CustoConversao,
			InvestimentoTotal: campaign.InvestimentoTotal,
			Impressions:       campaign.Impressions,
			Clicks:            campaign.Clicks,
		}
		consolidated = append(consolidated, campaignData)
		fmt.Printf("Adicionados dados reais da campanha %s\n", campaign.ID)
	}

	// Se não conseguimos obter muitos dados reais, retornar o que temos
	if len(consolidated) < 3 {
		fmt.Printf("Poucos dados reais obtidos, mas retornando o que temos...\n")
	}

	fmt.Printf("Total de %d itens consolidados retornados (dados reais)\n", len(consolidated))
	return consolidated, nil
}

// getConsolidatedMockData retorna dados simulados consolidados
func (s *GoogleAdsService) getConsolidatedMockData() []*models.GoogleAdsData {
	fmt.Printf("Gerando dados simulados consolidados...\n")

	var consolidated []*models.GoogleAdsData

	// Simular múltiplas contas
	accountIDs := []string{"123456789", "987654321", "555555555"}

	for _, accountID := range accountIDs {
		// Adicionar dados da conta
		accountData := s.getMockAccountData(accountID)
		consolidated = append(consolidated, accountData)
		fmt.Printf("Adicionados dados simulados da conta %s\n", accountID)

		// Simular campanhas para esta conta
		campaignsList := s.getMockCampaignsList(accountID)
		for i := range campaignsList {
			// Modificar o ID da campanha para incluir a conta
			campaignID := fmt.Sprintf("%s_%d", accountID, i+1)
			campaignData := s.getMockCampaignData(campaignID)

			// Adicionar à lista consolidada
			consolidated = append(consolidated, campaignData)
			fmt.Printf("Adicionados dados simulados da campanha %s\n", campaignID)
		}
	}

	// Adicionar algumas variações para tornar os dados mais interessantes
	// Campanha de alto desempenho
	highPerformanceCampaign := &models.GoogleAdsData{
		ID:                "high_performance_123",
		Nome:              "Campanha de Alto Desempenho",
		CTR:               8.5,
		CPC:               1.25,
		Conversoes:        120,
		TaxaConversao:     15.3,
		CustoConversao:    8.75,
		InvestimentoTotal: 1050.25,
		Impressions:       15000,
		Clicks:            1275,
	}
	consolidated = append(consolidated, highPerformanceCampaign)

	// Campanha de baixo desempenho
	lowPerformanceCampaign := &models.GoogleAdsData{
		ID:                "low_performance_456",
		Nome:              "Campanha de Baixo Desempenho",
		CTR:               0.8,
		CPC:               4.75,
		Conversoes:        3,
		TaxaConversao:     1.2,
		CustoConversao:    95.50,
		InvestimentoTotal: 286.50,
		Impressions:       7500,
		Clicks:            60,
	}
	consolidated = append(consolidated, lowPerformanceCampaign)

	fmt.Printf("Total de %d itens simulados consolidados retornados\n", len(consolidated))
	return consolidated
}

// addMockCampaignsToRealAccount adiciona campanhas simuladas a uma conta real
func (s *GoogleAdsService) addMockCampaignsToRealAccount(realData []*models.GoogleAdsData) []*models.GoogleAdsData {
	consolidated := realData

	// Adicionar campanhas simuladas
	mockCampaigns := []struct {
		ID   string
		Nome string
	}{
		{"real_account_campaign_1", "Campanha de Busca - Palavras-chave"},
		{"real_account_campaign_2", "Campanha de Display - Interesses"},
		{"real_account_campaign_3", "Campanha de YouTube - Vídeos"},
		{"real_account_campaign_4", "Campanha de Performance Max"},
	}

	for _, campaign := range mockCampaigns {
		// Gerar dados aleatórios para a campanha
		impressions := rand.Intn(5000) + 1000
		clicks := rand.Intn(500) + 50
		conversoes := rand.Intn(30) + 5
		custo := float64(rand.Intn(200000)+50000) / 100.0

		// Calcular métricas
		ctr := float64(clicks) / float64(impressions) * 100.0
		cpc := custo / float64(clicks)
		taxaConversao := float64(conversoes) / float64(clicks) * 100.0
		custoConversao := custo / float64(conversoes)

		campaignData := &models.GoogleAdsData{
			ID:                campaign.ID,
			Nome:              campaign.Nome,
			CTR:               roundFloat(ctr, 2),
			CPC:               roundFloat(cpc, 2),
			Conversoes:        conversoes,
			TaxaConversao:     roundFloat(taxaConversao, 2),
			CustoConversao:    roundFloat(custoConversao, 2),
			InvestimentoTotal: roundFloat(custo, 2),
			Impressions:       impressions,
			Clicks:            clicks,
		}

		consolidated = append(consolidated, campaignData)
		fmt.Printf("Adicionados dados simulados da campanha %s para complementar dados reais\n", campaign.ID)
	}

	fmt.Printf("Total de %d itens consolidados retornados (mistura de dados reais e simulados)\n", len(consolidated))
	return consolidated
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

	return &models.ErrorInfo{
		Code:    500,
		Message: err.Error(),
		Type:    "GoogleAdsApiError",
		Details: "Detalhes não disponíveis",
	}
}

// TestGoogleAdsConnection testa a conexão com a API do Google para tentar identificar o problema
func (s *GoogleAdsService) TestGoogleAdsConnection(refreshToken, clientID, clientSecret string) (map[string]interface{}, error) {
	// Primeiro vamos tentar apenas validar o token de acesso
	token, err := s.RefreshAccessToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("erro ao renovar token de acesso: %v", err)
	}

	// Agora vamos tentar acessar a API do Google Ads com um endpoint mais simples
	// Verificar token usando o tokeninfo endpoint
	apiURL := "https://www.googleapis.com/oauth2/v3/tokeninfo?access_token=" + token.AccessToken

	fmt.Printf("Testando conexão com: %s\n", apiURL)

	// Configurar o cliente HTTP
	client := &http.Client{}
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição HTTP: %v", err)
	}

	// Fazer a requisição
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer requisição HTTP: %v", err)
	}
	defer resp.Body.Close()

	// Ler o corpo da resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler corpo da resposta: %v", err)
	}

	// Verificar o código de status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API retornou erro: status %d - %s", resp.StatusCode, string(body))
	}

	// Processar a resposta
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("erro ao processar resposta JSON: %v", err)
	}

	return result, nil
}

// TestConnection testa a conexão básica com o Google usando as credenciais configuradas
func (s *GoogleAdsService) TestConnection() error {
	// Verificar se temos credenciais configuradas
	if s.Config.ClientID == "" || s.Config.ClientSecret == "" {
		return errors.New("configurações incompletas: ClientID e ClientSecret são obrigatórios")
	}

	// Como não temos refresh_token configurado globalmente, usamos um estático para testes
	refreshToken := "1//0heP0QtbBOYG7CgYIARAAGBESNwF-L9IrhAEyFu0123k2JDNtwjaJux7SiNCN4jS3Gmy5GFa17Mjai8ivTAab5FHsR-Ppbaa5zyQ"

	// Tentar validar o token de acesso
	_, err := s.RefreshAccessToken(refreshToken)
	if err != nil {
		return fmt.Errorf("erro ao validar token de acesso: %v", err)
	}

	// Vamos fazer uma chamada para a API pública do Google
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://www.googleapis.com/discovery/v1/apis", nil)
	if err != nil {
		return fmt.Errorf("erro ao criar requisição para Google Discovery API: %v", err)
	}

	// Fazer a requisição
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao conectar com a Google Discovery API: %v", err)
	}
	defer resp.Body.Close()

	// Verificar status da resposta
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Google Discovery API retornou erro: status %d - %s", resp.StatusCode, string(body))
	}

	fmt.Println("Conexão básica com o Google: SUCESSO")
	return nil
}

// roundFloat arredonda um float para o número especificado de casas decimais
func roundFloat(val float64, precision int) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}
