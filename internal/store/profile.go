package store

import "database/sql"

type CommunicationPreference struct {
	ID         int64
	Category   string
	Preference string
	Value      string
	Confidence float64
	Source     string
}

type SetCommunicationPreferenceParams struct {
	Category   string
	Preference string
	Value      string
	Confidence float64
	Source     string
}

func (s *Store) SetCommunicationPreference(p SetCommunicationPreferenceParams) (int64, error) {
	var id int64
	err := s.withTx(func(tx *sql.Tx) error {
		var existingID int64
		err := tx.QueryRow(
			"SELECT id FROM communication_profile WHERE category = ? AND preference = ?",
			p.Category, p.Preference,
		).Scan(&existingID)
		if err == nil {
			_, err = tx.Exec(
				"UPDATE communication_profile SET value = ?, confidence = ?, source = ?, updated_at = datetime('now') WHERE id = ?",
				p.Value, p.Confidence, p.Source, existingID,
			)
			id = existingID
			return err
		}
		if err != sql.ErrNoRows {
			return err
		}
		res, err := tx.Exec(
			"INSERT INTO communication_profile (category, preference, value, confidence, source) VALUES (?, ?, ?, ?, ?)",
			p.Category, p.Preference, p.Value, p.Confidence, p.Source,
		)
		if err != nil {
			return err
		}
		id, err = res.LastInsertId()
		return err
	})
	return id, err
}

func (s *Store) GetCommunicationProfile(category string) ([]CommunicationPreference, error) {
	query := "SELECT id, category, preference, value, confidence, COALESCE(source,'seed') FROM communication_profile WHERE 1=1"
	var args []any
	if category != "" {
		query += " AND category = ?"
		args = append(args, category)
	}
	query += " ORDER BY category, preference"
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []CommunicationPreference
	for rows.Next() {
		var cp CommunicationPreference
		if err := rows.Scan(&cp.ID, &cp.Category, &cp.Preference, &cp.Value, &cp.Confidence, &cp.Source); err != nil {
			return nil, err
		}
		results = append(results, cp)
	}
	return results, nil
}

func (s *Store) SeedProfileFromMap(prefs map[string]map[string]string) error {
	for category, prefs := range prefs {
		for pref, value := range prefs {
			_, err := s.SetCommunicationPreference(SetCommunicationPreferenceParams{
				Category:   category,
				Preference: pref,
				Value:      value,
				Confidence: 0.9,
				Source:     "seed",
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}
