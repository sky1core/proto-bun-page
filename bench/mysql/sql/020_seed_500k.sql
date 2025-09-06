-- Seed ~500k rows (heavier dataset); comment out if not needed
DELIMITER $$
CREATE PROCEDURE seed_items_500k()
BEGIN
  DECLARE i INT DEFAULT 0;
  WHILE i < 500000 DO
    INSERT INTO items (created_at, name, score, status, payload)
    VALUES (
      FROM_UNIXTIME(1700000000 + FLOOR(i/20)),
      CONCAT('name_', LPAD(i % 1000, 3, '0')),
      FLOOR(RAND()*101),
      IF(RAND() < 0.8, 1, 0),
      REPEAT('x', IF(RAND()<0.5, 0, 512))
    );
    SET i = i + 1;
  END WHILE;
END$$
DELIMITER ;
-- CALL seed_items_500k();

