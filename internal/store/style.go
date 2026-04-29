package store

import "database/sql"

type StyleFingerprint struct {
	ID         int64
	Category   string
	Pattern    string
	Frequency  int
	Examples   string
	Language   string
	Framework  string
	Project    string
	Confidence float64
	Source     string
}

type AddStyleFingerprintParams struct {
	Category   string
	Pattern    string
	Examples   string
	Language   string
	Framework  string
	Project    string
	Confidence float64
	Source     string
}

func (s *Store) AddStyleFingerprint(p AddStyleFingerprintParams) (int64, error) {
	var id int64
	err := s.withTx(func(tx *sql.Tx) error {
		var existingID int64
		var existingFreq int
		err := tx.QueryRow(
			"SELECT id, frequency FROM style_fingerprints WHERE category = ? AND pattern = ? AND COALESCE(language,'') = COALESCE(?,'') AND COALESCE(framework,'') = COALESCE(?,'')",
			p.Category, p.Pattern, p.Language, p.Framework,
		).Scan(&existingID, &existingFreq)
		if err == nil {
			_, err = tx.Exec("UPDATE style_fingerprints SET frequency = ?, updated_at = datetime('now') WHERE id = ?", existingFreq+1, existingID)
			id = existingID
			return err
		}
		if err != sql.ErrNoRows {
			return err
		}
		res, err := tx.Exec(
			"INSERT INTO style_fingerprints (category, pattern, examples, language, framework, project, confidence, source) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			p.Category, p.Pattern, p.Examples, p.Language, p.Framework, p.Project, p.Confidence, p.Source,
		)
		if err != nil {
			return err
		}
		id, err = res.LastInsertId()
		return err
	})
	return id, err
}

func (s *Store) GetStyleFingerprints(category, language string) ([]StyleFingerprint, error) {
	query := "SELECT id, category, pattern, frequency, COALESCE(examples,''), COALESCE(language,''), COALESCE(framework,''), COALESCE(project,''), confidence, COALESCE(source,'seed') FROM style_fingerprints WHERE 1=1"
	var args []any
	if category != "" {
		query += " AND category = ?"
		args = append(args, category)
	}
	if language != "" {
		query += " AND language = ?"
		args = append(args, language)
	}
	query += " ORDER BY frequency DESC, confidence DESC"
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []StyleFingerprint
	for rows.Next() {
		var sf StyleFingerprint
		if err := rows.Scan(&sf.ID, &sf.Category, &sf.Pattern, &sf.Frequency, &sf.Examples, &sf.Language, &sf.Framework, &sf.Project, &sf.Confidence, &sf.Source); err != nil {
			return nil, err
		}
		results = append(results, sf)
	}
	return results, nil
}

func (s *Store) DeleteStyleFingerprint(id int64) error {
	_, err := s.db.Exec("DELETE FROM style_fingerprints WHERE id = ?", id)
	return err
}
