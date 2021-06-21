CREATE TABLE IF NOT EXISTS `queue_throttle` (
  `name` VARCHAR(255) NOT NULL,
  `max_dispatches_per_second` FLOAT UNSIGNED NOT NULL,
  `max_burst_size` INT UNSIGNED NOT NULL,
  PRIMARY KEY (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=binary;
