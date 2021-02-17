-- MySQL dump 10.13  Distrib 5.7.33, for Linux (x86_64)
--
-- Host: localhost    Database: cl
-- ------------------------------------------------------
-- Server version	5.7.33-0ubuntu0.18.04.1

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `2003_ditch`
--

DROP TABLE IF EXISTS `2003_ditch`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `2003_ditch` (
  `date` date NOT NULL,
  `time` time NOT NULL,
  `id` int(11) NOT NULL,
  `sig` varchar(255) NOT NULL,
  `host` varchar(255) NOT NULL,
  KEY `id_index` (`id`) USING BTREE,
  KEY `sig_index` (`sig`) USING BTREE,
  KEY `date` (`date`)
) ENGINE=MyISAM DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `2003_likes`
--

DROP TABLE IF EXISTS `2003_likes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `2003_likes` (
  `date` date NOT NULL,
  `time` time NOT NULL,
  `id` int(11) NOT NULL,
  `sig` varchar(255) NOT NULL,
  `host` varchar(255) NOT NULL,
  KEY `id_index` (`id`) USING BTREE,
  KEY `sig_index` (`sig`) USING BTREE,
  KEY `date` (`date`)
) ENGINE=MyISAM DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `2009_arrsidan`
--

DROP TABLE IF EXISTS `2009_arrsidan`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `2009_arrsidan` (
  `id` int(11) NOT NULL,
  `date` varchar(20) NOT NULL,
  `plats` varchar(100) NOT NULL,
  `organisator` varchar(20) NOT NULL,
  `deltagare` varchar(255) NOT NULL,
  `losen` varchar(20) NOT NULL,
  `fularr` varchar(10) NOT NULL
) ENGINE=MyISAM DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cl2003_hosts`
--

DROP TABLE IF EXISTS `cl2003_hosts`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `cl2003_hosts` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `innow` int(11) NOT NULL DEFAULT '0',
  `name1` varchar(30) NOT NULL DEFAULT '',
  `name2` varchar(30) NOT NULL DEFAULT '',
  `pattern` varchar(255) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=689 DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cl2003_members`
--

DROP TABLE IF EXISTS `cl2003_members`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `cl2003_members` (
  `ID` int(11) NOT NULL AUTO_INCREMENT,
  `sig` varchar(10) DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  `alias` varchar(10) DEFAULT NULL,
  `ICQ` varchar(10) DEFAULT NULL,
  `tel` varchar(15) DEFAULT NULL,
  `mob` varchar(15) DEFAULT NULL,
  `kan` text,
  PRIMARY KEY (`ID`)
) ENGINE=MyISAM AUTO_INCREMENT=13 DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cl2003_msgs`
--

DROP TABLE IF EXISTS `cl2003_msgs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `cl2003_msgs` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `date` date NOT NULL DEFAULT '0000-00-00',
  `time` time NOT NULL DEFAULT '00:00:00',
  `msg` text NOT NULL,
  `status` smallint(6) NOT NULL DEFAULT '0',
  `cl` smallint(6) NOT NULL DEFAULT '0',
  `sig` varchar(255) NOT NULL DEFAULT '',
  `email` varchar(255) NOT NULL DEFAULT '',
  `place` varchar(255) NOT NULL DEFAULT '',
  `ip` varchar(15) DEFAULT NULL,
  `host` varchar(255) DEFAULT NULL,
  `olsug` int(11) NOT NULL DEFAULT '-1',
  `enheter` int(11) NOT NULL DEFAULT '0',
  `lat` float(18,15) DEFAULT NULL,
  `lon` float(18,15) DEFAULT NULL,
  `report` int(1) DEFAULT '0',
  KEY `date` (`date`),
  KEY `id` (`id`),
  KEY `lat_lon` (`lat`,`lon`)
) ENGINE=MyISAM AUTO_INCREMENT=248486 DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cl2003_msgs_kumpaner`
--

DROP TABLE IF EXISTS `cl2003_msgs_kumpaner`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `cl2003_msgs_kumpaner` (
  `id` int(11) NOT NULL,
  `number` int(4) DEFAULT NULL,
  KEY `id` (`id`)
) ENGINE=MyISAM DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cl2003_permissions`
--

DROP TABLE IF EXISTS `cl2003_permissions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `cl2003_permissions` (
  `id` int(11) NOT NULL DEFAULT '0',
  `user_id` int(11) NOT NULL DEFAULT '0',
  KEY `permission_id` (`id`),
  KEY `user_id` (`user_id`)
) ENGINE=MyISAM DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cl2003_t_shirt_order`
--

DROP TABLE IF EXISTS `cl2003_t_shirt_order`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `cl2003_t_shirt_order` (
  `ID` smallint(6) NOT NULL AUTO_INCREMENT,
  `name` varchar(5) NOT NULL DEFAULT '',
  `size` enum('S','M','L','XL','XXL') NOT NULL DEFAULT 'S',
  `number` smallint(6) NOT NULL DEFAULT '0',
  `quality` enum('tunn','tjock') NOT NULL DEFAULT 'tjock',
  `payed` enum('nej','ja') NOT NULL DEFAULT 'nej',
  PRIMARY KEY (`ID`)
) ENGINE=MyISAM AUTO_INCREMENT=105 DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cl2004_poll`
--

DROP TABLE IF EXISTS `cl2004_poll`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `cl2004_poll` (
  `ID` int(11) NOT NULL AUTO_INCREMENT,
  `theme` varchar(255) NOT NULL DEFAULT 'Supa?',
  `yae` varchar(50) NOT NULL DEFAULT 'Ja',
  `nay` varchar(50) NOT NULL DEFAULT 'Nej',
  `date` date NOT NULL DEFAULT '0000-00-00',
  `time` time NOT NULL DEFAULT '00:00:00',
  `ip` varchar(15) NOT NULL DEFAULT '',
  `host` varchar(255) NOT NULL DEFAULT '',
  PRIMARY KEY (`ID`)
) ENGINE=MyISAM AUTO_INCREMENT=1286 DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cl2004_poll_votes`
--

DROP TABLE IF EXISTS `cl2004_poll_votes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `cl2004_poll_votes` (
  `poll_ID` int(11) NOT NULL DEFAULT '0',
  `host` varchar(255) NOT NULL DEFAULT '',
  `vote` tinyint(4) NOT NULL DEFAULT '0',
  KEY `poll_ID_index` (`poll_ID`) USING BTREE,
  KEY `vote_index` (`vote`) USING BTREE
) ENGINE=MyISAM DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cl2007_members`
--

DROP TABLE IF EXISTS `cl2007_members`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `cl2007_members` (
  `id` int(10) NOT NULL AUTO_INCREMENT,
  `number` int(4) DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  `email` varchar(255) DEFAULT NULL,
  `im` varchar(100) NOT NULL,
  `phone` varchar(255) DEFAULT NULL,
  `adress` varchar(511) DEFAULT NULL,
  `adressurl` text,
  `title` varchar(255) DEFAULT NULL,
  `history` text,
  `picture` text,
  `password` varchar(60) DEFAULT NULL,
  `isvalid` int(1) DEFAULT NULL,
  `password_classic` varchar(50) DEFAULT '',
  `password_classic_resetstring` varchar(50) DEFAULT '',
  `password_resetstring` varchar(60) DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `number` (`number`)
) ENGINE=MyISAM AUTO_INCREMENT=400 DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cl2007_prospects`
--

DROP TABLE IF EXISTS `cl2007_prospects`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `cl2007_prospects` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `status` varchar(1) NOT NULL,
  `number` int(3) NOT NULL,
  `name` varchar(255) NOT NULL,
  `email` varchar(255) NOT NULL,
  `phone` varchar(255) NOT NULL,
  `history` text NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=275 DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cl2014_gcm`
--

DROP TABLE IF EXISTS `cl2014_gcm`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `cl2014_gcm` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `sig` varchar(20) CHARACTER SET utf8 COLLATE utf8_swedish_ci NOT NULL,
  `regId` varchar(256) CHARACTER SET utf8 COLLATE utf8_swedish_ci NOT NULL,
  `active` bit(1) NOT NULL DEFAULT b'1',
  `deviceId` varchar(50) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `u_regId` (`regId`)
) ENGINE=MyISAM AUTO_INCREMENT=331 DEFAULT CHARSET=latin1 COMMENT='Used for push notifications to devices.';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cl2015_arrsidan`
--

DROP TABLE IF EXISTS `cl2015_arrsidan`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `cl2015_arrsidan` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `namn` varchar(255) COLLATE utf8_swedish_ci DEFAULT NULL,
  `start_date` varchar(20) COLLATE utf8_swedish_ci DEFAULT NULL,
  `plats` varchar(100) COLLATE utf8_swedish_ci DEFAULT NULL,
  `organisator` varchar(20) COLLATE utf8_swedish_ci DEFAULT '',
  `deltagare` varchar(255) COLLATE utf8_swedish_ci DEFAULT '',
  `kanske` varchar(255) COLLATE utf8_swedish_ci DEFAULT '',
  `hetsade` varchar(255) COLLATE utf8_swedish_ci DEFAULT '',
  `losen` varchar(20) COLLATE utf8_swedish_ci DEFAULT NULL,
  `fularr` varchar(10) COLLATE utf8_swedish_ci DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2096 DEFAULT CHARSET=utf8 COLLATE=utf8_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cl_news`
--

DROP TABLE IF EXISTS `cl_news`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `cl_news` (
  `Id` int(11) NOT NULL AUTO_INCREMENT,
  `header` varchar(255) DEFAULT NULL,
  `body` text,
  `date` date DEFAULT NULL,
  `time` time DEFAULT NULL,
  PRIMARY KEY (`Id`)
) ENGINE=MyISAM AUTO_INCREMENT=232 DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cl_visitors`
--

DROP TABLE IF EXISTS `cl_visitors`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `cl_visitors` (
  `date` date DEFAULT NULL,
  `time` time DEFAULT NULL,
  `ip` varchar(15) DEFAULT NULL,
  `host` varchar(255) DEFAULT NULL,
  `comment` varchar(255) NOT NULL,
  `sig` varchar(15) DEFAULT NULL,
  `ts` datetime DEFAULT NULL,
  `ua` varchar(255) DEFAULT NULL,
  KEY `date` (`date`)
) ENGINE=MyISAM DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `hetsa`
--

DROP TABLE IF EXISTS `hetsa`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `hetsa` (
  `uid` int(11) NOT NULL AUTO_INCREMENT,
  `arr` varchar(50) NOT NULL,
  `name1` varchar(255) NOT NULL,
  `name2` varchar(255) NOT NULL,
  `email` varchar(255) NOT NULL,
  `comment` varchar(255) NOT NULL,
  `date` date NOT NULL,
  `time` time NOT NULL,
  `code` int(11) NOT NULL,
  `code2` int(11) NOT NULL,
  `mystring` varchar(12) NOT NULL,
  `lastdate` date NOT NULL,
  `lasttime` time NOT NULL,
  PRIMARY KEY (`uid`)
) ENGINE=MyISAM AUTO_INCREMENT=59 DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `poll_admin_visitors`
--

DROP TABLE IF EXISTS `poll_admin_visitors`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `poll_admin_visitors` (
  `date` date DEFAULT NULL,
  `time` time DEFAULT NULL,
  `ip` varchar(15) DEFAULT NULL,
  `host` varchar(255) DEFAULT NULL
) ENGINE=MyISAM DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `suplogg_arr`
--

DROP TABLE IF EXISTS `suplogg_arr`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `suplogg_arr` (
  `ID` int(11) NOT NULL DEFAULT '0',
  `Name` varchar(100) NOT NULL DEFAULT '',
  `Timestamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`ID`)
) ENGINE=MyISAM DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `suplogg_deltagare`
--

DROP TABLE IF EXISTS `suplogg_deltagare`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `suplogg_deltagare` (
  `Arr` int(11) NOT NULL DEFAULT '0',
  `Name` varchar(50) NOT NULL DEFAULT '',
  `Enheter` int(11) NOT NULL DEFAULT '0',
  KEY `Arr` (`Arr`)
) ENGINE=MyISAM DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `toastAnswers`
--

DROP TABLE IF EXISTS `toastAnswers`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `toastAnswers` (
  `id` int(11) DEFAULT NULL,
  `answer` int(11) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `toastQuestions`
--

DROP TABLE IF EXISTS `toastQuestions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `toastQuestions` (
  `question` varchar(200) COLLATE utf8_swedish_ci DEFAULT NULL,
  `answer1` varchar(100) COLLATE utf8_swedish_ci DEFAULT NULL,
  `answer2` varchar(100) COLLATE utf8_swedish_ci DEFAULT NULL,
  `answer3` varchar(100) COLLATE utf8_swedish_ci DEFAULT NULL,
  `answer4` varchar(100) COLLATE utf8_swedish_ci DEFAULT NULL,
  `correct` int(11) DEFAULT NULL,
  `position` int(11) DEFAULT NULL,
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `active` smallint(6) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `position_id` (`position`,`id`)
) ENGINE=InnoDB AUTO_INCREMENT=69 DEFAULT CHARSET=utf8 COLLATE=utf8_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2021-02-17 21:59:47
