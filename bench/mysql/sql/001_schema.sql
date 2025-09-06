CREATE TABLE items (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  created_at DATETIME NOT NULL,
  name VARCHAR(128) NOT NULL,
  score INT NOT NULL,
  status TINYINT NOT NULL,
  payload TEXT NULL
) ENGINE=InnoDB;

-- Indexes aligned with orderings
CREATE INDEX idx_items_created_desc_id_desc ON items (created_at DESC, id DESC);
CREATE INDEX idx_items_name_asc_id_asc ON items (name ASC, id ASC);
CREATE INDEX idx_items_score_desc_name_asc_id_asc ON items (score DESC, name ASC, id ASC);
CREATE INDEX idx_items_status_created_desc_id_desc ON items (status, created_at DESC, id DESC);

