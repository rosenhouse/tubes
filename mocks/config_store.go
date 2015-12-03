package mocks

func NewConfigStore() *ConfigStore {
	return &ConfigStore{}
}

type ConfigStore struct {
	GetCall struct {
		Receives struct {
			Key string
		}
		Returns struct {
			Value []byte
			Error error
		}
	}

	SetCall struct {
		Receives struct {
			Key   string
			Value []byte
		}
		Returns struct {
			Error error
		}
	}
}

func (s *ConfigStore) Get(key string) ([]byte, error) {
	s.GetCall.Receives.Key = key
	return s.GetCall.Returns.Value, s.GetCall.Returns.Error
}

func (s *ConfigStore) Set(key string, value []byte) error {
	s.SetCall.Receives.Key = key
	s.SetCall.Receives.Value = value
	return s.SetCall.Returns.Error
}

func NewFunctionalConfigStore() *FunctionalConfigStore {
	return &FunctionalConfigStore{
		Values: make(map[string][]byte),
		Errors: make(map[string]error),
	}
}

type FunctionalConfigStore struct {
	Values       map[string][]byte
	Errors       map[string]error
	IsEmptyError error
}

func (s *FunctionalConfigStore) Get(key string) ([]byte, error) {
	return s.Values[key], s.Errors[key]
}

func (s *FunctionalConfigStore) Set(key string, value []byte) error {
	s.Values[key] = value
	return s.Errors[key]
}

func (s *FunctionalConfigStore) IsEmpty() (bool, error) {
	return len(s.Values) == 0, s.IsEmptyError
}
