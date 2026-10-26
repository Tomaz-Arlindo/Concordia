USE concordia;

CREATE TABLE IF NOT EXISTS operation_logs (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NULL,
    operation VARCHAR(40) NOT NULL,
    entity VARCHAR(40) NOT NULL DEFAULT '',
    entity_id BIGINT UNSIGNED NULL,
    details VARCHAR(500) NOT NULL DEFAULT '',
    result_count INT NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    INDEX idx_operation_logs_operation (operation),
    INDEX idx_operation_logs_created (created_at),
    INDEX idx_operation_logs_user (user_id),
    CONSTRAINT fk_operation_logs_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE SET NULL
        ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
