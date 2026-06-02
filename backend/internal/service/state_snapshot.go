package service

import "go-sim-api/internal/model"

func (s *Service) Snapshot() model.GameState {
	s.mu.Lock()
	defer s.mu.Unlock()
	return cloneState(s.State)
}

func (s *Service) UpdatePreferences(prefs map[string]any) map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.State.PlayerPreferences == nil {
		s.State.PlayerPreferences = map[string]any{}
	}
	for k, v := range prefs {
		s.State.PlayerPreferences[k] = v
	}
	out := cloneMapStringAny(s.State.PlayerPreferences)
	s.saveStateLocked()
	return out
}

func (s *Service) AddMessage(msg model.Message) model.Message {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.State.Messages = append([]model.Message{msg}, s.State.Messages...)
	s.saveStateLocked()
	return msg
}

func cloneState(in model.GameState) model.GameState {
	out := in
	out.Companies = cloneCompanies(in.Companies)
	out.BotCompanies = cloneCompanies(in.BotCompanies)
	out.Orders = append([]model.MarketOrder(nil), in.Orders...)
	out.Trades = append([]model.Trade(nil), in.Trades...)
	out.Messages = append([]model.Message(nil), in.Messages...)
	out.Bonds = append([]model.Bond(nil), in.Bonds...)
	out.Notifications = append([]model.Notification(nil), in.Notifications...)
	out.ProductionJobs = append([]model.ProductionJob(nil), in.ProductionJobs...)
	out.GovernmentContracts = append([]model.GovContract(nil), in.GovernmentContracts...)
	out.Ledger = append([]model.LedgerEntry(nil), in.Ledger...)
	out.Auctions = append([]model.Auction(nil), in.Auctions...)
	out.DailyOrders = append([]model.Order(nil), in.DailyOrders...)
	out.PlayerPreferences = cloneMapStringAny(in.PlayerPreferences)
	out.Achievements = cloneSliceMapStringAny(in.Achievements)
	out.ContractsIn = cloneSliceMapStringAny(in.ContractsIn)
	out.ContractsOut = cloneSliceMapStringAny(in.ContractsOut)
	out.Executives = cloneSliceMapStringAny(in.Executives)
	return out
}

func cloneCompany(in model.Company) model.Company {
	out := in
	out.Inventory = cloneMapIntInt(in.Inventory)
	out.QualityInventory = cloneMapStringInt(in.QualityInventory)
	out.PlacedBuildings = cloneSliceMapStringAny(in.PlacedBuildings)
	out.UnplacedBuildings = cloneSliceMapStringAny(in.UnplacedBuildings)
	return out
}

func cloneCompanies(in []model.Company) []model.Company {
	out := make([]model.Company, len(in))
	for i := range in {
		out[i] = cloneCompany(in[i])
	}
	return out
}

func cloneMapIntInt(in map[int]int) map[int]int {
	if in == nil {
		return nil
	}
	out := make(map[int]int, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneMapStringInt(in map[string]int) map[string]int {
	if in == nil {
		return nil
	}
	out := make(map[string]int, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneMapStringAny(in map[string]any) map[string]any {
	if in == nil {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneSliceMapStringAny(in []map[string]any) []map[string]any {
	if in == nil {
		return nil
	}
	out := make([]map[string]any, len(in))
	for i := range in {
		out[i] = cloneMapStringAny(in[i])
	}
	return out
}
