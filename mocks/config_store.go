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
