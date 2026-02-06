package loaders

var (
	Empresa = TableSpec{
		Name: "empresa",
		Columns: []string{
			"cnpj_basico",
			"razao_social",
			"natureza_juridica",
			"qualificacao_responsavel",
			"capital_social",
			"porte_empresa",
			"ente_federativo_responsavel",
		},
	}
	Estabelecimento = TableSpec{
		Name: "estabelecimento",
		Columns: []string{
			"cnpj_basico","cnpj_ordem","cnpj_dv","identificador_matriz_filial","nome_fantasia","situacao_cadastral",
			"data_situacao_cadastral","motivo_situacao_cadastral","nome_cidade_exterior","pais","data_inicio_atividade",
			"cnae_fiscal_principal","cnae_fiscal_secundaria","tipo_logradouro","logradouro","numero","complemento","bairro",
			"cep","uf","municipio","ddd_1","telefone_1","ddd_2","telefone_2","ddd_fax","fax","correio_eletronico",
			"situacao_especial","data_situacao_especial",
		},
	}
	Socios = TableSpec{
		Name: "socios",
		Columns: []string{
			"cnpj_basico","identificador_socio","nome_socio_razao_social","cpf_cnpj_socio","qualificacao_socio",
			"data_entrada_sociedade","pais","representante_legal","nome_do_representante","qualificacao_representante_legal",
			"faixa_etaria",
		},
	}
	Simples = TableSpec{
		Name: "simples",
		Columns: []string{
			"cnpj_basico","opcao_pelo_simples","data_opcao_simples","data_exclusao_simples","opcao_mei","data_opcao_mei","data_exclusao_mei",
		},
	}
	Cnae = TableSpec{
		Name: "cnae",
		Columns: []string{"codigo","descricao"},
	}
	Moti = TableSpec{
		Name: "moti",
		Columns: []string{"codigo","descricao"},
	}
	Munic = TableSpec{
		Name: "munic",
		Columns: []string{"codigo","descricao"},
	}
	Natju = TableSpec{
		Name: "natju",
		Columns: []string{"codigo","descricao"},
	}
	Pais = TableSpec{
		Name: "pais",
		Columns: []string{"codigo","descricao"},
	}
	Quals = TableSpec{
		Name: "quals",
		Columns: []string{"codigo","descricao"},
	}
)
