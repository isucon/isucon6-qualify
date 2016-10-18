CREATE TABLE IF NOT EXISTS setting (
    name VARCHAR(128) NOT NULL,
    json TEXT NOT NULL,
    PRIMARY KEY (name)
) DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS teams (
    id INT UNSIGNED NOT NULL,
    name VARCHAR(128) NOT NULL,
    password VARCHAR(128) NOT NULL,
    ip_address VARCHAR(32),
    instance_name VARCHAR(255),
    category ENUM('general', 'students', 'official') NOT NULL,
    azure_resource_group VARCHAR(32) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY (name)
) DEFAULT CHARSET=utf8mb4;

-- 各チームの最新・最高スコア（Pass したもののみ）
CREATE TABLE IF NOT EXISTS team_scores (
    team_id INT UNSIGNED NOT NULL, -- teams.id
    latest_score BIGINT NOT NULL,
    best_score BIGINT NOT NULL,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (team_id)
) DEFAULT CHARSET=utf8mb4;

CREATE TABLE team_scores_snapshot LIKE team_scores;

-- 各チームのスコア（Pass したもののみ）
CREATE TABLE IF NOT EXISTS scores (
    id INT UNSIGNED NOT NULL AUTO_INCREMENT,
    team_id INT UNSIGNED NOT NULL, -- teams.id
    score BIGINT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY (team_id, created_at),
    KEY (created_at)
) DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS queues (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    team_id INT NOT NULL,
    status ENUM('waiting', 'running', 'done', 'aborted') NOT NULL DEFAULT 'waiting',
    ip_address VARCHAR(32) NOT NULL,
    bench_node VARCHAR(64) DEFAULT NULL,
    result_json TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY queues_team_status_idx (team_id, status)
) DEFAULT CHARSET=utf8mb4;
