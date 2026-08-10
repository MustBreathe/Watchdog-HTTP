package monitor

import "database/sql"

type MonitorService struct {
	db *sql.DB
}

func CreateMonitorService(db *sql.DB) *MonitorService {
	return &MonitorService{
		db: db,
	}
}

func (s MonitorService) GetMonitors() ([]MonitorDTO, error) {
	rows, err := s.db.Query("SELECT id, name, url FROM monitors")

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	monitors := make([]MonitorDTO, 0)

	for rows.Next() {
		var monitor MonitorDTO
		if err := rows.Scan(&monitor.ID, &monitor.Name, &monitor.URL); err != nil {
			return nil, err
		}
		monitors = append(monitors, monitor)
	}

	return monitors, rows.Err()
}
