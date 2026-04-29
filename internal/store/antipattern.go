package store

import "database/sql"

type AntiPattern struct {
	ID              int64
	Title           string
	Description     string
	Category        string
	Severity        string
	Language        string
	Framework       string
	ExampleBad      string
	ExampleGood     string
	OccurrenceCount int
	LastSeenAt      string
	Project         string
}

type AddAntiPatternParams struct {
	Title       string
	Description string
	Category    string
	Severity    string
	Language    string
	Framework   string
	ExampleBad  string
	ExampleGood string
	Project     string
}

func (s *Store) AddAntiPattern(p AddAntiPatternParams) (int64, error) {
	var id int64
	err := s.withTx(func(tx *sql.Tx) error {
		var existingID int64
		var existingCount int
		err := tx.QueryRow(
			"SELECT id, occurrence_count FROM anti_patterns WHERE title = ? AND COALESCE(language,'') = COALESCE(?,'')",
			p.Title, p.Language,
		).Scan(&existingID, &existingCount)
		if err == nil {
			_, err = tx.Exec("UPDATE anti_patterns SET occurrence_count = ?, last_seen_at = datetime('now'), updated_at = datetime('now') WHERE id = ?", existingCount+1, existingID)
			id = existingID
			return err
		}
		if err != sql.ErrNoRows {
			return err
		}
		res, err := tx.Exec(
			"INSERT INTO anti_patterns (title, description, category, severity, language, framework, example_bad, example_good, project) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
			p.Title, p.Description, p.Category, p.Severity, p.Language, p.Framework, p.ExampleBad, p.ExampleGood, p.Project,
		)
		if err != nil {
			return err
		}
		id, err = res.LastInsertId()
		return err
	})
	return id, err
}

func (s *Store) GetAntiPatterns(category, language string) ([]AntiPattern, error) {
	query := "SELECT id, title, COALESCE(description,''), category, severity, COALESCE(language,''), COALESCE(framework,''), COALESCE(example_bad,''), COALESCE(example_good,''), occurrence_count, last_seen_at, COALESCE(project,'') FROM anti_patterns WHERE 1=1"
	var args []any
	if category != "" {
		query += " AND category = ?"
		args = append(args, category)
	}
	if language != "" {
		query += " AND language = ?"
		args = append(args, language)
	}
	query += " ORDER BY occurrence_count DESC, severity DESC"
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []AntiPattern
	for rows.Next() {
		var ap AntiPattern
		if err := rows.Scan(&ap.ID, &ap.Title, &ap.Description, &ap.Category, &ap.Severity, &ap.Language, &ap.Framework, &ap.ExampleBad, &ap.ExampleGood, &ap.OccurrenceCount, &ap.LastSeenAt, &ap.Project); err != nil {
			return nil, err
		}
		results = append(results, ap)
	}
	return results, nil
}

func (s *Store) DeleteAntiPattern(id int64) error {
	_, err := s.db.Exec("DELETE FROM anti_patterns WHERE id = ?", id)
	return err
}
