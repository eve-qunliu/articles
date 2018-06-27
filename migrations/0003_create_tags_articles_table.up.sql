CREATE TABLE tags_articles
(
  id            serial PRIMARY KEY,
  tag_id        integer REFERENCES tags ON DELETE CASCADE,
  article_id    integer REFERENCES articles ON DELETE CASCADE,
  created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX index_tags_articles_on_tag_id_and_article_id ON tags_articles (tag_id, article_id);
CREATE TRIGGER update_tags_articles BEFORE UPDATE ON tags_articles FOR EACH ROW EXECUTE PROCEDURE set_updated_at();
