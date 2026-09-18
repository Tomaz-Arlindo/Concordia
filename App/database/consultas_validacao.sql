USE concordia;

-- Artigos e status
SELECT id, title, index_status, indexed_at
FROM articles
ORDER BY id;

-- Quantidade de termos
SELECT COUNT(*) AS total_terms
FROM terms;

-- Quantidade de postings
SELECT COUNT(*) AS total_occurrences
FROM term_occurrences;

-- Estatísticas por documento
SELECT
    a.id,
    a.title,
    s.document_length,
    s.indexed_terms,
    s.updated_at
FROM article_index_stats s
JOIN articles a ON a.id = s.article_id
ORDER BY a.id;

-- Exemplo de postings mais frequentes
SELECT
    t.term,
    t.document_frequency,
    o.article_id,
    o.term_frequency
FROM terms t
JOIN term_occurrences o ON o.term_id = t.id
ORDER BY o.term_frequency DESC, t.term
LIMIT 50;
