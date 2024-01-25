-- MySQL dump 10.13  Distrib 8.0.28, for Win64 (x86_64)
--
-- ------------------------------------------------------
-- Server version	8.0.28

--
-- Table structure for table `file_pool`
--

/*文件表，每条记录对应一个真实文件*/
DROP TABLE IF EXISTS `file_pool`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `file_pool` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL,
  `hash` varchar(32) DEFAULT NULL COMMENT '文件的唯一标识',
  `name` varchar(255) DEFAULT NULL,
  `ext` varchar(30) DEFAULT NULL COMMENT '文件扩展名',
  `size` int DEFAULT NULL COMMENT '文件大小',
  `file_key` varchar(255) DEFAULT NULL COMMENT '文件路径',
  `thumbnail` varchar(255) DEFAULT NULL,
  `link` int DEFAULT NULL,
  `store_type` int DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uuid_UNIQUE` (`uuid`),
  UNIQUE KEY `hash_UNIQUE` (`hash`)
) ENGINE=InnoDB AUTO_INCREMENT=58 DEFAULT CHARSET=utf8mb3;

--
-- Table structure for table `share`
--

/*分享表*/
DROP TABLE IF EXISTS `share`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `share` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL,
  `user_uuid` varchar(36) DEFAULT NULL,
  `file_uuid` varchar(36) DEFAULT NULL COMMENT '公共池中的唯一标识',
  `user_file_uuid` varchar(36) DEFAULT NULL COMMENT '用户池子中的唯一标识',
  `fullname` varchar(255) DEFAULT NULL,
  `code` varchar(45) DEFAULT NULL,
  `expire_time` datetime DEFAULT NULL COMMENT '失效日期',
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uuid_UNIQUE` (`uuid`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8mb3;


--
-- Table structure for table `trans`
--

/*传输记录表*/
DROP TABLE IF EXISTS `trans`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `trans` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL COMMENT 'uploadID / downloadID',
  `user_file_uuid` varchar(36) DEFAULT NULL,
  `file_key` varchar(255) DEFAULT NULL,
  `local_path` text,
  `remote_path` text COMMENT '云空间中文件路径',
  `hash` varchar(32) DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  `ext` varchar(30) DEFAULT NULL,
  `size` int DEFAULT NULL,
  `parent_uuid` varchar(36) DEFAULT NULL,
  `status` int NOT NULL COMMENT '0: 上传中\\n1: 上传成功\\n2: 上传失败',
  `user_uuid` varchar(36) NOT NULL,
  `isdown` int NOT NULL COMMENT '0为上传，1为下载',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uuid_UNIQUE` (`uuid`)
) ENGINE=InnoDB AUTO_INCREMENT=80 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='上传进度记录';

--
-- Table structure for table `user`
--

/*用户表*/
DROP TABLE IF EXISTS `user`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `user` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL,
  `name` varchar(60) DEFAULT NULL,
  `password` varchar(32) DEFAULT NULL,
  `email` varchar(100) DEFAULT NULL,
  `phone` varchar(11) DEFAULT NULL,
  `level` int DEFAULT NULL,
  `start_uuid` varchar(36) DEFAULT NULL,
  `now_volume` bigint DEFAULT NULL,
  `total_volume` bigint DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uuid_UNIQUE` (`uuid`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb3;


--
-- Table structure for table `user_file`
--

/*用户文件表，file_pool中真实文件的索引*/
DROP TABLE IF EXISTS `user_file`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_file` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL,
  `user_uuid` varchar(36) DEFAULT NULL,
  `parent_id` int DEFAULT NULL COMMENT '父节点id，用于查找',
  `file_uuid` varchar(36) DEFAULT NULL,
  `ext` varchar(255) DEFAULT NULL COMMENT '文件或文件夹类型',
  `name` varchar(255) DEFAULT NULL,
  `size` int DEFAULT NULL COMMENT '冗余字段',
  `hash` varchar(32) DEFAULT NULL COMMENT '冗余字段',
  `thumbnail` varchar(255) DEFAULT NULL COMMENT '冗余字段',
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uuid_UNIQUE` (`uuid`)
) ENGINE=InnoDB AUTO_INCREMENT=101 DEFAULT CHARSET=utf8mb3;
