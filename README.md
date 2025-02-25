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
