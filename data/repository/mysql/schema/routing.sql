CREATE TABLE IF NOT EXISTS `routing` (
  `job_category` VARCHAR(255) NOT NULL,
  `queue_name` VARCHAR(255) NOT NULL,
  PRIMARY KEY (`job_category`),
  UNIQUE KEY `job_queue` (`job_category`, `queue_name`)
) ENGINE=InnoDB DEFAULT CHARSET=binary;
