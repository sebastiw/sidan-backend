-- MySQL dump 10.13  Distrib 5.7.33, for Linux (x86_64)
--
-- Host: localhost    Database: cl
-- ------------------------------------------------------
-- Server version	5.7.33-0ubuntu0.18.04.1
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Dumping routines for database 'cl'
--
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_general_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = '' */ ;
DELIMITER ;;
CREATE PROCEDURE `ReadEntries`(IN `Skip` INT, IN `Take` INT, IN `user_id` VARCHAR(255))
    READS SQL DATA
BEGIN
	select
		m.id,
		m.date,
		concat("", m.time) as `time`,
		if( p.user_id IS NULL
			or group_concat(p.user_id) = "0",
			msg,
			if( m.sig = concat("#", user_id) or
				concat("#", group_concat(p.user_id separator ",#")) like concat("%#", user_id, "%"),
			concat("<small>hemlis Till #", group_concat(p.user_id separator ",#"), ":</small><br>", m.msg),
			"hemlis" )
		) as `msg`,
		m.status,
		m.sig,
		m.place,
		m.enheter,
		m.lat,
		m.lon,
		count(l.id) as Likes,
		p.user_id IS NOT NULL as Secret,
		p.user_id IS NOT NULL && group_concat(p.user_id) <> "0" as PersonalSecret
	from cl2003_msgs as m
	left join 2003_likes as l on m.id = l.id
	left join cl2003_permissions p on m.id = p.id
	group by m.id
	order by m.id desc
	limit Skip, Take;
END ;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_general_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = '' */ ;
DELIMITER ;;
CREATE PROCEDURE `ReadEntries_test`(IN `Skip` INT, IN `Take` INT, IN `user_id` VARCHAR(255))
    READS SQL DATA
BEGIN
	select
		m.id,
		m.date,
		concat("", m.time) as `time`,
		if( p.user_id IS NULL
			or group_concat(p.user_id) = "0",
			msg,
			if( m.sig = concat("#", user_id) or
				concat("#", group_concat(p.user_id separator ",#")) like concat("%#", user_id, "%"),
			concat("<small>hemlis Till #", group_concat(p.user_id separator ",#"), ":</small><br>", m.msg),
			"hemlis" )
		) as `msg`,
		m.status,
		m.sig,
		m.place,
		m.enheter,
		m.lat,
		m.lon,
		count(l.id) as Likes,
		p.user_id IS NOT NULL as Secret,
		p.user_id IS NOT NULL && group_concat(p.user_id) <> "0" as PersonalSecret
	from cl2003_msgs as m
	left join 2003_likes as l on m.id = l.id
	left join cl2003_permissions p on m.id = p.id
	group by m.id
	order by m.id desc
	limit Skip, Take;
END ;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2021-04-05  3:21:37
