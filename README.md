# POC Integrações Webhook

Este é um projeto de prova de conceito para integração de webhooks de diversos serviços como Kiwify, Hotmart e Kirvano.

## Requisitos

- Go 1.21 ou superior
- Git

## Como executar

1. Clone o repositório
   ```bash
   git clone https://github.com/renatoroquejani/integracoes-webhook.git
   cd integracoes-webhook
   ```

2. Execute o comando para baixar as dependências:
   ```bash
   go mod download
   ```

3. Execute a aplicação:
   ```bash
   go run main.go
   ```

O servidor estará rodando em http://localhost:8080

## Documentação da API

A documentação da API está disponível via Swagger UI em:
http://localhost:8080/swagger/index.html

## Endpoints disponíveis

### POST /webhook/kiwify

Recebe webhooks da Kiwify. Suporta tanto notificações de compra quanto de carrinho abandonado.

Parâmetros:
- Query: `signature` (obrigatório) - Assinatura para validação
- Body: Payload do webhook no formato JSON

Exemplo de uso com curl:
```bash
# Para uma compra aprovada
curl -X POST "http://localhost:8080/webhook/kiwify?signature=6ae35bad036fda55be4b27d5e9a93dd6f0992b62" \
  -H "Content-Type: application/json" \
  -d @payloads/kiwify/compra_aprovada.json

# Para um carrinho abandonado
curl -X POST "http://localhost:8080/webhook/kiwify?signature=2094d6f8b07871ad4a35f59b5f9051cf7c50f58e" \
  -H "Content-Type: application/json" \
  -d @payloads/kiwify/abandono_de_carrinho.json
```

### POST /webhook/hotmart

Recebe webhooks da Hotmart.

Exemplo de uso com curl:
```bash
curl -X POST http://localhost:8080/webhook/hotmart \
  -H "Content-Type: application/json" \
  -d @payloads/hotmart/compra_aprovada.json
```

### POST /webhook/kirvano

Recebe webhooks da Kirvano.

Exemplo de uso com curl:
```bash
curl -X POST http://localhost:8080/webhook/kirvano \
  -H "Content-Type: application/json" \
  -d @payloads/kirvano/compra_aprovada.json
```

## API de Integração com Meta Ads

Esta API permite consultar dados do Meta Ads (Facebook/Instagram) usando um token de acesso. As métricas disponíveis incluem:

- CTR (Click-Through Rate)
- CAC (Custo de Aquisição por Cliente)
- Investimento Total
- Número de Vendas

### Como obter um token de acesso do Meta Ads

Para utilizar esta API, você precisa de um token de acesso válido do Meta Ads. Siga os passos abaixo para obter um token:

1. Acesse o [Meta for Developers](https://developers.facebook.com/) e faça login com sua conta do Facebook
2. Crie um aplicativo no painel do desenvolvedor (Developer Dashboard)
3. Adicione o produto "Marketing API" ao seu aplicativo
4. Gere um token de acesso com as permissões: `ads_read`, `ads_management`, `business_management`
5. Para uso em produção, você precisará passar pelo processo de revisão do aplicativo

**Nota:** Para fins de teste, você pode usar um token de acesso temporário gerado no [Graph API Explorer](https://developers.facebook.com/tools/explorer/).

### Integração com a API real

Esta API se integra diretamente com a Graph API do Facebook para obter dados reais das suas campanhas publicitárias. A integração utiliza a biblioteca oficial [huandu/facebook](https://github.com/huandu/facebook) para Go.

A API realiza as seguintes operações:

1. Valida o token de acesso fornecido
2. Recupera as contas de anúncios associadas ao token
3. Obtém insights e métricas das campanhas ou contas de anúncios
4. Processa os dados para calcular CTR, CAC e outras métricas

**Importante:** Se a API do Facebook não estiver disponível ou retornar um erro, a API irá automaticamente usar dados simulados como fallback para fins de demonstração.

### Endpoints

#### POST /api/meta-ads

Consulta dados do Meta Ads usando um token fornecido no corpo da requisição.

**Exemplo de requisição:**

```json
{
  "token": "seu_token_de_acesso"
}
```

**Exemplo de resposta:**

```json
{
  "status": "success",
  "message": "Dados do Meta Ads consultados com sucesso",
  "data": {
    "ctr": 3.2,
    "cac": 12.45,
    "investimento": 2500.75,
    "numero_vendas": 120,
    "campanha_id": "23847239847",
    "campanha_nome": "Campanha de Teste",
    "data_inicio": "2025-02-06",
    "data_fim": "2025-03-06"
  }
}
```

#### GET /api/meta-ads/metricas

Consulta métricas específicas do Meta Ads usando um token fornecido via query parameter.

**Parâmetros:**

- `token` (obrigatório): Token de acesso ao Meta Ads
- `account_id` (opcional): ID da conta de anúncios
- `campaign_id` (opcional): ID da campanha

**Exemplo de requisição:**

```
GET /api/meta-ads/metricas?token=seu_token_de_acesso&campaign_id=123456789
```

**Exemplo de resposta:**

```json
{
  "status": "success",
  "message": "Métricas do Meta Ads consultadas com sucesso",
  "data": {
    "ctr": 4.5,
    "cac": 9.75,
    "investimento": 1850.25,
    "numero_vendas": 95,
    "campanha_id": "123456789",
    "campanha_nome": "Campanha Específica",
    "data_inicio": "2025-02-06",
    "data_fim": "2025-03-06"
  }
}
```

## Estrutura do projeto

```
.
├── docs/               # Documentação Swagger
├── models/            # Modelos de dados
├── payloads/          # Exemplos de payloads
│   ├── hotmart/      # Payloads da Hotmart
│   ├── kiwify/       # Payloads da Kiwify
│   └── kirvano/      # Payloads da Kirvano
├── main.go           # Código principal
├── go.mod           # Dependências Go
└── README.md        # Este arquivo
```

## Contribuindo

1. Faça um fork do projeto
2. Crie uma branch para sua feature (`git checkout -b feature/nova-feature`)
3. Faça commit das suas alterações (`git commit -am 'Adiciona nova feature'`)
4. Faça push para a branch (`git push origin feature/nova-feature`)
5. Abra um Pull Request
