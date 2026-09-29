-- Concordia - Schema MySQL 8
-- Cria o banco, tabelas, índices, chaves estrangeiras e fontes padrão.

CREATE DATABASE IF NOT EXISTS concordia
CHARACTER SET utf8mb4
COLLATE utf8mb4_unicode_ci;

USE concordia;

CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name VARCHAR(150) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(30) NOT NULL DEFAULT 'PESQUISADOR',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_users_email (email),
    INDEX idx_users_role (role)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS article_sources (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(500),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_article_sources_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS authors (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    orcid VARCHAR(50),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    INDEX idx_authors_name (name),
    UNIQUE KEY uk_authors_orcid (orcid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS articles (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    filename VARCHAR(255),
    file_path VARCHAR(1000),
    content_type VARCHAR(100),
    size BIGINT UNSIGNED,
    user_id BIGINT UNSIGNED,
    source_id BIGINT UNSIGNED,
    abstract_text TEXT,
    external_id VARCHAR(255),
    doi VARCHAR(255),
    language_code VARCHAR(10),
    publication_date DATE,
    checksum_sha256 CHAR(64),
    index_status VARCHAR(30) NOT NULL DEFAULT 'PENDING',
    indexed_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
        ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_articles_external_id (external_id),
    UNIQUE KEY uk_articles_doi (doi),
    UNIQUE KEY uk_articles_checksum (checksum_sha256),
    INDEX idx_articles_title (title),
    INDEX idx_articles_publication_date (publication_date),
    INDEX idx_articles_index_status (index_status),
    INDEX idx_articles_user (user_id),
    INDEX idx_articles_source (source_id),
    CONSTRAINT fk_articles_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE SET NULL
        ON UPDATE CASCADE,
    CONSTRAINT fk_articles_source
        FOREIGN KEY (source_id)
        REFERENCES article_sources(id)
        ON DELETE SET NULL
        ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS article_authors (
    article_id BIGINT UNSIGNED NOT NULL,
    author_id BIGINT UNSIGNED NOT NULL,
    author_order INT UNSIGNED NOT NULL DEFAULT 1,
    PRIMARY KEY (article_id, author_id),
    INDEX idx_article_authors_author (author_id),
    CONSTRAINT fk_article_authors_article
        FOREIGN KEY (article_id)
        REFERENCES articles(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE,
    CONSTRAINT fk_article_authors_author
        FOREIGN KEY (author_id)
        REFERENCES authors(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS terms (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    term VARCHAR(255) NOT NULL,
    document_frequency BIGINT UNSIGNED NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_terms_term (term),
    INDEX idx_terms_document_frequency (document_frequency)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS term_occurrences (
    term_id BIGINT UNSIGNED NOT NULL,
    article_id BIGINT UNSIGNED NOT NULL,
    term_frequency BIGINT UNSIGNED NOT NULL DEFAULT 0,
    positions JSON,
    PRIMARY KEY (term_id, article_id),
    INDEX idx_term_occurrences_article (article_id),
    INDEX idx_term_occurrences_term_frequency (term_frequency),
    CONSTRAINT fk_term_occurrences_term
        FOREIGN KEY (term_id)
        REFERENCES terms(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE,
    CONSTRAINT fk_term_occurrences_article
        FOREIGN KEY (article_id)
        REFERENCES articles(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS article_index_stats (
    article_id BIGINT UNSIGNED NOT NULL,
    document_length BIGINT UNSIGNED NOT NULL DEFAULT 0,
    indexed_terms BIGINT UNSIGNED NOT NULL DEFAULT 0,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
        ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (article_id),
    CONSTRAINT fk_article_index_stats_article
        FOREIGN KEY (article_id)
        REFERENCES articles(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO article_sources (name, description)
VALUES
    ('local', 'Artigo importado manualmente pelo usuário'),
    ('arxiv', 'Artigos provenientes do arXiv'),
    ('pubmed', 'Artigos provenientes do PubMed'),
    ('openalex', 'Metadados provenientes do OpenAlex')
ON DUPLICATE KEY UPDATE
    description = VALUES(description);
