package sdmscim

import (
	"github.com/strongdm/scimsdk/scimsdk"
)

type MockUserIterator struct {
	index  int
	buffer []*scimsdk.User
	err    error
}

type MockGroupIterator struct {
	index  int
	buffer []*scimsdk.Group
	err    error
}

func (m *MockUserIterator) Next() bool {
	if m.index < len(m.buffer)-1 {
		m.index++
		return true
	}
	return false
}

func (m *MockUserIterator) Value() *scimsdk.User {
	if m.index > len(m.buffer)-1 {
		return nil
	}
	return m.buffer[m.index]
}

func (m *MockUserIterator) IsEmpty() bool {
	return false
}

func (m *MockUserIterator) Err() error {
	return m.err
}

func (m *MockGroupIterator) Next() bool {
	if m.index < len(m.buffer)-1 {
		m.index++
		return true
	}
	return false
}

func (m *MockGroupIterator) Value() *scimsdk.Group {
	if m.index > len(m.buffer)-1 {
		return nil
	}
	return m.buffer[m.index]
}

func (m *MockGroupIterator) IsEmpty() bool {
	return false
}

func (m *MockGroupIterator) Err() error {
	return m.err
}
