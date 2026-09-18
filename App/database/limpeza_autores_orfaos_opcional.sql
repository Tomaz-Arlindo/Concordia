USE concordia;

-- 1) Conferir autores que não possuem vínculo com nenhum artigo atual.
SELECT
    au.id,
    au.name,
    au.orcid
FROM authors au
LEFT JOIN article_authors aa ON aa.author_id = au.id
WHERE aa.author_id IS NULL
ORDER BY au.name;

-- 2) Métrica usada pela aplicação após este patch.
SELECT COUNT(DISTINCT aa.author_id) AS autores_vinculados
FROM article_authors aa
JOIN articles a ON a.id = aa.article_id;

-- 3) LIMPEZA OPCIONAL.
-- Só descomente se o grupo decidir que autores sem artigos não devem ser preservados.
-- DELETE au
-- FROM authors au
-- LEFT JOIN article_authors aa ON aa.author_id = au.id
-- WHERE aa.author_id IS NULL;
