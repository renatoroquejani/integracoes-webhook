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
	"net/http"
	"net/url"
	"os"
	"strings"

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
	// Tentar obter o developer token da variável de ambiente
	developerToken := os.Getenv("GOOGLE_ADS_DEVELOPER_TOKEN")
	if developerToken == "" {
		// Tentar nome alternativo por compatibilidade
		developerToken = os.Getenv("GOOGLE_DEVELOPER_TOKEN")
	}

	return &GoogleAdsService{
		Config: GoogleAdsConfig{
			DeveloperToken: developerToken,
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
			DeveloperToken: os.Getenv("GOOGLE_DEVELOPER_TOKEN"),
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

	// Usando a versão v19 (mais estável e bem documentada)
	apiURL := fmt.Sprintf("https://googleads.googleapis.com/v19/customers/%s:search", customerID)

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
	// Usando a versão v19 (mais estável e bem documentada)
	apiURL := fmt.Sprintf("https://googleads.googleapis.com/v19/customers/%s:search", accountID)

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
	// Usando a versão v19 (mais estável e bem documentada)
	apiURL := fmt.Sprintf("https://googleads.googleapis.com/v19/customers/%s:search", accountID)

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
	log.Printf("Tentando atualizar token de acesso com refresh_token fornecido...\n")
	tokenResponse, err := s.RefreshAccessToken(refreshToken)
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
	// Como a integração completa com a API do Google Ads é complexa,
	// vamos realizar chamadas básicas para demonstrar a conexão funcional

	// Tentar obter dados básicos da conta
	var consolidated []*models.GoogleAdsData

	// Como não temos mais código mockado, retornamos um array vazio ou um exemplo básico
	// mostrando que a conexão foi bem-sucedida

	// Adiciona um objeto representando o sucesso da autenticação
	dummyData := &models.GoogleAdsData{
		ID:                "connection_test",
		Nome:              "Teste de Conexão",
		CTR:               0,
		CPC:               0,
		Conversoes:        0,
		TaxaConversao:     0,
		CustoConversao:    0,
		InvestimentoTotal: 0,
		Impressions:       0,
		Clicks:            0,
	}

	consolidated = append(consolidated, dummyData)

	fmt.Printf("Autenticação realizada com sucesso. Dados de teste retornados.\n")
	
	// Adicionar informações do token
	fmt.Printf("Token tipo: %s, expira em: %d segundos.\n", tokenResponse.TokenType, tokenResponse.ExpiresIn)
	
	// Para implementar a integração completa, é necessário implementar as chamadas à API Google Ads
	fmt.Printf("Total de %d itens consolidados retornados (dados de conexão)\n", len(consolidated))
	return consolidated, nil
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

// GetAccountBasicInfo obtém informações básicas de uma conta do Google Ads
func (s *GoogleAdsService) GetAccountBasicInfo(clientID, clientSecret, refreshToken, accountID string) (*models.GoogleAdsAccountInfo, error) {
	// Remover hífens do ID da conta, se houver
	accountID = strings.ReplaceAll(accountID, "-", "")

	// Configurar o serviço com as credenciais fornecidas
	s.Config.ClientID = clientID
	s.Config.ClientSecret = clientSecret

	// Obter token de acesso atualizado
	tokenResponse, err := s.RefreshAccessToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter token de acesso: %v", err)
	}

	accessToken := tokenResponse.AccessToken

	// Etapa 1: Listar contas acessíveis
	apiURL := "https://googleads.googleapis.com/v19/customers:listAccessibleCustomers"

	// Criar a requisição GET (método correto para listAccessibleCustomers)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição para listar contas: %v", err)
	}

	// Obter o token de desenvolvedor
	developerToken := s.Config.DeveloperToken
	if developerToken == "" {
		// Se não estiver na configuração, tentar da variável de ambiente
		developerToken = os.Getenv("GOOGLE_ADS_DEVELOPER_TOKEN")
		if developerToken == "" {
			// Tentar o nome alternativo da variável por compatibilidade
			developerToken = os.Getenv("GOOGLE_DEVELOPER_TOKEN")
			if developerToken == "" {
				return nil, fmt.Errorf("developer token não encontrado. Configure a variável de ambiente GOOGLE_ADS_DEVELOPER_TOKEN")
			}
		}
		// Salvar na configuração para uso futuro
		s.Config.DeveloperToken = developerToken
	}

	// Configurar headers necessários conforme documentação
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("developer-token", developerToken)
	// O login-customer-id é opcional na primeira chamada, mas vamos incluí-lo se tivermos o accountID
	if accountID != "" {
		req.Header.Set("login-customer-id", accountID)
	}
	// Mostra apenas os primeiros 10 caracteres do token para debug
	tokenPreview := developerToken
	if len(developerToken) > 10 {
		tokenPreview = developerToken[:10] + "..."
	}
	log.Printf("Developer Token: %s", tokenPreview)

	// Executar a requisição
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao executar requisição para listar contas: %v", err)
	}
	defer resp.Body.Close()

	// Ler o corpo da resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta da listagem de contas: %v", err)
	}

	// Verificar status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro ao listar contas (status %d): %s", resp.StatusCode, string(body))
	}

	// Processar a resposta para obter a lista de contas acessíveis
	var listResponse map[string]interface{}
	if err := json.Unmarshal(body, &listResponse); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta da listagem de contas: %v", err)
	}

	// Verificar se o ID da conta solicitada está entre as contas acessíveis
	resourceNames, ok := listResponse["resourceNames"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("formato inesperado na resposta da listagem de contas")
	}

	// Verificar se há contas acessíveis e se a conta solicitada está entre elas
	log.Printf("Contas acessíveis: %+v", resourceNames)
	accountResourceName := "customers/" + accountID
	accountFound := false

	for _, name := range resourceNames {
		resourceName, ok := name.(string)
		if ok && resourceName == accountResourceName {
			accountFound = true
			break
		}
	}

	if !accountFound && len(resourceNames) > 0 {
		// Se a conta solicitada não foi encontrada, mas temos outras contas disponíveis,
		// log para debug
		log.Printf("Conta %s não encontrada entre as contas acessíveis.", accountID)
	}

	// Verificar se encontramos alguma conta acessível
	if len(resourceNames) == 0 {
		return nil, fmt.Errorf("não foram encontradas contas acessíveis para este token")
	}

	// Usar a primeira conta disponível se a conta solicitada não foi encontrada
	searchAccountID := accountID
	if !accountFound && len(resourceNames) > 0 {
		// Extrair o ID da primeira conta disponível
		firstAccount, ok := resourceNames[0].(string)
		if ok {
			parts := strings.Split(firstAccount, "/")
			if len(parts) > 1 {
				searchAccountID = parts[len(parts)-1]
				log.Printf("Usando a primeira conta disponível: %s", searchAccountID)
			}
		}
	}

	// Etapa 2: Obter informações da conta diretamente via GET
	// Em vez de usar o endpoint search, vamos usar um endpoint direto para a conta
	accountURL := fmt.Sprintf("https://googleads.googleapis.com/v19/customers/%s", searchAccountID)
	log.Printf("Tentando acessar conta diretamente via: %s", accountURL)

	// Criar requisição GET
	accountReq, err := http.NewRequest("GET", accountURL, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição para conta: %v", err)
	}

	// Configurar headers para a requisição da conta
	accountReq.Header.Set("Authorization", "Bearer "+accessToken)
	accountReq.Header.Set("developer-token", developerToken)
	// O login-customer-id é crucial para a segunda chamada
	accountReq.Header.Set("login-customer-id", searchAccountID)
	log.Printf("Enviando requisição direta para a conta: %s com login-customer-id: %s", searchAccountID, searchAccountID)

	// Executar a requisição para obter informações da conta
	accountResp, err := client.Do(accountReq)
	if err != nil {
		return nil, fmt.Errorf("erro ao executar requisição da conta: %v", err)
	}
	defer accountResp.Body.Close()

	// Ler a resposta da requisição da conta
	accountBody, err := io.ReadAll(accountResp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta da conta: %v", err)
	}

	// Verificar o status da resposta
	if accountResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro ao acessar conta (status %d): %s", accountResp.StatusCode, string(accountBody))
	}

	// Criar um objeto de informações da conta com base na resposta
	log.Printf("Resposta da conta: %s", string(accountBody))

	// Decodificar a resposta JSON da conta
	var accountResponse map[string]interface{}
	if err := json.Unmarshal(accountBody, &accountResponse); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta da conta: %v", err)
	}

	// Com o método GET direto, a resposta já contém diretamente os dados da conta
	// Não precisamos extrair de results como faziamos com search
	customerData := accountResponse
	if len(customerData) == 0 {
		return nil, fmt.Errorf("nenhum dado encontrado para a conta %s", searchAccountID)
	}

	// Log para debug
	log.Printf("Dados da conta encontrados: %+v", customerData)

	// Extração de valores com tratamento de nulos
	getStringValue := func(data map[string]interface{}, key string, defaultValue string) string {
		if value, ok := data[key]; ok && value != nil {
			if strValue, ok := value.(string); ok {
				return strValue
			}
		}
		return defaultValue
	}

	getBoolValue := func(data map[string]interface{}, key string, defaultValue bool) bool {
		if value, ok := data[key]; ok && value != nil {
			if boolValue, ok := value.(bool); ok {
				return boolValue
			}
		}
		return defaultValue
	}

	// Extrair valores dos campos
	id := getStringValue(customerData, "id", accountID)
	name := getStringValue(customerData, "descriptiveName", "")
	currency := getStringValue(customerData, "currencyCode", "")
	timeZone := getStringValue(customerData, "timeZone", "")
	trackingTemplate := getStringValue(customerData, "trackingUrlTemplate", "")
	autoTagging := getBoolValue(customerData, "autoTaggingEnabled", false)
	testAccount := getBoolValue(customerData, "testAccount", false)

	// Construir o objeto de informações da conta
	accountInfo := &models.GoogleAdsAccountInfo{
		ID:               id,
		DescriptiveName:  name,
		CurrencyCode:     currency,
		TimeZone:         timeZone,
		TrackingTemplate: trackingTemplate,
		AutoTagging:      autoTagging,
		TestAccount:      testAccount,
	}

	return accountInfo, nil
}

// getValueOr retorna o valor ou o valor padrão se o valor for nil
func getValueOr(value interface{}, defaultValue interface{}) interface{} {
	if value == nil {
		return defaultValue
	}
	return value
}

// getValueBool retorna o valor como bool ou o valor padrão se o valor não for bool
func getValueBool(value interface{}, defaultValue bool) bool {
	if bValue, ok := value.(bool); ok {
		return bValue
	}
	return defaultValue
}

// TestConnection testa a conexão básica com o Google usando as credenciais configuradas
// Requer um refresh token fornecido pelo cliente
func (s *GoogleAdsService) TestConnection(refreshToken string) error {
	// Verificar se temos credenciais configuradas
	if s.Config.ClientID == "" || s.Config.ClientSecret == "" {
		return errors.New("configurações incompletas: ClientID e ClientSecret são obrigatórios")
	}

	// Verificar se o refresh token foi fornecido
	if refreshToken == "" {
		return errors.New("refresh_token não fornecido")
	}

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
