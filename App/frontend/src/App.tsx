import { FormEvent, useEffect, useMemo, useState } from 'react'
import './style.css'

import {
  DeleteArticle,
  GetArticle,
  GetCollectionStats,
  GetCurrentUser,
  Health,
  ImportArticle,
  ImportPDFBatch,
  ListArticles,
  ListSources,
  Login,
  Logout,
  Register,
  SearchBM25Filtered,
  RunIndexingBenchmark,
  SelectPDF,
  StartIndexing,
  GetIndexingOverview,
  UpdateArticle,
} from '../wailsjs/go/main/App'

import type {
  ArticleDetail,
  ArticleSource,
  ArticleSummary,
  AuthResponse,
  CollectionStats,
  HealthStatus,
  IndexingOverview,
  IndexingReport,
  BenchmarkReport,
  SearchResult,
  User,
} from './types'

type Screen =
  | 'dashboard'
  | 'search'
  | 'articles'
  | 'import'
  | 'indexing'
  | 'metrics'
  | 'edit'

const emptyStats: CollectionStats = {
  total_articles: 0,
  pending_articles: 0,
  indexed_articles: 0,
  total_authors: 0,
  total_sources: 0,
}

const emptyIndexingOverview: IndexingOverview = {
  pending: 0,
  processing: 0,
  indexed: 0,
  errors: 0,
  terms: 0,
  occurrences: 0,
  documents: 0,
}

const emptyArticleForm = {
  title: '',
  description: '',
  sourceId: 0,
  authors: '',
  languageCode: 'pt-BR',
  publicationDate: '',
  doi: '',
  externalId: '',
}


type BatchImportReport = {
  success: boolean
  message: string
  selected: number
  imported: number
  duplicates: number
  failed: number
  errors: string[]
}

function App() {
  const [mode, setMode] = useState<'login' | 'register'>('login')
  const [user, setUser] = useState<User | null>(null)
  const [health, setHealth] = useState<HealthStatus>({
    status: 'checking',
    database: 'checking',
  })
  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState('')
  const [authForm, setAuthForm] = useState({
    name: '',
    email: '',
    password: '',
  })

  const [screen, setScreen] = useState<Screen>('dashboard')
  const [articles, setArticles] = useState<ArticleSummary[]>([])
  const [sources, setSources] = useState<ArticleSource[]>([])
  const [stats, setStats] = useState<CollectionStats>(emptyStats)
  const [query, setQuery] = useState('')
  const [selectedFile, setSelectedFile] = useState<{ name: string; size: number } | null>(null)
  const [articleForm, setArticleForm] = useState(emptyArticleForm)
  const [editingArticle, setEditingArticle] = useState<ArticleDetail | null>(null)
  const [detailArticle, setDetailArticle] = useState<ArticleDetail | null>(null)
  const [workspaceMessage, setWorkspaceMessage] = useState('')
  const [indexingOverview, setIndexingOverview] = useState<IndexingOverview>(emptyIndexingOverview)
  const [indexingReport, setIndexingReport] = useState<IndexingReport | null>(null)
  const [workerCount, setWorkerCount] = useState(4)
  const [reindexAll, setReindexAll] = useState(false)
  const [indexingLoading, setIndexingLoading] = useState(false)
  const [searchMode, setSearchMode] = useState<'bm25' | 'metadata'>('bm25')
  const [searchFilters, setSearchFilters] = useState({
    source: '',
    author: '',
    language: '',
    dateFrom: '',
    dateTo: '',
    fuzzy: true,
  })
  const [benchmarkReport, setBenchmarkReport] = useState<BenchmarkReport | null>(null)
  const [benchmarkLoading, setBenchmarkLoading] = useState(false)
  const [batchLoading, setBatchLoading] = useState(false)
  const [batchReport, setBatchReport] = useState<BatchImportReport | null>(null)

  useEffect(() => {
    Health()
      .then(setHealth)
      .catch(() =>
        setHealth({ status: 'error', database: 'disconnected' }),
      )

    GetCurrentUser()
      .then((current) => {
        if (current) setUser(current as User)
      })
      .catch(() => undefined)
  }, [])

  useEffect(() => {
    if (!user) return
    refreshWorkspace()
  }, [user])

  const databaseOnline = health.database === 'connected'
  const isCurator = user?.role === 'CURADOR'

  const roleLabel = useMemo(() => {
    if (!user) return ''
    return isCurator ? 'Curador' : 'Pesquisador'
  }, [user, isCurator])

  async function refreshWorkspace() {
    try {
      const [loadedArticles, loadedSources, loadedStats, loadedIndexing] = await Promise.all([
        ListArticles(),
        ListSources(),
        GetCollectionStats(),
        GetIndexingOverview(),
      ])
      setArticles((loadedArticles || []) as ArticleSummary[])
      setSources((loadedSources || []) as ArticleSource[])
      setStats((loadedStats || emptyStats) as CollectionStats)
      setIndexingOverview((loadedIndexing || emptyIndexingOverview) as IndexingOverview)
    } catch {
      setWorkspaceMessage('Não foi possível atualizar os dados do projeto.')
    }
  }

  async function handleAuthSubmit(event: FormEvent) {
    event.preventDefault()
    setMessage('')
    setLoading(true)

    try {
      let response: AuthResponse

      if (mode === 'register') {
        response = await Register(
          authForm.name,
          authForm.email,
          authForm.password,
        )
      } else {
        response = await Login(authForm.email, authForm.password)
      }

      setMessage(response.message)

      if (response.success && response.user) {
        setUser(response.user)
        setAuthForm({ name: '', email: '', password: '' })
        setScreen('dashboard')
      }
    } catch {
      setMessage('Ocorreu um erro ao comunicar com o backend Go.')
    } finally {
      setLoading(false)
    }
  }

  async function handleLogout() {
    await Logout()
    setUser(null)
    setScreen('dashboard')
    setMessage('')
    setWorkspaceMessage('')
    setDetailArticle(null)
  }

  async function handleSearch() {
    setWorkspaceMessage('')
    try {
      if (!query.trim()) {
        const result = await ListArticles()
        setArticles((result || []) as ArticleSummary[])
        setSearchMode('metadata')
        return
      }

      const result = await SearchBM25Filtered(
        query,
        searchFilters.source,
        searchFilters.author,
        searchFilters.language,
        searchFilters.dateFrom,
        searchFilters.dateTo,
        searchFilters.fuzzy,
        50,
      )
      const ranked = (result || []) as SearchResult[]
      setArticles(ranked.map((item) => ({ ...item, created_at: '' })))
      setSearchMode('bm25')
      if (ranked.length === 0) {
        setWorkspaceMessage('Nenhum resultado BM25 encontrado. Verifique se os artigos já foram indexados.')
      }
    } catch {
      setWorkspaceMessage('Falha ao pesquisar artigos pelo índice invertido.')
    }
  }

  async function handleBenchmark() {
    setBenchmarkLoading(true)
    setWorkspaceMessage('')
    setBenchmarkReport(null)

    try {
      const response = await RunIndexingBenchmark()
      setWorkspaceMessage(response.message)
      if (response.success && response.report) {
        setBenchmarkReport(response.report as BenchmarkReport)
      }
      await refreshWorkspace()
    } catch {
      setWorkspaceMessage('Falha ao executar o benchmark de indexação.')
    } finally {
      setBenchmarkLoading(false)
    }
  }

  async function handleIndexing() {
    setIndexingLoading(true)
    setWorkspaceMessage('')
    setIndexingReport(null)

    try {
      const response = await StartIndexing(workerCount, reindexAll)
      setWorkspaceMessage(response.message)
      if (response.report) {
        setIndexingReport(response.report as IndexingReport)
      }
      await refreshWorkspace()
    } catch {
      setWorkspaceMessage('Falha ao executar a indexação paralela.')
    } finally {
      setIndexingLoading(false)
    }
  }

  async function openArticle(id: number) {
    const response = await GetArticle(id)
    if (response.success && response.article) {
      setDetailArticle(response.article as ArticleDetail)
    } else {
      setWorkspaceMessage(response.message || 'Não foi possível abrir o artigo.')
    }
  }

  async function selectPDF() {
    setWorkspaceMessage('')
    const response = await SelectPDF()
    if (response.success) {
      setSelectedFile({ name: response.name ?? 'arquivo.pdf', size: response.size ?? 0 })
    } else if (response.message !== 'Nenhum arquivo selecionado.') {
      setWorkspaceMessage(response.message)
    }
  }

  async function submitImport(event: FormEvent) {
    event.preventDefault()
    setLoading(true)
    setWorkspaceMessage('')

    try {
      const response = await ImportArticle(
        articleForm.title,
        articleForm.description,
        Number(articleForm.sourceId),
        articleForm.authors,
        articleForm.languageCode,
        articleForm.publicationDate,
        articleForm.doi,
        articleForm.externalId,
      )

      setWorkspaceMessage(response.message)

      if (response.success) {
        setArticleForm({
          ...emptyArticleForm,
          sourceId: sources[0]?.id || 0,
        })
        setSelectedFile(null)
        await refreshWorkspace()
        setScreen('articles')
      }
    } finally {
      setLoading(false)
    }
  }

  async function startEdit(article: ArticleDetail) {
    setEditingArticle(article)
    setArticleForm({
      title: article.title || '',
      description: article.description || '',
      sourceId: article.source_id || 0,
      authors: article.authors || '',
      languageCode: article.language_code || '',
      publicationDate: article.publication_date || '',
      doi: article.doi || '',
      externalId: article.external_id || '',
    })
    setDetailArticle(null)
    setScreen('edit')
  }

  async function submitEdit(event: FormEvent) {
    event.preventDefault()
    if (!editingArticle) return

    setLoading(true)
    setWorkspaceMessage('')

    try {
      const response = await UpdateArticle(
        editingArticle.id,
        articleForm.title,
        articleForm.description,
        Number(articleForm.sourceId),
        articleForm.authors,
        articleForm.languageCode,
        articleForm.publicationDate,
        articleForm.doi,
        articleForm.externalId,
      )
      setWorkspaceMessage(response.message)

      if (response.success) {
        setEditingArticle(null)
        await refreshWorkspace()
        setScreen('articles')
      }
    } finally {
      setLoading(false)
    }
  }

  async function deleteArticle(id: number) {
    const accepted = window.confirm(
      'Tem certeza que deseja excluir este artigo? O PDF armazenado também será removido.',
    )
    if (!accepted) return

    const response = await DeleteArticle(id)
    setWorkspaceMessage(response.message)

    if (response.success) {
      setDetailArticle(null)
      await refreshWorkspace()
    }
  }


  async function importBatchForBenchmark() {
    const localSource =
      sources.find((source) => source.name.toLowerCase() === 'local') ||
      sources[0]

    if (!localSource) {
      setWorkspaceMessage('Nenhuma fonte está disponível para a importação em lote.')
      return
    }

    setBatchLoading(true)
    setBatchReport(null)
    setWorkspaceMessage('')

    try {
      const response = await ImportPDFBatch(
        localSource.id,
        articleForm.languageCode || 'pt-BR',
      )

      const normalized: BatchImportReport = {
        success: response.success ?? false,
        message: response.message ?? '',
        selected: response.selected ?? 0,
        imported: response.imported ?? 0,
        duplicates: response.duplicates ?? 0,
        failed: response.failed ?? 0,
        errors: response.errors ?? [],
      }

      setBatchReport(normalized)

      if (normalized.imported > 0) {
        await refreshWorkspace()
      }
    } catch {
      setWorkspaceMessage('Não foi possível concluir a importação em lote.')
    } finally {
      setBatchLoading(false)
    }
  }

  function navigate(next: Screen) {
    setWorkspaceMessage('')
    setDetailArticle(null)
    setScreen(next)

    if (next === 'import' && articleForm.sourceId === 0 && sources.length > 0) {
      setArticleForm((current) => ({
        ...current,
        sourceId: sources[0].id,
      }))
    }

    if (next === 'articles') {
      refreshWorkspace()
    }
  }

  if (!user) {
    return (
      <main className="auth-layout">
        <section className="presentation">
          <div className="presentation-inner">
            <div className="brand large">
              <div className="brand-mark">C</div>
              <div>
                <strong>Concordia</strong>
                <span>Motor de Busca Científica</span>
              </div>
            </div>

            <span className="eyebrow light">BUSCA CIENTÍFICA PARALELA</span>
            <h1>
              Informação acadêmica,
              <br />
              encontrada com eficiência.
            </h1>
            <p>
              Indexação paralela, índice invertido e busca por relevância em
              uma aplicação desktop construída com Go.
            </p>

            <div className="tech-list">
              <span>Wails</span>
              <span>Go</span>
              <span>React</span>
              <span>MySQL</span>
            </div>
          </div>
        </section>

        <section className="auth-panel">
          <div className="auth-card">
            <div className="status-row">
              <span
                className={`database-pill ${databaseOnline ? 'online' : 'offline'}`}
              >
                <span />
                {databaseOnline ? 'Banco conectado' : 'Banco desconectado'}
              </span>
            </div>

            <span className="eyebrow">BEM-VINDO</span>
            <h2>{mode === 'login' ? 'Acesse o Concordia' : 'Crie sua conta'}</h2>
            <p className="muted">
              {mode === 'login'
                ? 'Entre com suas credenciais para continuar.'
                : 'Novos cadastros recebem o perfil Pesquisador.'}
            </p>

            <form onSubmit={handleAuthSubmit}>
              {mode === 'register' && (
                <label>
                  Nome
                  <input
                    value={authForm.name}
                    onChange={(event) =>
                      setAuthForm({ ...authForm, name: event.target.value })
                    }
                    placeholder="Seu nome"
                    autoComplete="name"
                  />
                </label>
              )}

              <label>
                Email
                <input
                  type="email"
                  value={authForm.email}
                  onChange={(event) =>
                    setAuthForm({ ...authForm, email: event.target.value })
                  }
                  placeholder="nome@email.com"
                  autoComplete="email"
                />
              </label>

              <label>
                Senha
                <input
                  type="password"
                  value={authForm.password}
                  onChange={(event) =>
                    setAuthForm({ ...authForm, password: event.target.value })
                  }
                  placeholder="Mínimo de 8 caracteres"
                  autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
                />
              </label>

              {message && <div className="message">{message}</div>}

              <button className="primary" disabled={loading || !databaseOnline}>
                {loading ? 'Aguarde...' : mode === 'login' ? 'Entrar' : 'Criar conta'}
              </button>
            </form>

            <button
              className="switch-mode"
              onClick={() => {
                setMode(mode === 'login' ? 'register' : 'login')
                setMessage('')
              }}
            >
              {mode === 'login'
                ? 'Não possui conta? Criar cadastro'
                : 'Já possui conta? Fazer login'}
            </button>
          </div>
        </section>
      </main>
    )
  }

  return (
    <main className="app-shell">
      <aside className="sidebar">
        <div className="brand">
          <div className="brand-mark">C</div>
          <div>
            <strong>Concordia</strong>
            <span>Busca científica</span>
          </div>
        </div>

        <nav>
          <button
            className={`nav-item ${screen === 'dashboard' ? 'active' : ''}`}
            onClick={() => navigate('dashboard')}
          >
            Visão geral
          </button>
          <button
            className={`nav-item ${screen === 'search' ? 'active' : ''}`}
            onClick={() => navigate('search')}
          >
            Pesquisar
          </button>

          {isCurator && (
            <>
              <button
                className={`nav-item ${screen === 'articles' || screen === 'edit' ? 'active' : ''}`}
                onClick={() => navigate('articles')}
              >
                Artigos
              </button>
              <button
                className={`nav-item ${screen === 'import' ? 'active' : ''}`}
                onClick={() => navigate('import')}
              >
                Importar
              </button>
              <button
                className={`nav-item ${screen === 'indexing' ? 'active' : ''}`}
                onClick={() => navigate('indexing')}
              >
                Indexação
              </button>
              <button
                className={`nav-item ${screen === 'metrics' ? 'active' : ''}`}
                onClick={() => navigate('metrics')}
              >
                Métricas
              </button>
            </>
          )}
        </nav>

        <div className="sidebar-user">
          <strong>{user.name}</strong>
          <span>{roleLabel}</span>
        </div>

        <button className="logout" onClick={handleLogout}>
          Sair
        </button>
      </aside>

      <section className="content">
        <header className="topbar">
          <div>
            <span className="eyebrow">CONCORDIA</span>
            <h1>{screenTitle(screen)}</h1>
            <p>{screenSubtitle(screen, roleLabel)}</p>
          </div>

          <div className={`database-pill ${databaseOnline ? 'online' : 'offline'}`}>
            <span />
            MySQL {databaseOnline ? 'conectado' : 'desconectado'}
          </div>
        </header>

        {workspaceMessage && <div className="workspace-message">{workspaceMessage}</div>}

        {screen === 'dashboard' && (
          <>
            <div className="cards">
              <article className="metric-card">
                <span>Artigos</span>
                <strong>{stats.total_articles}</strong>
                <small>Documentos cadastrados</small>
              </article>
              <article className="metric-card">
                <span>Pendentes</span>
                <strong>{stats.pending_articles}</strong>
                <small>Aguardando indexação</small>
              </article>
              <article className="metric-card">
                <span>Indexados</span>
                <strong>{stats.indexed_articles}</strong>
                <small>Disponíveis para ranking</small>
              </article>
              <article className="metric-card">
                <span>Autores</span>
                <strong>{stats.total_authors}</strong>
                <small>Autores vinculados ao acervo</small>
              </article>
            </div>

            <section className="hero-panel">
              <span className="eyebrow">SPRINT 2 — VERSÃO FUNCIONAL</span>
              <h2>Motor de busca paralelo para artigos científicos</h2>
              <p>
                Versão funcional da Sprint 2 com curadoria, persistência, indexação paralela, índice
                invertido, filtros e ranking BM25 sobre o conteúdo extraído dos PDFs.
              </p>
              <div className="architecture">
                <span>React + TypeScript</span>
                <b>→</b>
                <span>Wails</span>
                <b>→</b>
                <span>Go</span>
                <b>→</b>
                <span>MySQL</span>
              </div>
            </section>
          </>
        )}

        {screen === 'search' && (
          <section className="page-card">
            <div className="section-heading">
              <div>
                <h2>Pesquisar artigos</h2>
                <p>Ranking real por relevância usando índice invertido + BM25.</p>
              </div>
            </div>

            <div className="search-row">
              <input
                value={query}
                onChange={(event) => setQuery(event.target.value)}
                placeholder="Ex.: inteligência artificial, João Silva, arxiv..."
                onKeyDown={(event) => {
                  if (event.key === 'Enter') handleSearch()
                }}
              />
              <button className="primary compact" onClick={handleSearch}>
                Pesquisar
              </button>
            </div>

            <div className="search-filters">
              <label>
                Fonte
                <select
                  value={searchFilters.source}
                  onChange={(event) =>
                    setSearchFilters({ ...searchFilters, source: event.target.value })
                  }
                >
                  <option value="">Todas</option>
                  {sources.map((source) => (
                    <option key={source.id} value={source.name}>
                      {source.name}
                    </option>
                  ))}
                </select>
              </label>

              <label>
                Autor
                <input
                  value={searchFilters.author}
                  onChange={(event) =>
                    setSearchFilters({ ...searchFilters, author: event.target.value })
                  }
                  placeholder="Nome do autor"
                />
              </label>

              <label>
                Idioma
                <input
                  value={searchFilters.language}
                  onChange={(event) =>
                    setSearchFilters({ ...searchFilters, language: event.target.value })
                  }
                  placeholder="pt-BR"
                />
              </label>

              <label>
                Data inicial
                <input
                  type="date"
                  value={searchFilters.dateFrom}
                  onChange={(event) =>
                    setSearchFilters({ ...searchFilters, dateFrom: event.target.value })
                  }
                />
              </label>

              <label>
                Data final
                <input
                  type="date"
                  value={searchFilters.dateTo}
                  onChange={(event) =>
                    setSearchFilters({ ...searchFilters, dateTo: event.target.value })
                  }
                />
              </label>

              <label className="checkbox-label filter-checkbox">
                <input
                  type="checkbox"
                  checked={searchFilters.fuzzy}
                  onChange={(event) =>
                    setSearchFilters({ ...searchFilters, fuzzy: event.target.checked })
                  }
                />
                Busca aproximada
              </label>
            </div>

            <div className="notice">
              {searchMode === 'bm25'
                ? 'Resultados ordenados por BM25. A busca aproximada tolera pequenos erros de digitação nos termos.'
                : 'Listagem completa do acervo.'}
            </div>

            <ArticleList
              articles={articles}
              isCurator={false}
              onOpen={openArticle}
            />
          </section>
        )}

        {screen === 'articles' && isCurator && (
          <section className="page-card">
            <div className="section-heading">
              <div>
                <h2>Acervo de artigos</h2>
                <p>Gerencie os documentos persistidos no Concordia.</p>
              </div>
              <button className="secondary" onClick={() => navigate('import')}>
                Importar PDF
              </button>
            </div>

            <ArticleList
              articles={articles}
              isCurator
              onOpen={openArticle}
              onDelete={deleteArticle}
            />
          </section>
        )}

        {screen === 'import' && isCurator && (
          <>
            <ArticleForm
              title="Importar novo artigo"
              description="O PDF será copiado para storage/articles e os metadados serão persistidos no MySQL."
              form={articleForm}
              setForm={setArticleForm}
              sources={sources}
              selectedFile={selectedFile}
              onSelectPDF={selectPDF}
              onSubmit={submitImport}
              loading={loading}
              submitLabel="Importar artigo"
            />

            <section className="page-card batch-import-card">
              <div className="section-heading">
                <div>
                  <span className="eyebrow">DATASET PARA BENCHMARK</span>
                  <h2>Importação em lote</h2>
                  <p>
                    Selecione vários PDFs de uma vez. O nome de cada arquivo será usado
                    como título, a fonte será <strong>local</strong> e o idioma padrão será
                    <strong> {articleForm.languageCode || 'pt-BR'}</strong>.
                  </p>
                </div>

                <button
                  className="primary compact"
                  type="button"
                  disabled={batchLoading}
                  onClick={importBatchForBenchmark}
                >
                  {batchLoading ? 'Importando...' : 'Selecionar vários PDFs'}
                </button>
              </div>

              <div className="notice">
                Use esta opção apenas para montar rapidamente um acervo de teste para o
                benchmark da Sprint 2. Para artigos que serão demonstrados individualmente,
                prefira a importação completa acima e preencha os metadados.
              </div>

              {batchReport && (
                <div className="batch-report">
                  <div className="batch-report-grid">
                    <article>
                      <span>Selecionados</span>
                      <strong>{batchReport.selected}</strong>
                    </article>
                    <article>
                      <span>Importados</span>
                      <strong>{batchReport.imported}</strong>
                    </article>
                    <article>
                      <span>Duplicados</span>
                      <strong>{batchReport.duplicates}</strong>
                    </article>
                    <article>
                      <span>Falhas</span>
                      <strong>{batchReport.failed}</strong>
                    </article>
                  </div>

                  <p>{batchReport.message}</p>

                  {batchReport.errors.length > 0 && (
                    <div className="batch-errors">
                      <strong>Arquivos com falha</strong>
                      {batchReport.errors.map((item, index) => (
                        <span key={`${item}-${index}`}>{item}</span>
                      ))}
                    </div>
                  )}
                </div>
              )}
            </section>
          </>
        )}

        {screen === 'edit' && isCurator && editingArticle && (
          <ArticleForm
            title={`Editar artigo #${editingArticle.id}`}
            description="Atualize metadados, fonte e autores. O arquivo PDF não será substituído."
            form={articleForm}
            setForm={setArticleForm}
            sources={sources}
            onSubmit={submitEdit}
            loading={loading}
            submitLabel="Salvar alterações"
            onCancel={() => navigate('articles')}
          />
        )}

        {screen === 'indexing' && isCurator && (
          <section className="page-card">
            <div className="section-heading">
              <div>
                <span className="eyebrow">GOROUTINES + WORKER POOL</span>
                <h2>Indexação paralela</h2>
                <p>
                  Extrai texto dos PDFs, tokeniza em workers concorrentes e persiste
                  o índice invertido no MySQL.
                </p>
              </div>
            </div>

            <div className="cards indexing-cards">
              <article className="metric-card">
                <span>Pendentes</span>
                <strong>{indexingOverview.pending}</strong>
                <small>Status PENDING</small>
              </article>
              <article className="metric-card">
                <span>Indexados</span>
                <strong>{indexingOverview.indexed}</strong>
                <small>Documentos processados</small>
              </article>
              <article className="metric-card">
                <span>Termos</span>
                <strong>{indexingOverview.terms}</strong>
                <small>Vocabulário do índice</small>
              </article>
              <article className="metric-card">
                <span>Ocorrências</span>
                <strong>{indexingOverview.occurrences}</strong>
                <small>Postings persistidos</small>
              </article>
            </div>

            <div className="indexing-control">
              <label>
                Workers / goroutines
                <input
                  type="number"
                  min={1}
                  max={32}
                  value={workerCount}
                  onChange={(event) =>
                    setWorkerCount(Math.max(1, Math.min(32, Number(event.target.value) || 1)))
                  }
                />
              </label>

              <label className="checkbox-label">
                <input
                  type="checkbox"
                  checked={reindexAll}
                  onChange={(event) => setReindexAll(event.target.checked)}
                />
                Reindexar também documentos já indexados
              </label>

              <button
                className="primary compact"
                onClick={handleIndexing}
                disabled={indexingLoading}
              >
                {indexingLoading ? 'Indexando...' : 'Iniciar indexação'}
              </button>
            </div>

            <div className="architecture indexing-flow">
              <span>PDF</span><b>→</b><span>extração</span><b>→</b>
              <span>workers</span><b>→</b><span>tokens</span><b>→</b>
              <span>índice invertido</span><b>→</b><span>MySQL</span>
            </div>

            {indexingReport && (
              <div className="indexing-report">
                <div>
                  <strong>{indexingReport.indexed}</strong> indexados
                </div>
                <div>
                  <strong>{indexingReport.failed}</strong> falhas
                </div>
                <div>
                  <strong>{indexingReport.workers}</strong> workers
                </div>
                <div>
                  <strong>{indexingReport.duration_ms} ms</strong> duração
                </div>

                {indexingReport.errors?.length > 0 && (
                  <div className="indexing-errors">
                    <strong>Erros encontrados</strong>
                    {indexingReport.errors.map((error, index) => (
                      <span key={`${error}-${index}`}>{error}</span>
                    ))}
                  </div>
                )}
              </div>
            )}

            <div className="notice">
              PDFs escaneados sem camada de texto podem retornar erro de extração.
              O bloco atual não usa OCR.
            </div>
          </section>
        )}

        {screen === 'metrics' && isCurator && (
          <section className="page-card">
            <div className="section-heading">
              <div>
                <h2>Métricas do acervo</h2>
                <p>Indicadores do acervo e do processamento persistidos no MySQL.</p>
              </div>
            </div>

            <div className="cards">
              <article className="metric-card">
                <span>Total de artigos</span>
                <strong>{stats.total_articles}</strong>
                <small>Acervo persistido</small>
              </article>
              <article className="metric-card">
                <span>Termos</span>
                <strong>{indexingOverview.terms}</strong>
                <small>Vocabulário invertido</small>
              </article>
              <article className="metric-card">
                <span>Ocorrências</span>
                <strong>{indexingOverview.occurrences}</strong>
                <small>Postings persistidos</small>
              </article>
              <article className="metric-card">
                <span>Autores</span>
                <strong>{stats.total_authors}</strong>
                <small>Autores vinculados a artigos</small>
              </article>
            </div>

            <div className="benchmark-panel">
              <div className="section-heading">
                <div>
                  <h3>Benchmark de paralelismo</h3>
                  <p>Separa processamento paralelo, persistência MySQL e tempo fim-a-fim.</p>
                </div>
                <button
                  className="primary compact"
                  onClick={handleBenchmark}
                  disabled={benchmarkLoading}
                >
                  {benchmarkLoading ? 'Executando...' : 'Executar benchmark'}
                </button>
              </div>

              {benchmarkReport ? (
                <>
                  <div className="benchmark-summary-grid">
                    <article>
                      <span>Melhor speedup de processamento</span>
                      <strong>{Math.max(...benchmarkReport.results.map((item) => item.processing_speedup || 0)).toFixed(2)}x</strong>
                      <small>Etapa onde as goroutines atuam</small>
                    </article>
                    <article>
                      <span>Maior participação do MySQL</span>
                      <strong>{(Math.max(...benchmarkReport.results.map((item) => item.database_share || 0)) * 100).toFixed(1)}%</strong>
                      <small>Parcela serial do tempo total</small>
                    </article>
                  </div>

                  <div className="benchmark-table benchmark-table-detailed">
                    <div className="benchmark-head">
                      <span>Workers</span>
                      <span>Processamento</span>
                      <span>Speedup proc.</span>
                      <span>Eficiência proc.</span>
                      <span>MySQL</span>
                      <span>Total</span>
                      <span>Speedup total</span>
                      <span>% MySQL</span>
                      <span>Falhas</span>
                    </div>
                    {benchmarkReport.results.map((item) => (
                      <div className="benchmark-row" key={item.workers}>
                        <strong>{item.workers}</strong>
                        <span>{item.processing_ms} ms</span>
                        <span>{item.processing_speedup ? `${item.processing_speedup.toFixed(2)}x` : '—'}</span>
                        <span>{item.processing_efficiency ? `${(item.processing_efficiency * 100).toFixed(1)}%` : '—'}</span>
                        <span>{item.database_ms} ms</span>
                        <span>{item.duration_ms} ms</span>
                        <span>{item.total_speedup ? `${item.total_speedup.toFixed(2)}x` : '—'}</span>
                        <span>{item.database_share ? `${(item.database_share * 100).toFixed(1)}%` : '0.0%'}</span>
                        <span>{item.failed}</span>
                      </div>
                    ))}
                  </div>

                  <div className="benchmark-explanation">
                    <div>
                      <strong>Processamento</strong>
                      <span>Leitura do PDF, extração, normalização, tokenização e índice local.</span>
                    </div>
                    <div>
                      <strong>MySQL</strong>
                      <span>Persistência de termos/postings e recálculo de document frequency.</span>
                    </div>
                    <div>
                      <strong>Tempo total</strong>
                      <span>Soma das duas etapas medidas pelo benchmark acadêmico.</span>
                    </div>
                  </div>
                  <div className="notice">{benchmarkReport.note}</div>
                </>
              ) : (
                <div className="empty-benchmark">
                  Execute o benchmark com vários PDFs para demonstrar o efeito do paralelismo.
                </div>
              )}
            </div>
          </section>
        )}
      </section>

      {detailArticle && (
        <div className="modal-backdrop" onClick={() => setDetailArticle(null)}>
          <article className="detail-modal" onClick={(event) => event.stopPropagation()}>
            <div className="detail-header">
              <div>
                <span className="eyebrow">ARTIGO #{detailArticle.id}</span>
                <h2>{detailArticle.title}</h2>
              </div>
              <button className="icon-button" onClick={() => setDetailArticle(null)}>
                ×
              </button>
            </div>

            <div className="detail-grid">
              <Detail label="Autores" value={detailArticle.authors || 'Não informado'} />
              <Detail label="Fonte" value={detailArticle.source_name || 'Não informada'} />
              <Detail label="Idioma" value={detailArticle.language_code || 'Não informado'} />
              <Detail label="Publicação" value={detailArticle.publication_date || 'Não informada'} />
              <Detail label="Status" value={detailArticle.index_status} />
              <Detail label="Arquivo" value={detailArticle.filename || 'Sem arquivo'} />
              <Detail label="DOI" value={detailArticle.doi || 'Não informado'} />
              <Detail label="External ID" value={detailArticle.external_id || 'Não informado'} />
            </div>

            <div className="description-box">
              <strong>Descrição</strong>
              <p>{detailArticle.description || 'Nenhuma descrição cadastrada.'}</p>
            </div>

            {isCurator && (
              <div className="modal-actions">
                <button className="secondary" onClick={() => startEdit(detailArticle)}>
                  Editar metadados
                </button>
                <button className="danger" onClick={() => deleteArticle(detailArticle.id)}>
                  Excluir artigo
                </button>
              </div>
            )}
          </article>
        </div>
      )}
    </main>
  )
}

function ArticleList({
  articles,
  isCurator,
  onOpen,
  onDelete,
}: {
  articles: ArticleSummary[]
  isCurator: boolean
  onOpen: (id: number) => void
  onDelete?: (id: number) => void
}) {
  if (articles.length === 0) {
    return (
      <div className="empty-state">
        <strong>Nenhum artigo encontrado.</strong>
        <span>Importe um PDF ou ajuste sua pesquisa.</span>
      </div>
    )
  }

  return (
    <div className="article-list">
      {articles.map((article) => (
        <article className="article-row" key={article.id}>
          <div className="article-main">
            <div className="article-badges">
              <span className="badge">{article.source_name || 'sem fonte'}</span>
              <span className={`status-badge ${article.index_status.toLowerCase()}`}>
                {article.index_status}
              </span>
              {'score' in article && typeof (article as any).score === 'number' && (
                <span className="score-badge">BM25 {(article as any).score.toFixed(3)}</span>
              )}
            </div>
            <h3>{article.title}</h3>
            <p>
              {article.authors || 'Autores não informados'}
              {article.publication_date ? ` • ${article.publication_date}` : ''}
            </p>
            {'matched_terms' in article && (article as any).matched_terms && (
              <small className="matched-terms">
                Termos: {(article as any).matched_terms}
              </small>
            )}
          </div>

          <div className="article-actions">
            <button className="secondary small" onClick={() => onOpen(article.id)}>
              Detalhes
            </button>
            {isCurator && onDelete && (
              <button className="danger ghost small" onClick={() => onDelete(article.id)}>
                Excluir
              </button>
            )}
          </div>
        </article>
      ))}
    </div>
  )
}

function ArticleForm({
  title,
  description,
  form,
  setForm,
  sources,
  selectedFile,
  onSelectPDF,
  onSubmit,
  loading,
  submitLabel,
  onCancel,
}: {
  title: string
  description: string
  form: typeof emptyArticleForm
  setForm: React.Dispatch<React.SetStateAction<typeof emptyArticleForm>>
  sources: ArticleSource[]
  selectedFile?: { name: string; size: number } | null
  onSelectPDF?: () => void
  onSubmit: (event: FormEvent) => void
  loading: boolean
  submitLabel: string
  onCancel?: () => void
}) {
  return (
    <section className="page-card">
      <div className="section-heading">
        <div>
          <h2>{title}</h2>
          <p>{description}</p>
        </div>
      </div>

      <form className="article-form" onSubmit={onSubmit}>
        {onSelectPDF && (
          <div className="file-picker">
            <div>
              <strong>{selectedFile?.name || 'Nenhum PDF selecionado'}</strong>
              <span>
                {selectedFile
                  ? formatBytes(selectedFile.size)
                  : 'Selecione um arquivo PDF local'}
              </span>
            </div>
            <button className="secondary" type="button" onClick={onSelectPDF}>
              Selecionar PDF
            </button>
          </div>
        )}

        <div className="form-grid two">
          <label>
            Título *
            <input
              value={form.title}
              onChange={(event) => setForm({ ...form, title: event.target.value })}
              placeholder="Título do artigo"
            />
          </label>

          <label>
            Fonte *
            <select
              value={form.sourceId || ''}
              onChange={(event) =>
                setForm({ ...form, sourceId: Number(event.target.value) })
              }
            >
              <option value="">Selecione...</option>
              {sources.map((source) => (
                <option key={source.id} value={source.id}>
                  {source.name}
                </option>
              ))}
            </select>
          </label>
        </div>

        <label>
          Autores
          <input
            value={form.authors}
            onChange={(event) => setForm({ ...form, authors: event.target.value })}
            placeholder="João Silva, Maria Santos"
          />
          <small>Separe os autores por vírgulas.</small>
        </label>

        <label>
          Descrição
          <textarea
            rows={4}
            value={form.description}
            onChange={(event) =>
              setForm({ ...form, description: event.target.value })
            }
            placeholder="Resumo curto ou descrição do documento"
          />
        </label>

        <div className="form-grid three">
          <label>
            Idioma
            <input
              value={form.languageCode}
              onChange={(event) =>
                setForm({ ...form, languageCode: event.target.value })
              }
              placeholder="pt-BR"
            />
          </label>

          <label>
            Data de publicação
            <input
              type="date"
              value={form.publicationDate}
              onChange={(event) =>
                setForm({ ...form, publicationDate: event.target.value })
              }
            />
          </label>

          <label>
            DOI
            <input
              value={form.doi}
              onChange={(event) => setForm({ ...form, doi: event.target.value })}
              placeholder="10.xxxx/..."
            />
          </label>
        </div>

        <label>
          ID externo
          <input
            value={form.externalId}
            onChange={(event) =>
              setForm({ ...form, externalId: event.target.value })
            }
            placeholder="arXiv:..., PMID..., OpenAlex..."
          />
        </label>

        <div className="form-actions">
          {onCancel && (
            <button className="secondary" type="button" onClick={onCancel}>
              Cancelar
            </button>
          )}
          <button className="primary compact" disabled={loading}>
            {loading ? 'Aguarde...' : submitLabel}
          </button>
        </div>
      </form>
    </section>
  )
}

function Detail({ label, value }: { label: string; value: string }) {
  return (
    <div className="detail-item">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  )
}

function screenTitle(screen: Screen) {
  const titles: Record<Screen, string> = {
    dashboard: 'Visão geral',
    search: 'Pesquisa',
    articles: 'Artigos',
    import: 'Importar artigo',
    indexing: 'Indexação',
    metrics: 'Métricas',
    edit: 'Editar artigo',
  }
  return titles[screen]
}

function screenSubtitle(screen: Screen, role: string) {
  if (screen === 'dashboard') return `Olá. Perfil ativo: ${role}.`
  if (screen === 'search') return 'Consulte o acervo científico do Concordia.'
  if (screen === 'articles') return 'Gerencie documentos, fontes e metadados.'
  if (screen === 'import') return 'Adicione PDFs ao acervo de forma controlada.'
  if (screen === 'edit') return 'Atualize os metadados persistidos no MySQL.'
  if (screen === 'indexing') return 'Prepare o acervo para o motor de busca paralelo.'
  return 'Indicadores do acervo e da indexação.'
}

function formatBytes(value: number) {
  if (!value) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  const index = Math.min(
    Math.floor(Math.log(value) / Math.log(1024)),
    units.length - 1,
  )
  return `${(value / 1024 ** index).toFixed(index === 0 ? 0 : 1)} ${units[index]}`
}

export default App
