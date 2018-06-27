CREATE OR REPLACE FUNCTION set_updated_at()
  RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TABLE articles
(
  id            serial PRIMARY KEY,
  title         TEXT NOT NULL,
  body          TEXT NOT NULL,
  date          varchar(255) NOT NULL,
  created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX index_articles_on_date ON articles (date);
CREATE TRIGGER update_articles BEFORE UPDATE ON articles FOR EACH ROW EXECUTE PROCEDURE set_updated_at();
