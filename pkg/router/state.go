package router

import "sync"

type State struct {
	current sync.Map
}

//func (sm *State) Expect(id int64, what string) {
//	sm.state.Store(id, what)
//}

func (s *State) SetState(id int64, st string) {
	s.current.Store(id, st)
}

func (s *State) GetState(id int64) string {
	if val, ok := s.current.Load(id); ok {
		return val.(string)
	}
	return ""
}

func (sm *State) Clear(id int64) {
	sm.current.Delete(id)
}

//func (sm *State) Delete(id int64) {
//	sm.Store(id, nil)
//}
//
//func (sm *State) Store(id int64, val interface{}) {
//	if val == nil {
//		sm.data.Delete(id)
//	} else {
//		sm.data.Store(id, val)
//	}
//}
//
//func (sm *State) Load(id int64) interface{} {
//	if val, ok := sm.data.Load(id); ok {
//		return val
//	}
//	return nil
//}
