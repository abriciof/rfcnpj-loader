# rfcnpj-loader (Go)

Carregador dos Dados Abertos do CNPJ (Receita Federal) para Postgres, com:
- descoberta de arquivos via WebDAV **PROPFIND** (XML)
- automação de mês (carrega sempre o próximo mês disponível)
- download/extract paralelos
- carga streaming via `COPY`
- notificação via SMTP (Gmail) ao final

## Como rodar (Docker)

1) Copie `.env.example` -> `.env` e ajuste as variáveis (DB e e-mail se quiser).
2) Execute:

```bash
docker compose up --build
```

## Como funciona a automação mensal

O loader salva no Postgres (tabela `rfcnpj_meta`) duas chaves:
- `loaded_month` (ex.: `2026-01`)
- `loaded_url` (URL usada no PROPFIND)

Em cada execução:
1. lê `loaded_month`
2. calcula o próximo mês
3. faz PROPFIND no template `DAV_LIST_URL_TEMPLATE` com `%s` = mês alvo
4. se existir (XML retornado), baixa/extrai/carreca
5. se não existir, finaliza informando que já está atualizado

Você pode forçar um mês com `FORCE_MONTH=YYYY-MM`.

## Switches equivalentes aos blocos comentados do Python

- `ENABLE_DOWNLOAD`: se `false`, **não baixa** (usa o que já estiver em `OUTPUT_FILES_PATH`)
- `ENABLE_EXTRACT`: se `false`, **não extrai** (usa o que já estiver em `EXTRACTED_FILES_PATH`)
- `CREATE_INDEXES`: se `true`, cria índices (cnpj_basico) nas principais tabelas
