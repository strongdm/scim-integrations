package idp

type MockIdP struct{}

func NewMockIdP() *MockIdP {
	return &MockIdP{}
}

func (MockIdP) Fetch() ([]map[string]interface{}, error) {
	data := []map[string]interface{}{
		{
			"GivenName":  "My",
			"FamilyName": "Name",
			"UserName":   "username@example.com",
		},
	}
	return data, nil
}
