package monitor

import "database/sql"

ty

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

	monitors := make([]MonitorDTO, rows.length)


	for i:=0; rows.Next(); i++ {
		var monitor MonitorDTO
		if err := rows.Scan(&monitor.ID, &monitor.Name, &monitor.URL); err != nil {
			return nil, err
		}
		monitors := append(monitors, monitor)
	}


	return err
}