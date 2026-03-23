-- GoAdmin MySQL migration 010
-- Create the application databases used by GoAdmin.

CREATE DATABASE IF NOT EXISTS `goadmin`
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

CREATE DATABASE IF NOT EXISTS `goadmin_dev`
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

CREATE DATABASE IF NOT EXISTS `goadmin_prod`
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;
