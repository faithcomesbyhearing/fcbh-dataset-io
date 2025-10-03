-- MySQL dump 10.13  Distrib 8.4.5, for macos14.7 (arm64)
--
-- Host: 127.0.0.1    Database: dbp_NEWDATA
-- ------------------------------------------------------
-- Server version	8.0.39

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `languages`
--

DROP TABLE IF EXISTS `languages`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `languages` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `glotto_id` char(8) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `iso` char(3) COLLATE utf8mb4_unicode_ci NOT NULL,
  `iso2B` char(3) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `iso2T` char(3) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `iso1` char(2) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `name` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `maps` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `development` text COLLATE utf8mb4_unicode_ci,
  `use` text COLLATE utf8mb4_unicode_ci,
  `location` text COLLATE utf8mb4_unicode_ci,
  `area` text COLLATE utf8mb4_unicode_ci,
  `population` int DEFAULT NULL,
  `population_notes` text COLLATE utf8mb4_unicode_ci,
  `notes` text COLLATE utf8mb4_unicode_ci,
  `typology` text COLLATE utf8mb4_unicode_ci,
  `writing` text COLLATE utf8mb4_unicode_ci,
  `latitude` double(11,7) DEFAULT NULL,
  `longitude` double(11,7) DEFAULT NULL,
  `status_id` char(2) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `country_id` char(2) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `rolv_code` varchar(10) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `sensitivity` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'Low',
  `pseudonym` tinyint unsigned NOT NULL DEFAULT '0',
  `deleted_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `languages_glotto_id_unique` (`glotto_id`),
  UNIQUE KEY `languages_iso2b_unique` (`iso2B`),
  UNIQUE KEY `languages_iso2t_unique` (`iso2T`),
  UNIQUE KEY `languages_iso1_unique` (`iso1`),
  KEY `languages_iso_index` (`iso`),
  KEY `language_status_foreign_key` (`status_id`),
  KEY `country_id_foreign_key` (`country_id`),
  FULLTEXT KEY `ft_index_languages_name` (`name`),
  CONSTRAINT `FK_countries_languages` FOREIGN KEY (`country_id`) REFERENCES `countries` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT `FK_language_status_languages` FOREIGN KEY (`status_id`) REFERENCES `language_status` (`id`) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=34905 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `alphabets`
--

DROP TABLE IF EXISTS `alphabets`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `alphabets` (
  `script` char(4) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `unicode_pdf` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `family` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `type` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `white_space` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `open_type_tag` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `complex_positioning` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `requires_font` tinyint(1) NOT NULL DEFAULT '0',
  `unicode` tinyint(1) NOT NULL DEFAULT '1',
  `diacritics` tinyint(1) DEFAULT NULL,
  `contextual_forms` tinyint(1) DEFAULT NULL,
  `reordering` tinyint(1) DEFAULT NULL,
  `case` tinyint(1) DEFAULT NULL,
  `split_graphs` tinyint(1) DEFAULT NULL,
  `status` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `baseline` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `ligatures` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `direction` char(3) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `direction_notes` text COLLATE utf8mb4_unicode_ci,
  `sample` text COLLATE utf8mb4_unicode_ci,
  `sample_img` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `description` text COLLATE utf8mb4_unicode_ci,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`script`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `alphabet_language`
--

DROP TABLE IF EXISTS `alphabet_language`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `alphabet_language` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `script_id` char(4) COLLATE utf8mb4_unicode_ci NOT NULL,
  `language_id` int unsigned NOT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `alphabet_language_language_id_foreign` (`language_id`),
  KEY `alphabet_language_script_id_index` (`script_id`),
  CONSTRAINT `FK_alphabets_alphabet_language` FOREIGN KEY (`script_id`) REFERENCES `alphabets` (`script`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `FK_languages_alphabet_language` FOREIGN KEY (`language_id`) REFERENCES `languages` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=328 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `countries`
--

DROP TABLE IF EXISTS `countries`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `countries` (
  `id` char(2) COLLATE utf8mb4_unicode_ci NOT NULL,
  `iso_a3` char(3) COLLATE utf8mb4_unicode_ci NOT NULL,
  `fips` char(2) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `wfb` tinyint(1) NOT NULL DEFAULT '0',
  `ethnologue` tinyint(1) NOT NULL DEFAULT '0',
  `continent` char(2) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `introduction` text COLLATE utf8mb4_unicode_ci,
  `overview` text COLLATE utf8mb4_unicode_ci,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `countries_iso_a3_unique` (`iso_a3`),
  FULLTEXT KEY `ft_index_countries_name_iso_a3` (`name`,`iso_a3`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `organizations`
--

DROP TABLE IF EXISTS `organizations`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `organizations` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `slug` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `abbreviation` char(6) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `notes` text COLLATE utf8mb4_unicode_ci,
  `primaryColor` varchar(7) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `secondaryColor` varchar(7) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `inactive` tinyint(1) DEFAULT '0',
  `url_facebook` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `url_website` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `url_donate` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `url_twitter` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `address` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `address2` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `city` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `state` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `country` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `zip` int unsigned DEFAULT NULL,
  `phone` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `email` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `email_director` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `latitude` double(11,7) DEFAULT NULL,
  `longitude` double(11,7) DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `organizations_slug_unique` (`slug`),
  UNIQUE KEY `organizations_abbreviation_unique` (`abbreviation`),
  KEY `organizations_country_foreign` (`country`),
  CONSTRAINT `FK_countries_organizations` FOREIGN KEY (`country`) REFERENCES `countries` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=4279 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `permission_pattern`
--

DROP TABLE IF EXISTS `permission_pattern`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `permission_pattern` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `mode_audio` tinyint NOT NULL DEFAULT '0',
  `mode_video` tinyint NOT NULL DEFAULT '0',
  `mode_text` tinyint NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=168 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `permission_pattern_access_group`
--

DROP TABLE IF EXISTS `permission_pattern_access_group`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `permission_pattern_access_group` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `permission_pattern_id` int unsigned NOT NULL,
  `access_groups_id` int unsigned NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_permission_pattern_access_group` (`permission_pattern_id`,`access_groups_id`),
  KEY `access_groups_id` (`access_groups_id`),
  CONSTRAINT `permission_pattern_access_group_ibfk_1` FOREIGN KEY (`permission_pattern_id`) REFERENCES `permission_pattern` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `permission_pattern_access_group_ibfk_2` FOREIGN KEY (`access_groups_id`) REFERENCES `access_groups` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=2217 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `bibles`
--

DROP TABLE IF EXISTS `bibles`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `bibles` (
  `id` varchar(12) COLLATE utf8mb4_unicode_ci NOT NULL,
  `language_id` int unsigned NOT NULL,
  `versification` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'protestant',
  `numeral_system_id` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `date` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `scope` varchar(8) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `script` char(4) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `derived` text COLLATE utf8mb4_unicode_ci,
  `copyright` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `priority` tinyint unsigned NOT NULL DEFAULT '0',
  `reviewed` tinyint(1) DEFAULT '0',
  `notes` text COLLATE utf8mb4_unicode_ci,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `bibles_id_unique` (`id`),
  KEY `bibles_language_id_foreign` (`language_id`),
  KEY `bibles_numeral_system_id_foreign` (`numeral_system_id`),
  KEY `bibles_script_foreign` (`script`),
  KEY `priority` (`priority`),
  CONSTRAINT `FK_alphabets_bibles` FOREIGN KEY (`script`) REFERENCES `alphabets` (`script`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `FK_languages_bibles` FOREIGN KEY (`language_id`) REFERENCES `languages` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `FK_numeral_systems_bibles` FOREIGN KEY (`numeral_system_id`) REFERENCES `numeral_systems` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `books`
--

DROP TABLE IF EXISTS `books`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `books` (
  `id` char(3) COLLATE utf8mb4_unicode_ci NOT NULL,
  `id_usfx` char(2) COLLATE utf8mb4_unicode_ci NOT NULL,
  `id_osis` varchar(12) COLLATE utf8mb4_unicode_ci NOT NULL,
  `book_testament` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `book_group` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `chapters` int unsigned DEFAULT NULL,
  `verses` int unsigned DEFAULT NULL,
  `name` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `notes` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `testament_order` tinyint unsigned DEFAULT NULL,
  `protestant_order` tinyint unsigned DEFAULT NULL,
  `luther_order` tinyint unsigned DEFAULT NULL,
  `synodal_order` tinyint unsigned DEFAULT NULL,
  `german_order` tinyint unsigned DEFAULT NULL,
  `kjva_order` tinyint unsigned DEFAULT NULL,
  `vulgate_order` tinyint unsigned DEFAULT NULL,
  `lxx_order` tinyint unsigned DEFAULT NULL,
  `orthodox_order` tinyint unsigned DEFAULT NULL,
  `nrsva_order` tinyint unsigned DEFAULT NULL,
  `catholic_order` tinyint unsigned DEFAULT NULL,
  `finnish_order` tinyint unsigned DEFAULT NULL,
  `messianic_order` tinyint DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `bible_translations`
--

DROP TABLE IF EXISTS `bible_translations`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `bible_translations` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `language_id` int unsigned NOT NULL,
  `bible_id` varchar(12) COLLATE utf8mb4_unicode_ci NOT NULL,
  `vernacular` tinyint(1) NOT NULL DEFAULT '0',
  `vernacular_trade` tinyint(1) NOT NULL DEFAULT '0',
  `name` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` text COLLATE utf8mb4_unicode_ci,
  `background` text COLLATE utf8mb4_unicode_ci,
  `notes` text COLLATE utf8mb4_unicode_ci,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique` (`language_id`,`bible_id`,`vernacular`),
  KEY `bible_translations_language_id_foreign` (`language_id`),
  KEY `bible_translations_bible_id_foreign` (`bible_id`),
  FULLTEXT KEY `ft_index_bible_translations_name` (`name`),
  CONSTRAINT `FK_bibles_bible_translations` FOREIGN KEY (`bible_id`) REFERENCES `bibles` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `FK_languages_bible_translations` FOREIGN KEY (`language_id`) REFERENCES `languages` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=17541 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `book_translations`
--

DROP TABLE IF EXISTS `book_translations`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `book_translations` (
  `language_id` int unsigned NOT NULL,
  `book_id` char(3) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name_long` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `name_short` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name_abbreviation` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`language_id`,`book_id`),
  KEY `book_translations_language_id_foreign` (`language_id`),
  KEY `book_translations_book_id_foreign` (`book_id`),
  CONSTRAINT `FK_books_book_translations` FOREIGN KEY (`book_id`) REFERENCES `books` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT,
  CONSTRAINT `FK_languages_book_translations` FOREIGN KEY (`language_id`) REFERENCES `languages` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `assets`
--

DROP TABLE IF EXISTS `assets`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `assets` (
  `id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `organization_id` int unsigned NOT NULL,
  `asset_type` varchar(12) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `hidden` tinyint(1) NOT NULL DEFAULT '0',
  `base_name` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `buckets_id_unique` (`id`),
  KEY `buckets_organization_id_foreign` (`organization_id`),
  CONSTRAINT `FK_organizations_assets` FOREIGN KEY (`organization_id`) REFERENCES `organizations` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `bible_fileset_types`
--

DROP TABLE IF EXISTS `bible_fileset_types`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `bible_fileset_types` (
  `id` tinyint unsigned NOT NULL AUTO_INCREMENT,
  `set_type_code` varchar(18) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `mode_id` tinyint unsigned NOT NULL DEFAULT '1',
  PRIMARY KEY (`id`),
  UNIQUE KEY `bible_fileset_types_set_type_code_unique` (`set_type_code`),
  UNIQUE KEY `bible_fileset_types_name_unique` (`name`),
  KEY `bible_fileset_types_mode_index` (`mode_id`),
  CONSTRAINT `FK_bible_fileset_modes_bible_fileset_types` FOREIGN KEY (`mode_id`) REFERENCES `bible_fileset_modes` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=20 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `bible_fileset_sizes`
--

DROP TABLE IF EXISTS `bible_fileset_sizes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `bible_fileset_sizes` (
  `id` tinyint unsigned NOT NULL AUTO_INCREMENT,
  `set_size_code` char(9) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `bible_fileset_sizes_set_size_code_unique` (`set_size_code`),
  UNIQUE KEY `bible_fileset_sizes_name_unique` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `bible_fileset_modes`
--

DROP TABLE IF EXISTS `bible_fileset_modes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `bible_fileset_modes` (
  `id` tinyint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `bible_fileset_modes_name_unique` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `license_group`
--

DROP TABLE IF EXISTS `license_group`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `license_group` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `permission_pattern_id` int unsigned DEFAULT '100',
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `copyright` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `is_copyright_combined` tinyint(1) NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_name` (`name`),
  KEY `FK_permission_pattern` (`permission_pattern_id`),
  CONSTRAINT `FK_permission_pattern` FOREIGN KEY (`permission_pattern_id`) REFERENCES `permission_pattern` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=24414 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `license_group_licensor`
--

DROP TABLE IF EXISTS `license_group_licensor`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `license_group_licensor` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `license_group_id` int unsigned NOT NULL,
  `organization_id` int unsigned NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `UK_license_group_licensor_organization_license_group` (`organization_id`,`license_group_id`),
  KEY `FK_license_group_license_group_licensor` (`license_group_id`),
  CONSTRAINT `FK_license_group_license_group_licensor` FOREIGN KEY (`license_group_id`) REFERENCES `license_group` (`id`),
  CONSTRAINT `FK_organizations_license_group_licensor` FOREIGN KEY (`organization_id`) REFERENCES `organizations` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=16484 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `bible_filesets`
--

DROP TABLE IF EXISTS `bible_filesets`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `bible_filesets` (
  `id` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `hash_id` char(12) COLLATE utf8mb4_unicode_ci NOT NULL,
  `asset_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `set_type_code` varchar(18) COLLATE utf8mb4_unicode_ci NOT NULL,
  `set_size_code` char(9) COLLATE utf8mb4_unicode_ci NOT NULL,
  `mode_id` tinyint unsigned NOT NULL,
  `license_group_id` int unsigned DEFAULT NULL,
  `published_snm` tinyint(1) NOT NULL DEFAULT '0',
  `hidden` tinyint(1) NOT NULL DEFAULT '0',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `content_loaded` tinyint(1) NOT NULL DEFAULT '0',
  `archived` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`hash_id`),
  UNIQUE KEY `unique_prefix_for_s3` (`id`,`asset_id`,`set_type_code`),
  UNIQUE KEY `unique_id_type` (`id`,`set_type_code`),
  KEY `bible_filesets_bucket_id_foreign` (`asset_id`),
  KEY `bible_filesets_set_type_code_foreign` (`set_type_code`),
  KEY `bible_filesets_set_size_code_foreign` (`set_size_code`),
  KEY `bible_filesets_id_index` (`id`),
  KEY `bible_filesets_hash_id_index` (`hash_id`),
  KEY `FK_license_group` (`license_group_id`),
  CONSTRAINT `FK_assets_bible_filesets` FOREIGN KEY (`asset_id`) REFERENCES `assets` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `FK_bible_fileset_sizes_bible_filesets` FOREIGN KEY (`set_size_code`) REFERENCES `bible_fileset_sizes` (`set_size_code`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `FK_bible_fileset_types_bible_filesets` FOREIGN KEY (`set_type_code`) REFERENCES `bible_fileset_types` (`set_type_code`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `FK_license_group` FOREIGN KEY (`license_group_id`) REFERENCES `license_group` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `bible_fileset_tags`
--

DROP TABLE IF EXISTS `bible_fileset_tags`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `bible_fileset_tags` (
  `hash_id` varchar(12) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `admin_only` tinyint(1) NOT NULL DEFAULT '1',
  `notes` text COLLATE utf8mb4_unicode_ci,
  `iso` char(3) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'eng',
  `language_id` int unsigned NOT NULL DEFAULT '6414',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'UTC',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'UTC',
  PRIMARY KEY (`hash_id`,`name`,`language_id`),
  KEY `bible_fileset_tags_hash_id_index` (`hash_id`),
  KEY `bible_fileset_tags_iso_index` (`iso`),
  KEY `language_id` (`language_id`),
  KEY `description` (`description`(4)),
  KEY `hashid_name_index` (`hash_id`,`name`),
  KEY `name_index` (`name`),
  CONSTRAINT `FK_bible_filesets_bible_fileset_tags` FOREIGN KEY (`hash_id`) REFERENCES `bible_filesets` (`hash_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `bible_files`
--

DROP TABLE IF EXISTS `bible_files`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `bible_files` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `hash_id` varchar(12) COLLATE utf8mb4_unicode_ci NOT NULL,
  `book_id` char(3) COLLATE utf8mb4_unicode_ci NOT NULL,
  `chapter_start` tinyint unsigned DEFAULT NULL,
  `chapter_end` tinyint unsigned DEFAULT NULL,
  `verse_start` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `verse_end` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `is_complete_chapter` tinyint(1) NOT NULL DEFAULT '0',
  `file_name` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `file_size` int unsigned DEFAULT NULL,
  `duration` int unsigned DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `verse_sequence` tinyint unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_hash_book_chapter_verse_filename` (`hash_id`,`book_id`,`chapter_start`,`verse_sequence`,`file_name`),
  KEY `bible_files_book_id_foreign` (`book_id`),
  CONSTRAINT `FK_bible_filesets_bible_files` FOREIGN KEY (`hash_id`) REFERENCES `bible_filesets` (`hash_id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `FK_books_bible_files` FOREIGN KEY (`book_id`) REFERENCES `books` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE=InnoDB AUTO_INCREMENT=5139444 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `bible_file_timestamps`
--

DROP TABLE IF EXISTS `bible_file_timestamps`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `bible_file_timestamps` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `bible_file_id` int unsigned NOT NULL,
  `verse_start` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `verse_end` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `timestamp` float NOT NULL,
  `timestamp_end` float DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `verse_sequence` tinyint unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `bible_file_timestamps_file_id_foreign` (`bible_file_id`),
  CONSTRAINT `FK_bible_files_bible_file_timestamps` FOREIGN KEY (`bible_file_id`) REFERENCES `bible_files` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=4424155 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `bible_file_stream_bandwidths`
--

DROP TABLE IF EXISTS `bible_file_stream_bandwidths`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `bible_file_stream_bandwidths` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `bible_file_id` int unsigned NOT NULL,
  `file_name` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `bandwidth` int unsigned NOT NULL,
  `resolution_width` int unsigned DEFAULT NULL,
  `resolution_height` int unsigned DEFAULT NULL,
  `codec` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `stream` tinyint(1) NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `bible_file_video_resolutions_bible_file_id_foreign` (`bible_file_id`),
  CONSTRAINT `FK_bible_files_bible_file_stream_bandwidths` FOREIGN KEY (`bible_file_id`) REFERENCES `bible_files` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=809354 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `bible_file_stream_bytes`
--

DROP TABLE IF EXISTS `bible_file_stream_bytes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `bible_file_stream_bytes` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `stream_bandwidth_id` int unsigned NOT NULL,
  `runtime` double(8,2) NOT NULL,
  `bytes` int NOT NULL,
  `offset` int NOT NULL,
  `timestamp_id` int unsigned NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `FK_bible_file_bandwidth_stream_bytes` (`stream_bandwidth_id`),
  KEY `FK_bible_file_timestamp_stream_bytes` (`timestamp_id`),
  CONSTRAINT `FK_bible_file_bandwidth_stream_bytes` FOREIGN KEY (`stream_bandwidth_id`) REFERENCES `bible_file_stream_bandwidths` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `FK_bible_file_timestamp_stream_bytes` FOREIGN KEY (`timestamp_id`) REFERENCES `bible_file_timestamps` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=7182736 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `bible_fileset_connections`
--

DROP TABLE IF EXISTS `bible_fileset_connections`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `bible_fileset_connections` (
  `hash_id` char(12) COLLATE utf8mb4_unicode_ci NOT NULL,
  `bible_id` varchar(12) COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`hash_id`,`bible_id`),
  UNIQUE KEY `unique_hash_id` (`hash_id`),
  KEY `bible_fileset_connections_hash_id_foreign` (`hash_id`),
  KEY `bible_fileset_connections_bible_id_index` (`bible_id`),
  KEY `index_hash_id` (`hash_id`),
  CONSTRAINT `FK_bible_filesets_bible_fileset_connections` FOREIGN KEY (`hash_id`) REFERENCES `bible_filesets` (`hash_id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `FK_bibles_bible_fileset_connections` FOREIGN KEY (`bible_id`) REFERENCES `bibles` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `access_groups`
--

DROP TABLE IF EXISTS `access_groups`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `access_groups` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `lpts_fieldname` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `mode_id` tinyint unsigned NOT NULL DEFAULT '1',
  `display_order` tinyint unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `FK_bible_fileset_modes` (`mode_id`),
  KEY `access_groups_display_order_index` (`display_order`),
  CONSTRAINT `FK_bible_fileset_modes` FOREIGN KEY (`mode_id`) REFERENCES `bible_fileset_modes` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=2046 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `access_types`
--

DROP TABLE IF EXISTS `access_types`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `access_types` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(24) COLLATE utf8mb4_unicode_ci NOT NULL,
  `country_id` char(2) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `continent_id` char(2) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `allowed` tinyint(1) NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `access_types_country_id_foreign` (`country_id`),
  CONSTRAINT `FK_countries_access_types` FOREIGN KEY (`country_id`) REFERENCES `countries` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `access_group_types`
--

DROP TABLE IF EXISTS `access_group_types`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `access_group_types` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `access_group_id` int unsigned NOT NULL,
  `access_type_id` int unsigned NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `access_group_types_access_group_id_foreign` (`access_group_id`),
  KEY `access_group_types_access_type_id_foreign` (`access_type_id`),
  CONSTRAINT `FK_access_groups_access_group_types` FOREIGN KEY (`access_group_id`) REFERENCES `access_groups` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `FK_access_types_access_group_types` FOREIGN KEY (`access_type_id`) REFERENCES `access_types` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `access_group_filesets`
--

DROP TABLE IF EXISTS `access_group_filesets`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `access_group_filesets` (
  `access_group_id` int unsigned NOT NULL,
  `hash_id` char(12) COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`access_group_id`,`hash_id`),
  KEY `FK_access_group_filesets__hash_id` (`hash_id`),
  CONSTRAINT `FK_access_groups_access_group_filesets` FOREIGN KEY (`access_group_id`) REFERENCES `access_groups` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `FK_bible_filesets_access_group_filesets` FOREIGN KEY (`hash_id`) REFERENCES `bible_filesets` (`hash_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2025-10-02 13:15:29
