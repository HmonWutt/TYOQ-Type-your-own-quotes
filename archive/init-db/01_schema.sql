CREATE TABLE quotes (
  id SERIAL PRIMARY KEY,
  text TEXT NOT NULL,
  author VARCHAR(100) DEFAULT 'Unknown',
  source TEXT DEFAULT 'Unknown',
  tags JSONB DEFAULT '[]'::jsonb,
  word_count INT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP

);
