import { useCallback, useEffect, useState } from 'react'

import {
  ExportLogsCSV,
  GetLogSummary,
  ListOperationLogs,
} from '../wailsjs/go/main/App'

type OperationLog = {
  id: number
  user_id: number
  user_name: string
  operation: string
  entity: string
  entity_id: number
  details: string
  result_count: number
  created_at: string
}

type OperationCount = {
  operation: string
  total: number
}

type TermCount = {
  term: string
  total: number
}

type LogSummary = {
  total: number
  today: number
  by_operation: OperationCount[]
  top_searches: TermCount[]
}

const emptySummary: LogSummary = {
  total: 0,
  today: 0,
  by_operation: [],
  top_searches: [],
}

const operationLabels: Record<string, string> = {
  LOGIN: 'Login',
  LOGOUT: 'Logout',
  CADASTRO_USUARIO: 'Cadastro de usuário',
  IMPORTACAO_ARTIGO: 'Importação de artigo',
  IMPORTACAO_LOTE: 'Importação em lote',
  EDICAO_ARTIGO: 'Edição de artigo',
  EXCLUSAO_ARTIGO: 'Exclusão de artigo',
  INDEXACAO: 'Indexação',
  BENCHMARK: 'Benchmark',
  BUSCA: 'Busca',
  EXPORTACAO_CSV: 'Exportação CSV',
}

function describeOperation(operation: string): string {
  return operationLabels[operation] ?? operation
}

function formatDateTime(value: string): string {
  if (!value) return '—'
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) return value
  return parsed.toLocaleString('pt-BR')
}

type Props = {
  isCurator: boolean
}

export default function HistoryPanel({ isCurator }: Props) {
  const [logs, setLogs] = useState<OperationLog[]>([])
  const [summary, setSummary] = useState<LogSummary>(emptySummary)
  const [operation, setOperation] = useState('')
  const [dateFrom, setDateFrom] = useState('')
  const [dateTo, setDateTo] = useState('')
  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState('')
  const [messageKind, setMessageKind] = useState<'info' | 'error' | 'success'>('info')

  const load = useCallback(async () => {
    setLoading(true)
    setMessage('')

    try {
      const [loadedLogs, loadedSummary] = await Promise.all([
        ListOperationLogs(operation, dateFrom, dateTo, 200),
        GetLogSummary(),
      ])

      const list = (loadedLogs || []) as OperationLog[]
      setLogs(list)

      const response = loadedSummary as { success: boolean; summary: LogSummary }
      setSummary(response?.summary ?? emptySummary)

      if (list.length === 0) {
        setMessageKind('info')
        setMessage('Nenhuma operação registrada para os filtros selecionados.')
      }
    } catch (err) {
      console.error('Falha ao carregar o histórico:', err)
      setMessageKind('error')
      setMessage('Não foi possível carregar o histórico de operações.')
    } finally {
      setLoading(false)
    }
  }, [operation, dateFrom, dateTo])

  useEffect(() => {
    void load()
    // Carrega apenas na montagem; os filtros são aplicados pelo botão.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  function clearFilters() {
    setOperation('')
    setDateFrom('')
    setDateTo('')
  }

  async function handleExport() {
    setMessage('')

    try {
      const response = (await ExportLogsCSV(operation, dateFrom, dateTo, 500)) as {
        success: boolean
        message: string
        path?: string
      }

      setMessageKind(response.success ? 'success' : 'error')
      setMessage(
        response.success && response.path
          ? `${response.message} Arquivo: ${response.path}`
          : response.message,
      )

      if (response.success) {
        void load()
      }
    } catch (err) {
      console.error('Falha ao exportar o histórico:', err)
      setMessageKind('error')
      setMessage('Não foi possível exportar o histórico.')
    }
  }

  return (
    <section className="page-card">
      <div className="section-heading">
        <div>
          <h2>Histórico de operações</h2>
          <p>Registro das ações realizadas no sistema, persistido no MySQL.</p>
        </div>
        {isCurator && (
          <button className="primary compact" onClick={handleExport} disabled={loading}>
            Exportar CSV
          </button>
        )}
      </div>

      <div className="cards">
        <article className="metric-card">
          <span>Operações registradas</span>
          <strong>{summary.total}</strong>
          <small>Total no histórico</small>
        </article>
        <article className="metric-card">
          <span>Hoje</span>
          <strong>{summary.today}</strong>
          <small>Operações nas últimas 24h</small>
        </article>
        <article className="metric-card">
          <span>Tipos de operação</span>
          <strong>{summary.by_operation.length}</strong>
          <small>Categorias distintas</small>
        </article>
        <article className="metric-card">
          <span>Exibindo</span>
          <strong>{logs.length}</strong>
          <small>Registros filtrados</small>
        </article>
      </div>

      <div className="log-filters">
        <label>
          Operação
          <select value={operation} onChange={(event) => setOperation(event.target.value)}>
            <option value="">Todas</option>
            {Object.keys(operationLabels).map((key) => (
              <option key={key} value={key}>
                {operationLabels[key]}
              </option>
            ))}
          </select>
        </label>

        <label>
          De
          <input
            type="date"
            value={dateFrom}
            onChange={(event) => setDateFrom(event.target.value)}
          />
        </label>

        <label>
          Até
          <input
            type="date"
            value={dateTo}
            onChange={(event) => setDateTo(event.target.value)}
          />
        </label>

        <div className="log-filter-actions">
          <button className="primary compact" onClick={() => void load()} disabled={loading}>
            {loading ? 'Carregando...' : 'Aplicar filtros'}
          </button>
          <button className="log-clear" onClick={clearFilters} disabled={loading}>
            Limpar
          </button>
        </div>
      </div>

      {message && <div className={`workspace-message is-${messageKind}`}>{message}</div>}

      {summary.by_operation.length > 0 && (
        <div className="log-report">
          <h3>Relatório por tipo de operação</h3>
          <ul>
            {summary.by_operation.map((item) => (
              <li key={item.operation}>
                <span>{describeOperation(item.operation)}</span>
                <strong>{item.total}</strong>
              </li>
            ))}
          </ul>
        </div>
      )}

      {summary.top_searches.length > 0 && (
        <div className="log-report">
          <h3>Termos mais pesquisados</h3>
          <ul>
            {summary.top_searches.map((item) => (
              <li key={item.term}>
                <span>{item.term}</span>
                <strong>{item.total}</strong>
              </li>
            ))}
          </ul>
        </div>
      )}

      <div className="log-table-wrapper">
        <table className="log-table">
          <thead>
            <tr>
              <th>Data e hora</th>
              <th>Usuário</th>
              <th>Operação</th>
              <th>Detalhes</th>
              <th>Resultados</th>
            </tr>
          </thead>
          <tbody>
            {logs.map((item) => (
              <tr key={item.id}>
                <td>{formatDateTime(item.created_at)}</td>
                <td>{item.user_name || '—'}</td>
                <td>{describeOperation(item.operation)}</td>
                <td>{item.details || '—'}</td>
                <td>{item.operation === 'BUSCA' ? item.result_count : '—'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  )
}
