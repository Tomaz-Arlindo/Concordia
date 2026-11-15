# Sprint 8 — Concordia

## Identificação da equipe

- **Grupo 17**

- **Integrantes:**
    - Maria Karolina Rodrigues de Andrade
    - Pedro Felipe Oliveira Tavares
    - Pedro Oliveira Barros Batista de Lima
    - Tomaz Arlindo Silva Ribeiro
    - Vladison Lucas Costa Dos Santos
- **Turma:** CC-8-MB
- **Nome do projeto:** Concordia

---

## 1. Todas as funcionalidades implementadas

O sistema reúne três módulos que funcionam de forma integrada: o do Pesquisador, que consulta o acervo; o do Curador, que forma e mantém esse acervo; e o do Administrador, que controla quem acessa o sistema e com qual perfil. Por cima deles há um histórico de operações alimentado por todos os fluxos.

### 1.1 Acesso ao sistema

O usuário se cadastra com nome, email e senha de no mínimo 8 caracteres, guardada com hash bcrypt, e entra com email e senha. Quando os dados estão errados, o sistema mostra apenas "Email ou senha inválidos.", sem indicar qual deles falhou. Todo cadastro público recebe o perfil Pesquisador.

Na primeira execução, se ainda não existir um Administrador, um pop-up obriga a criar um. Ele explica que o Administrador é quem gerencia os usuários e que a criação é obrigatória. Fechar o pop-up, pelo botão ou pela tecla Esc, abre uma confirmação e, se o usuário confirmar, o aplicativo é encerrado.

Na tela de login e na barra superior há um indicador do estado da conexão com o MySQL. Quando o banco está indisponível, um botão refaz a conexão e os serviços sem precisar reiniciar o aplicativo.

### 1.2 Módulo do Pesquisador

A **Visão geral** mostra quatro indicadores calculados no banco: total de artigos, artigos pendentes de indexação, artigos indexados e autores vinculados ao acervo.

A **Pesquisa** é o centro do módulo. O usuário digita uma consulta e o sistema devolve os artigos ordenados por relevância, calculada com BM25 sobre o índice invertido. Cada resultado mostra título, autores, fonte, data, a pontuação e os termos que casaram com a consulta. Com a busca aproximada ligada, um termo que não existe no índice é substituído pelo termo indexado mais próximo, medido pela distância de Levenshtein, de modo que uma consulta com erro de digitação ainda encontra resultados.

A pesquisa pode ser refinada por quatro filtros, combináveis entre si: fonte, autor, idioma e período de publicação (data inicial e final). Com o campo de busca vazio, a tela lista o acervo completo, e qualquer artigo pode ser aberto para ver seus metadados. Os artigos também podem ser exportados em PDF.

O **Histórico** permite consultar as operações registradas, filtrando por tipo e por período.

### 1.3 Módulo do Curador

O Curador forma o acervo importando PDFs. Na importação individual ele escolhe o arquivo e cadastra título, fonte, autores, descrição, idioma, data de publicação, DOI e identificador externo. Na importação em lote escolhe vários PDFs de uma vez, e o sistema usa o nome do arquivo como título, a fonte local e o idioma `pt-BR`; ao final exibe um relatório com os arquivos selecionados, importados, duplicados e com falha.

Os artigos importados podem ser listados, editados (metadados e autores, sem trocar o PDF) e excluídos, com confirmação.

A **Indexação** processa os artigos pendentes. O Curador define quantos workers usar e pode optar por reindexar também os documentos já indexados. A tela mostra os contadores do índice (pendentes, indexados, termos e ocorrências) e, ao final, o relatório da execução: quantos documentos foram indexados, quantas falhas ocorreram, quantos workers foram usados, a duração e o erro de cada documento que falhou.

A tela de **Métricas** reúne o total de artigos, termos, ocorrências e autores, e o benchmark de paralelismo, que executa a indexação com 1, 2, 4 e 8 workers e apresenta, para cada configuração, o tempo de processamento, o speedup, a eficiência, o tempo gasto no MySQL, o tempo total e o percentual de MySQL. Acima da tabela, dois destaques resumem o melhor speedup de processamento e a maior participação do MySQL no tempo total.

O Curador também pode exportar o histórico em CSV, respeitando os filtros aplicados na tela, em um formato que o Excel abre com os acentos corretos.

### 1.4 Módulo do Administrador

O Administrador executa tudo o que o Curador executa e, em exclusivo, gerencia os usuários. A tela Usuários lista id, nome, email, data de criação e função de cada conta, com uma busca em tempo real que procura em qualquer uma dessas colunas e ignora acentos. Na própria linha da tabela ele altera a função de um usuário entre Pesquisador e Curador, e pode excluir usuários; as duas ações pedem confirmação. Quando um usuário é excluído, os artigos que ele importou continuam no acervo.

As contas de Administrador são protegidas: não podem ter a função alterada nem ser excluídas, o que garante que o sistema nunca fique sem quem o administre.

### 1.5 Histórico de operações

O sistema registra automaticamente, no MySQL, as operações dos usuários: login, logout, cadastro, importação individual e em lote, edição e exclusão de artigo, indexação, benchmark, busca e exportação do histórico. Cada registro guarda o usuário da sessão, o tipo da operação, a entidade afetada, os detalhes, a quantidade de resultados (nas buscas) e a data e hora. Se a gravação do registro falhar, o erro vai apenas para o console e a operação principal continua normalmente.

A tela Histórico apresenta quatro indicadores (operações registradas, operações do dia, tipos de operação e registros exibidos), um relatório de operações por tipo, o ranking dos termos mais pesquisados e a tabela de registros.

### 1.6 Fluxo principal integrado

Os módulos se conectam em um único fluxo de uso:

```
Primeira execução
  → Criação do Administrador
  → Cadastro de usuário (perfil Pesquisador)
  → Administrador altera a função do usuário para Curador
  → Curador importa PDFs (individual ou em lote)   [artigos PENDING]
  → Curador executa a indexação paralela            [artigos INDEXED]
  → Pesquisador pesquisa, filtra e abre os artigos
  → Todas as operações aparecem no Histórico
  → Curador ou Administrador exporta o Histórico em CSV
```

### 1.7 Alterações de escopo

Algumas decisões mudaram em relação à ideia inicial do projeto:

- **Da linha de comando para um aplicativo desktop.** A proposta começou como uma ferramenta simples de linha de comando e evoluiu para um sistema completo, com dois perfis de usuário, gestão de acesso e acompanhamento. O núcleo de indexação e busca continua em Go, compilado e paralelo, como exigido pela disciplina.
- **De PostgreSQL para MySQL 8.** O PostgreSQL era uma recomendação ainda não fechada pelo grupo. O MySQL atende ao modelo relacional do projeto, com as tabelas `terms` e `term_occurrences` para o índice invertido, e é usado em todo o sistema.
- **Fontes arXiv, PubMed e OpenAlex.** As três fontes estão cadastradas e identificam a origem de cada artigo, mas os documentos entram no sistema por importação de PDF, e não por integração automática com as APIs. O foco ficou no núcleo técnico, a indexação paralela e a busca por relevância sobre o texto dos PDFs.
- **Sem OCR.** PDFs escaneados, sem camada de texto, não têm o conteúdo extraído: eles são marcados com erro na indexação, e a tela de Indexação avisa sobre essa limitação.

---

## 2. Controle de permissões

### 2.1 Perfis e como o controle funciona

O sistema tem três perfis. O Pesquisador consulta o acervo e é o perfil de todo cadastro público. O Curador forma e mantém o acervo e recebe esse perfil de um Administrador, pela tela Usuários. O Administrador é criado na primeira execução, pelo pop-up obrigatório, e acumula as permissões do Curador com a gestão de usuários.

O controle é feito em duas camadas. Na interface, o menu lateral e as telas aparecem conforme o perfil: o Pesquisador vê apenas Visão geral, Pesquisar e Histórico; o Curador acrescenta Artigos, Importar, Indexação e Métricas; o Administrador acrescenta Usuários. No backend, cada operação confere o usuário da sessão antes de executar. As operações de curadoria só aceitam Curador e Administrador, e as de gestão de usuários só aceitam o Administrador, de modo que uma chamada feita fora da tela também é recusada.

Três regras reforçam o controle. O formulário de cadastro não oferece escolha de perfil, e o cadastro grava sempre Pesquisador. A criação de um Administrador só é aceita quando não existe nenhum. E as contas de Administrador não podem ser alteradas nem excluídas.

### 2.2 Matriz de permissões

| Funcionalidade | Pesquisador | Curador | Administrador |
| --- | --- | --- | --- |
| Cadastro e login | Sim | Sim | Sim |
| Visão geral | Sim | Sim | Sim |
| Pesquisar, filtrar e abrir detalhes | Sim | Sim | Sim |
| Consultar o Histórico | Sim | Sim | Sim |
| Exportar o Histórico em CSV | Não | Sim | Sim |
| Tela Artigos (listar, editar e excluir) | Não | Sim | Sim |
| Importar artigos (individual e em lote) | Não | Sim | Sim |
| Executar a indexação | Não | Sim | Sim |
| Métricas e benchmark | Não | Sim | Sim |
| Gestão de usuários | Não | Não | Sim |
| Alterar ou excluir contas de Administrador | Não | Não | Não |

### 2.3 Exemplos de acessos permitidos e restritos

**Pesquisador.** Pode pesquisar com filtros, abrir os detalhes de um artigo e consultar o Histórico, onde o botão de exportação não aparece. Tudo o que é de curadoria é recusado pelo backend, mesmo que a chamada não venha pela tela, com mensagens como:

- importar: "Somente Curador pode importar artigos.";
- editar: "Somente Curador pode editar artigos.";
- excluir: "Somente Curador pode excluir artigos.";
- indexar: "Somente Curador pode iniciar a indexação.";
- benchmark: "Somente Curador pode executar benchmark.";
- exportar o histórico: "Somente Curador pode exportar o histórico."

**Curador.** Pode importar PDFs, editar e excluir artigos, indexar, ver as Métricas, executar o benchmark e exportar o Histórico em CSV. Por outro lado, o menu Usuários não aparece para ele e o backend recusa a gestão de usuários: "Somente o administrador pode alterar funções." e "Somente o administrador pode excluir usuários."; ao tentar listar usuários, recebe uma lista vazia.

**Administrador.** Pode tudo o que o Curador pode e, além disso, altera a função de um usuário entre Pesquisador e Curador (com a mensagem "Função de X alterada para Curador.") e exclui usuários, sempre após confirmação. Mesmo ele não consegue alterar ou excluir outro Administrador: a linha mostra "Protegido" e o backend responde "Administradores não podem ser alterados." ou "Administradores não podem ser excluídos."

**Casos gerais.** Tentar criar um segundo Administrador resulta em "Já existe um administrador cadastrado.", e usar qualquer operação sem login resulta em "Usuário não autenticado."

### 2.4 Roteiro de demonstração

1. Entrar como Pesquisador, conferir o menu (Visão geral, Pesquisar e Histórico), pesquisar com filtros e abrir o Histórico, que não tem o botão de exportação.
2. Entrar como Curador, conferir o menu de curadoria, importar um PDF, indexar e exportar o Histórico.
3. Entrar como Administrador e abrir a tela Usuários.
4. Alterar a função de um usuário de Pesquisador para Curador, confirmando no pop-up.
5. Entrar com esse usuário e conferir que os menus de curadoria passaram a aparecer.
6. Voltar ao Administrador e conferir que a conta de Administrador aparece como "Protegido".

---

## 3. Melhorias de usabilidade

### 3.1 Clareza e organização

Cada usuário vê somente o que pode usar: o menu lateral muda conforme o perfil, e o rodapé do menu mostra o nome e a função de quem está logado. Os números mais importantes de cada tela aparecem em cartões no topo, com uma legenda curta, o que permite ler a situação do acervo de relance.

Quando uma operação longa termina, o sistema mostra o que aconteceu. A indexação e a importação em lote apresentam um relatório com o que foi processado, as falhas e o erro de cada documento. O benchmark traz o resumo dos principais indicadores, a tabela detalhada e uma explicação do que cada etapa mede (processamento, MySQL e tempo total). No Histórico, os relatórios de operações por tipo e de termos mais pesquisados ficam ao lado da tabela de registros.

Todas as telas seguem o mesmo padrão visual, com os mesmos cartões, botões primários, secundários e de perigo, indicadores de status e janelas de confirmação.

### 3.2 Facilidade de uso

A tela de pesquisa traz um exemplo no campo de busca e agrupa fonte, autor, idioma, período e busca aproximada logo abaixo dele. Com a busca vazia, o acervo é listado por inteiro, e quando uma consulta não encontra nada a mensagem orienta o usuário a verificar se os artigos já foram indexados.

Os formulários de importação e edição marcam título e fonte como obrigatórios com asterisco, trazem exemplos nos campos de idioma, DOI e identificador externo, e explicam o campo de autores com a dica "Separe os autores por vírgulas." Ao escolher um PDF, o nome e o tamanho do arquivo aparecem antes do envio.

Durante as operações, os botões mudam de texto e ficam desabilitados ("Importando...", "Indexando...", "Executando...", "Criando..."), o que evita cliques repetidos e deixa claro que o sistema está trabalhando. As telas de consulta exibem mensagens de sucesso e de erro e textos próprios para listas vazias, como "Nenhum artigo encontrado." e "Nenhum usuário encontrado."

Ações que não podem ser desfeitas pedem confirmação: a exclusão de artigo, a exclusão de usuário e a alteração de função. A confirmação identifica o item afetado e, na troca de função, mostra a mudança de X para Y.

### 3.3 Administração e conexão

A tela Usuários tem busca em tempo real por id, nome, email, data ou função, e a função é alterada diretamente na linha da tabela. A conta em uso aparece com a marca "você", e as contas de Administrador aparecem como "Protegido", sem opções de alteração. O pop-up de criação do Administrador explica por que a criação é obrigatória e pede confirmação antes de encerrar o aplicativo. O indicador do banco, com o botão de reconexão, deixa visível um problema de conexão e o resolve sem reiniciar.

### 3.4 Navegação

A troca de telas passa por uma única função, que limpa mensagens antigas, fecha o painel de detalhes e atualiza a lista ao abrir Artigos. Depois de entrar, cadastrar-se ou criar o Administrador, o usuário cai na Visão geral. Depois de importar um artigo, o sistema abre a lista de Artigos já com o novo registro. Durante a edição de um artigo, o item "Artigos" continua marcado no menu, e salvar ou cancelar volta para a lista. Na tela de importação, a fonte padrão já vem selecionada. Ao reconectar com o banco, o sistema verifica de novo se existe Administrador e atualiza os dados da tela.

---

## 4. Revisão geral das regras de negócio

A revisão conferiu se o comportamento do sistema corresponde ao que foi definido para o projeto, área por área.

### 4.1 Regras revisadas

**Cadastro e acesso.** Nome, email e senha são obrigatórios, e a senha tem no mínimo 8 caracteres. O email é normalizado (minúsculas, sem espaços nas pontas) e é único. A senha é armazenada com hash bcrypt. O cadastro público grava sempre o perfil Pesquisador, e o sistema só é usado depois que existe um Administrador, criado uma única vez.

**Formação do acervo.** A importação exige título, fonte e arquivo PDF, e só aceita PDFs, verificados pela assinatura do conteúdo. O hash SHA-256 do arquivo é único, o que impede importar o mesmo PDF duas vezes mesmo com outro nome. Todo artigo importado nasce com a situação `PENDING` e fica vinculado ao usuário que o importou. Os autores são informados separados por vírgula: vazios são ignorados, repetidos são unificados e autores já cadastrados são reaproveitados. Na importação em lote, a falha em um arquivo não interrompe os demais. A edição altera metadados e autores, mas não substitui o PDF. A exclusão pede confirmação e remove o registro, o índice do documento, os vínculos com autores e o arquivo em disco. O cadastro de artigo e autores acontece em uma única transação, e se o banco falhar o arquivo recém-copiado é removido.

**Indexação.** Por padrão são indexados os artigos `PENDING` ou `ERROR`, com opção de reindexar também os já indexados. O número de workers vai de 1 até o dobro dos núcleos da máquina, com mínimo de 4 e teto de 32. Os workers leem o PDF, extraem o texto, normalizam e tokenizam, enquanto um único writer grava no MySQL. O índice de cada documento é gravado em uma transação, e a situação só passa para `INDEXED` ao final dela. Os termos são normalizados (minúsculas e sem acentos), e palavras comuns e termos com menos de 2 ou mais de 120 caracteres são descartados. PDFs sem texto ficam com erro, e o ciclo de situação de cada artigo é `PENDING`, `PROCESSING` e, depois, `INDEXED` ou `ERROR`.

**Pesquisa.** Apenas artigos indexados entram no ranking. A pontuação usa BM25 com `k1 = 1,2` e `b = 0,75`. A consulta é tratada da mesma forma que o texto indexado: minúsculas, sem acentos, sem palavras comuns e sem repetições. A busca aproximada só é aplicada a termos ausentes do índice, tolera até 2 edições (1 para termos de até 4 caracteres) e dá ao termo corrigido peso 0,72, para que nunca supere um acerto exato. Os filtros são cumulativos, ou seja, todos os critérios informados precisam ser atendidos, e artigos sem data de publicação ficam de fora quando há filtro de data. O limite é de 50 resultados por padrão e de 100 no máximo, e a busca vazia lista o acervo completo.

**Administração e histórico.** O Administrador só atribui as funções Pesquisador ou Curador. As contas de Administrador não podem ser alteradas nem excluídas. Ao excluir um usuário, os artigos que ele importou e os registros do histórico permanecem, sem responsável. As operações são registradas quando concluídas, com o usuário autenticado na sessão, e uma falha no registro nunca interrompe a operação principal. A exportação do histórico é restrita a Curador e Administrador, em CSV com separador `;` e codificação UTF-8 com BOM, respeitando os filtros da tela.

### 4.2 Inconsistências identificadas e corrigidas

- **O perfil Curador só podia ser obtido alterando o banco diretamente.** Agora a função é atribuída pelo Administrador na tela Usuários.
- **O sistema podia ser usado sem nenhum administrador.** Na primeira execução, o pop-up de criação do Administrador é obrigatório.
- **O Administrador não tinha as permissões de curadoria.** A validação de curadoria passou a aceitar os perfis Curador e Administrador.
- **Rebaixar ou excluir o Administrador deixaria o sistema sem gestão.** As contas de Administrador são protegidas no backend e também na consulta ao banco.
- **Uma falha de conexão na abertura exigia reiniciar o aplicativo.** O botão de reconexão refaz a conexão e os serviços sem reiniciar.
- **Deadlocks no MySQL (erro 1213) durante a indexação.** Várias transações concorrentes inseriam os mesmos termos na tabela `terms`. Hoje os workers apenas processam os documentos e um único writer grava no banco.
- **Chave primária duplicada em `term_occurrences`.** "Martinez" e "Martínez" eram termos distintos no Go, mas o mesmo registro no MySQL por causa da collation `utf8mb4_unicode_ci`. Os termos passaram a ser normalizados (minúsculas e remoção de acentos), com testes automatizados em `tokenizer_test.go`.

### 4.3 Coerência dos fluxos principais

**Na formação do acervo**, a validação do PDF, a checagem de duplicidade por hash, a situação inicial `PENDING` e a indexação com writer único levam cada artigo de `PENDING` a `INDEXED`, e só então ele fica disponível para a pesquisa.

**Na consulta**, a combinação de ranking apenas sobre artigos indexados, BM25, busca aproximada e filtros cumulativos entrega resultados ordenados por relevância, com a pontuação e os termos encontrados.

**Na gestão de acesso**, o cadastro sempre como Pesquisador, a função atribuída pelo Administrador e a proteção das contas de Administrador fazem com que cada usuário acesse apenas o que o seu perfil permite.

**Na rastreabilidade**, o registro automático das operações com o usuário da sessão, a consulta com filtros e a exportação restrita deixam as ações dos usuários registradas e consultáveis no Histórico.
