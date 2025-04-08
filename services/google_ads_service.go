package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"poc-integracoes-onm/models"
)

// GoogleAdsConfig contém as configurações necessárias para autenticação com a API do Google Ads
type GoogleAdsConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	State        string
}

// GoogleAdsService implementa o serviço para integração com o Google Ads
type GoogleAdsService struct {
	// Configurações do serviço
	Config GoogleAdsConfig
}

// NewGoogleAdsService cria uma nova instância do serviço Google Ads
func NewGoogleAdsService() *GoogleAdsService {
	return &GoogleAdsService{}
}

// NewGoogleAdsServiceWithConfig cria uma nova instância do serviço Google Ads com configurações específicas
func NewGoogleAdsServiceWithConfig(clientID, clientSecret, redirectURI, state string) *GoogleAdsService {
	return &GoogleAdsService{
		Config: GoogleAdsConfig{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURI:  redirectURI,
			State:        state,
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

	// Usando a versão v1 (mais estável e bem documentada)
	apiURL := fmt.Sprintf("https://googleads.googleapis.com/v1/customers/%s:search", customerID)

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
