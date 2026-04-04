ALTER TABLE `cl2007_prospects`
    MODIFY COLUMN `number` BIGINT NOT NULL,
    ADD CONSTRAINT `uq_prospect_number` UNIQUE (`number`),
    ADD CONSTRAINT `chk_prospect_number_positive` CHECK (`number` > 0);
