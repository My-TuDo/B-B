-- 001_init.sql

SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `users` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `username` VARCHAR(50) NOT NULL,
    `email` VARCHAR(100) NOT NULL,
    `password_hash` VARCHAR(255) NOT NULL,
    `avatar` VARCHAR(500) DEFAULT '',
    `nickname` VARCHAR(50) DEFAULT '',
    `bio` VARCHAR(500) DEFAULT '',
    `role` TINYINT UNSIGNED DEFAULT 1,
    `created_at` DATETIME(3) NOT NULL,
    `updated_at` DATETIME(3) NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_users_username` (`username`),
    UNIQUE INDEX `idx_users_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `categories` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `name` VARCHAR(50) NOT NULL,
    `slug` VARCHAR(50) NOT NULL,
    `created_at` DATETIME(3) NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_categories_slug` (`slug`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `videos` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT UNSIGNED NOT NULL,
    `title` VARCHAR(100) NOT NULL,
    `description` TEXT,
    `cover_url` VARCHAR(500) DEFAULT '',
    `video_url` VARCHAR(500) NOT NULL,
    `duration` BIGINT UNSIGNED DEFAULT 0,
    `file_size` BIGINT UNSIGNED DEFAULT 0,
    `category_id` BIGINT UNSIGNED DEFAULT 0,
    `status` TINYINT DEFAULT 0,
    `views` BIGINT UNSIGNED DEFAULT 0,
    `created_at` DATETIME(3) NOT NULL,
    `updated_at` DATETIME(3) NOT NULL,
    PRIMARY KEY (`id`),
    INDEX `idx_videos_user_id` (`user_id`),
    INDEX `idx_videos_category_id` (`category_id`),
    INDEX `idx_videos_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Seed categories
INSERT INTO `categories` (`name`, `slug`, `created_at`) VALUES
('动画', 'anime', NOW(3)),
('音乐', 'music', NOW(3)),
('游戏', 'game', NOW(3)),
('知识', 'knowledge', NOW(3)),
('生活', 'life', NOW(3)),
('影视', 'movie', NOW(3)),
('科技', 'tech', NOW(3));
