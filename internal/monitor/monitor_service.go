package monitor

import (
	"database/sql"
	"net/url"
	"strings"
	"time"
	"watchdog/main.go/internal/apperrors"

	"github.com/google/uuid"
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
	if err := validateMonitor(monitor); err != nil {
		return -1, err
	}

	if strings.TrimSpace(monitor.ID) == "" {
		monitor.ID = uuid.New().String()
	}

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
		return -1, apperrors.New(apperrors.INTERNAL_ERROR, "failed to insert monitor", err)
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return -1, apperrors.New(apperrors.INTERNAL_ERROR, "failed to get inserted monitor id", err)
	}

	return lastInsertID, nil
}

func (m MonitorService) UpdateMonitor(monitor MonitorDTO) *apperrors.AppError {
	if strings.TrimSpace(monitor.ID) == "" {
		return apperrors.New(apperrors.VALIDATION_ERROR, "monitor id is required", nil)
	}

	if err := validateMonitor(monitor); err != nil {
		return err
	}

	result, dbErr := m.db.Exec(`
        UPDATE monitors SET
            name = ?,
            url = ?,
            method = ?,
            interval_seconds = ?,
            timeout_seconds = ?,
            expected_status = ?,
            expected_body_substring = ?,
            headers_json = ?,
            enabled = ?,
            updated_at = ?
        WHERE id = ?
    `,
		monitor.Name,
		monitor.URL,
		monitor.Method,
		monitor.IntervalSeconds,
		monitor.TimeoutSeconds,
		monitor.ExpectedStatus,
		toNullString(monitor.ExpectedBodySubstring),
		toNullString(monitor.HeadersJSON),
		monitor.Enabled,
		time.Now(),
		monitor.ID,
	)

	if dbErr != nil {
		return apperrors.New(apperrors.INTERNAL_ERROR, "failed to update monitor", dbErr)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.New(apperrors.INTERNAL_ERROR, "failed to determine updated monitors", err)
	}

	if rowsAffected == 0 {
		return apperrors.New(apperrors.NOT_FOUND, "monitor not found", nil)
	}

	return nil
}

func (m MonitorService) DeleteMonitor(monitor MonitorDTO) *apperrors.AppError {
	return m.DeleteMonitorByID(monitor.ID)
}

func (m MonitorService) DeleteMonitorByID(id string) *apperrors.AppError {
	result, err := m.db.Exec("DELETE FROM monitors WHERE id = ?", id)
	if err != nil {
		return apperrors.New(apperrors.INTERNAL_ERROR, "failed to remove monitor", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.New(apperrors.INTERNAL_ERROR, "failed to determine removed monitors", err)
	}

	if rowsAffected == 0 {
		return apperrors.New(apperrors.NOT_FOUND, "monitor not found", nil)
	}

	return nil

}

func validateMonitor(m MonitorDTO) *apperrors.AppError {

	if strings.TrimSpace(m.Name) == "" {
		return apperrors.New(apperrors.VALIDATION_ERROR, "monitor name is required", nil)
	}
	if strings.TrimSpace(m.URL) == "" {
		return apperrors.New(apperrors.VALIDATION_ERROR, "monitor url is required", nil)
	}
	if _, err := url.ParseRequestURI(m.URL); err != nil {
		return apperrors.New(apperrors.VALIDATION_ERROR, "monitor url is invalid", err)
	}

	switch strings.ToUpper(strings.TrimSpace(m.Method)) {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS":
	default:
		return apperrors.New(apperrors.VALIDATION_ERROR, "monitor method is invalid", nil)
	}

	if m.IntervalSeconds <= 0 {
		return apperrors.New(apperrors.VALIDATION_ERROR, "interval_seconds must be greater than 0", nil)
	}
	if m.TimeoutSeconds <= 0 {
		return apperrors.New(apperrors.VALIDATION_ERROR, "timeout_seconds must be greater than 0", nil)
	}
	if m.ExpectedStatus < 100 || m.ExpectedStatus > 599 {
		return apperrors.New(apperrors.VALIDATION_ERROR, "expected_status must be between 100 and 599", nil)
	}
	if strings.TrimSpace(m.CreatedAt) == "" {
		return apperrors.New(apperrors.VALIDATION_ERROR, "created_at is required", nil)
	}
	if strings.TrimSpace(m.UpdatedAt) == "" {
		return apperrors.New(apperrors.VALIDATION_ERROR, "updated_at is required", nil)
	}

	return nil
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
