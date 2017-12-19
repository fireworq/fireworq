CREATE TABLE IF NOT EXISTS `{{.JobQueue}}` (
  `job_id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `next_try` BIGINT UNSIGNED NOT NULL,
  `grabber_id` BIGINT UNSIGNED,
  `status` ENUM('claimed', 'grabbed') NOT NULL DEFAULT 'claimed',
  `created_at` BIGINT UNSIGNED NOT NULL,
  `retry_count` INT UNSIGNED NOT NULL DEFAULT 0,
  `retry_delay` INT UNSIGNED NOT NULL DEFAULT 0,
  `fail_count` INT UNSIGNED NOT NULL DEFAULT 0,

  `category` VARCHAR(255) NOT NULL,
  `url` BLOB,
  `payload` MEDIUMBLOB,
  `timeout` INT UNSIGNED,

  PRIMARY KEY (`job_id`),
  KEY `grab` (`status`, `next_try`)
) ENGINE=InnoDB DEFAULT CHARSET=binary;
