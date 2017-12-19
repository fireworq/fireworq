CREATE TABLE IF NOT EXISTS `{{.Failure}}` (
  `failure_id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `job_id` BIGINT UNSIGNED NOT NULL,
  `category` VARCHAR(255) NOT NULL,
  `url` BLOB,
  `payload` MEDIUMBLOB,
  `result` MEDIUMBLOB,
  `fail_count` INT UNSIGNED NOT NULL,
  `failed_at` BIGINT UNSIGNED NOT NULL,
  `created_at` BIGINT UNSIGNED NOT NULL,
  PRIMARY KEY (`failure_id`),
  KEY `creation_order` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=binary;
