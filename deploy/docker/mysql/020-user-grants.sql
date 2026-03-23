-- GoAdmin MySQL migration 020
-- Create the application user and grant database privileges.

CREATE USER IF NOT EXISTS 'goadmin'@'%' IDENTIFIED BY 'goadmin';
GRANT ALL PRIVILEGES ON `goadmin`.* TO 'goadmin'@'%';
GRANT ALL PRIVILEGES ON `goadmin_dev`.* TO 'goadmin'@'%';
GRANT ALL PRIVILEGES ON `goadmin_prod`.* TO 'goadmin'@'%';

FLUSH PRIVILEGES;
