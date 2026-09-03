package monitor

import (
	"database/sql"
	"watchdog/main.go/internal/apperrors"
)

type MonitorService struct {
	db *sql.DB
}

func CreateMonitorService(db *sql.DB) *MonitorService {
	return &MonitorService{
		db: db,
	}
}

func (s MonitorService) GetMonitors() ([]MonitorDTO, *apperrors.AppError) {
	rows, err := s.db.Query(`SELECT
            id,
            name,
            url,
            method,
            interval_seconds,
            timeout_seconds,
            expected_status,
            expected_body_substring,
            headers_json,
            enabled,
            created_at,
            updated_at
        FROM monitors`)

	if err != nil {
		return nil, apperrors.New(apperrors.INTERNAL_ERROR, err.Error(), err)
	}
	defer rows.Close()

	monitors := make([]MonitorDTO, 0)
	var expectedBody sql.NullString
	var headersJSON sql.NullString

	for rows.Next() {
		var monitor MonitorDTO
		if err := rows.Scan(&monitor.ID,
			&monitor.Name,
			&monitor.URL,
			&monitor.Method,
			&monitor.IntervalSeconds,
			&monitor.TimeoutSeconds,
			&monitor.ExpectedStatus,
			&expectedBody,
			&headersJSON,
			&monitor.Enabled,
			&monitor.CreatedAt,
			&monitor.UpdatedAt); err != nil {
			return nil, apperrors.New(apperrors.INTERNAL_ERROR, err.Error(), err)
		}

		if expectedBody.Valid {
			monitor.ExpectedBodySubstring = &expectedBody.String
		}

		if headersJSON.Valid {
			monitor.HeadersJSON = &headersJSON.String
		}

		monitors = append(monitors, monitor)
	}

	if err := rows.Err(); err != nil {
		return nil, apperrors.New(apperrors.INTERNAL_ERROR, err.Error(), err)
	}
	return monitors, nil
}

func (m MonitorService) RetrieveMonitorByID(id string) (*MonitorDTO, *apperrors.AppError) {
	row := m.db.QueryRow(`SELECT
            id,
            name,
            url,
            method,
            interval_seconds,
            timeout_seconds,
            expected_status,
            expected_body_substring,
            headers_json,
            enabled,
            created_at,
            updated_at
        FROM monitors WHERE id = ?`, id)

	var (
		monitor      MonitorDTO
		expectedBody sql.NullString
		headersJSON  sql.NullString
	)

	err := row.Scan(&monitor.ID,
		&monitor.Name,
		&monitor.URL,
		&monitor.Method,
		&monitor.IntervalSeconds,
		&monitor.TimeoutSeconds,
		&monitor.ExpectedStatus,
		&expectedBody,
		&headersJSON,
		&monitor.Enabled,
		&monitor.CreatedAt,
		&monitor.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, apperrors.New(apperrors.NOT_FOUND, "monitor not found", err)
	}

	if err != nil {
		return nil, apperrors.New(apperrors.INTERNAL_ERROR, err.Error(), err)
	}

	if expectedBody.Valid {
		monitor.ExpectedBodySubstring = &expectedBody.String
	}

	if headersJSON.Valid {
		monitor.HeadersJSON = &headersJSON.String
	}

	return &monitor, nil
}

func (m MonitorService) CreateMonitor(monitor MonitorDTO) (int64, *apperrors.AppError) {
	result, err := m.db.Exec(`
        INSERT INTO monitors (
            id,
            name,
            url,
            method,
            interval_seconds,
            timeout_seconds,
            expected_status,
            expected_body_substring,
            headers_json,
            enabled,
            created_at,
            updated_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `,
		monitor.ID,
		monitor.Name,
		monitor.URL,
		monitor.Method,
		monitor.IntervalSeconds,
		monitor.TimeoutSeconds,
		monitor.ExpectedStatus,
		toNullString(monitor.ExpectedBodySubstring),
		toNullString(monitor.HeadersJSON),
		monitor.Enabled,
		monitor.CreatedAt,
		monitor.UpdatedAt,
	)

	if err != nil {
		return -1, apperrors.New(apperrors.INTERNAL_ERROR, "unable to insert data ", err)
	}

	lastInsertID, err := result.LastInsertId()

	if err != nil {
		return -1, apperrors.New(apperrors.INTERNAL_ERROR, "unable determin id", err)
	}

	return lastInsertID, nil

}

func toNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{
		String: *s,
		Valid:  true,
	}
}
