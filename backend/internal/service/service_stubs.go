package service

import "time"

func (s *Service) ExecutiveCatalog() []map[string]any {
	return []map[string]any{
		{
			"id": "exec-1", "name": "Alice Chen", "title": "Chief Operating Officer",
			"skills": map[string]any{"management": 75, "finance": 60, "science": 20},
			"salary": 5000, "availability": "immediate",
		},
		{
			"id": "exec-2", "name": "Bob Smith", "title": "Chief Technology Officer",
			"skills": map[string]any{"management": 40, "science": 85, "finance": 30},
			"salary": 6000, "availability": "immediate",
		},
		{
			"id": "exec-3", "name": "Carol Davis", "title": "Chief Financial Officer",
			"skills": map[string]any{"management": 50, "finance": 80, "science": 25},
			"salary": 5500, "availability": "in_contract",
		},
	}
}

func (s *Service) RecruitExecutive(execID string) map[string]any {
	return map[string]any{
		"ok": true, "executiveId": execID,
		"moneyDeducted": 5000,
		"joinedAt":      s.now().UTC().Format(time.RFC3339),
	}
}

func (s *Service) TrainExecutive(execID string) map[string]any {
	return map[string]any{
		"ok": true, "executiveId": execID,
		"skillGained": "management", "newLevel": 76,
	}
}

func (s *Service) PoachExecutive(_ map[string]any) map[string]any {
	return map[string]any{"ok": false, "reason": "poach_failed"}
}

func (s *Service) ExecutiveDetail(execID string) (map[string]any, error) {
	return map[string]any{
		"id": execID, "name": "Alice Chen", "title": "Chief Operating Officer",
		"skills": map[string]any{"management": 75, "finance": 60, "science": 20},
		"salary": 5000, "morale": 85, "contractLength": 30,
	}, nil
}

func (s *Service) IncomingOffers() []map[string]any {
	return []map[string]any{}
}

func (s *Service) RespondToOffer(_ map[string]any) map[string]any {
	return map[string]any{"ok": true}
}

func (s *Service) RocketProjects() []map[string]any {
	return []map[string]any{
		{
			"id": "rocket-1", "name": "Scout I", "type": "sounding_rocket",
			"status": "available", "cost": 100000.0,
			"payloadCapacity": 50, "range": "suborbital",
		},
		{
			"id": "rocket-2", "name": "Orbiter I", "type": "orbital",
			"status": "in_progress", "progress": 60, "cost": 500000.0,
			"payloadCapacity": 500, "range": "leo",
		},
		{
			"id": "rocket-3", "name": "Explorer II", "type": "deep_space",
			"status": "locked", "progress": 0, "cost": 2000000.0,
			"payloadCapacity": 2000, "range": "interplanetary",
		},
	}
}

func (s *Service) CreateRocketProject(_ map[string]any) map[string]any {
	return map[string]any{
		"ok": true,
		"project": map[string]any{
			"id": "rocket-4", "name": "New Rocket", "type": "sounding_rocket",
			"status": "in_progress", "progress": 0,
			"createdAt": s.now().UTC().Format(time.RFC3339),
		},
	}
}

func (s *Service) LaunchHistory() []map[string]any {
	return []map[string]any{
		{
			"id": "launch-1", "rocket": "Scout I",
			"status": "success", "payload": "Atmosphere Probe",
			"launchedAt": s.now().Add(-72 * time.Hour).UTC().Format(time.RFC3339),
		},
	}
}

func (s *Service) LaunchRocket(_ map[string]any) map[string]any {
	return map[string]any{
		"ok": true, "launchId": "launch-2", "status": "launched",
		"launchedAt": s.now().UTC().Format(time.RFC3339),
	}
}

func (s *Service) AvailableComponents() []map[string]any {
	return []map[string]any{
		{"id": "comp-1", "name": "Rocket Fuel", "resourceId": 83, "quantity": 1000},
		{"id": "comp-2", "name": "Solid Rocket", "resourceId": 85, "quantity": 500},
		{"id": "comp-3", "name": "Rocket Engine", "resourceId": 86, "quantity": 100},
		{"id": "comp-4", "name": "Ion Drive", "resourceId": 88, "quantity": 25},
		{"id": "comp-5", "name": "Guidance System", "resourceId": 90, "quantity": 50},
	}
}
