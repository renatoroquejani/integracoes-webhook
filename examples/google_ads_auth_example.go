package main

import (
	"fmt"
	"log"

	"poc-integracoes-onm/services"
)

func main() {
	// Substitua pelos valores do seu projeto no Google Cloud Console
	clientID := "YOUR_GOOGLE_CLIENT_ID"
	clientSecret := "YOUR_GOOGLE_CLIENT_SECRET"
	redirectURI := "https://yourapp.com/oauth2callback"
	state := "random_csrf_token" // Use um valor aleatório para prevenir CSRF

	// Criar um serviço do Google Ads com as configurações
	googleService := services.NewGoogleAdsServiceWithConfig(
		clientID,
		clientSecret,
		redirectURI,
		state,
	)

	// Etapa 1: Obter a URL de autorização do Google
	authURL, err := googleService.GetAuthURL()
	if err != nil {
		log.Fatal("Erro ao gerar a URL de autorização:", err)
	}

	fmt.Println("Redirecione o usuário para a seguinte URL para autorização:")
	fmt.Println(authURL)

	// Em um cenário real, o usuário seria redirecionado para a URL acima e,
	// após a autorização, o Google redirecionaria para a URL de callback com um parâmetro "code".
	// Para fins de demonstração, vamos ler esse código manualmente.
	fmt.Println("\nInforme o código de autorização obtido:")
	var code string
	fmt.Scan(&code)

	// Etapa 2: Trocar o código de autorização por um token de acesso
	tokenResponse, err := googleService.ExchangeCodeForToken(code)
	if err != nil {
		log.Fatal("Erro ao obter o token de acesso:", err)
	}

	fmt.Println("\nToken obtido com sucesso!")
	fmt.Println("Access Token:", tokenResponse.AccessToken)
	fmt.Println("Token Type:", tokenResponse.TokenType)
	fmt.Println("Expires In:", tokenResponse.ExpiresIn, "segundos")
	
	// Importante: o Refresh Token só é retornado na primeira autorização
	if tokenResponse.RefreshToken != "" {
		fmt.Println("Refresh Token:", tokenResponse.RefreshToken)
		fmt.Println("\nIMPORTANTE: Guarde o Refresh Token em um local seguro.")
		fmt.Println("Ele será necessário para renovar o Access Token sem precisar de nova autorização.")
	} else {
		fmt.Println("\nNota: Nenhum Refresh Token foi retornado.")
		fmt.Println("Isso pode acontecer se você já autorizou este aplicativo anteriormente.")
	}

	// Demonstração de como renovar o token usando o refresh token
	if tokenResponse.RefreshToken != "" {
		fmt.Println("\n--- Demonstração de renovação de token ---")
		fmt.Println("Renovando o Access Token usando o Refresh Token...")
		
		newTokenResponse, err := googleService.RefreshAccessToken(tokenResponse.RefreshToken)
		if err != nil {
			fmt.Println("Erro ao renovar o token:", err)
		} else {
			fmt.Println("Novo Access Token obtido com sucesso!")
			fmt.Println("Novo Access Token:", newTokenResponse.AccessToken)
			fmt.Println("Expira em:", newTokenResponse.ExpiresIn, "segundos")
		}
	}

	fmt.Println("\n--- Próximos passos ---")
	fmt.Println("1. Use o Access Token para fazer chamadas à API do Google Ads")
	fmt.Println("2. Armazene o Refresh Token de forma segura para uso futuro")
	fmt.Println("3. Use o Refresh Token para obter novos Access Tokens quando expirarem")
}
