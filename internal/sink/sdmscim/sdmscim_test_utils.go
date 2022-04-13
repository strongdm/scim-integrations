package sdmscim

import (
	"context"
	"github.com/strongdm/scimsdk"
	scimmodels "github.com/strongdm/scimsdk/models"
)

type MockUserIterator struct {
	index  int
	buffer []*scimmodels.User
	err    error
}

type MockGroupIterator struct {
	index  int
	buffer []*scimmodels.Group
	err    error
}

func (m *MockUserIterator) Next() bool {
	if m.index < len(m.buffer)-1 {
		m.index++
		return true
	}
	return false
}

func (m *MockUserIterator) Value() *scimmodels.User {
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

func (m *MockGroupIterator) Value() *scimmodels.Group {
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

type MockUserModule struct {
	CreateFunc  func(context.Context, scimmodels.CreateUser) (*scimmodels.User, error)
	ListFunc    func(context.Context, *scimmodels.PaginationOptions) scimmodels.Iterator[scimmodels.User]
	FindFunc    func(context.Context, string) (*scimmodels.User, error)
	ReplaceFunc func(context.Context, string, scimmodels.ReplaceUser) (*scimmodels.User, error)
	UpdateFunc  func(context.Context, string, scimmodels.UpdateUser) (bool, error)
	DeleteFunc  func(context.Context, string) (bool, error)
}

func NewMockUserModule() scimsdk.UserModule {
	return &MockUserModule{}
}

func (mum *MockUserModule) Create(ctx context.Context, user scimmodels.CreateUser) (*scimmodels.User, error) {
	return mum.CreateFunc(ctx, user)
}

func (mum *MockUserModule) List(ctx context.Context, paginationOpts *scimmodels.PaginationOptions) scimmodels.Iterator[scimmodels.User] {
	return mum.ListFunc(ctx, paginationOpts)
}

func (mum *MockUserModule) Find(ctx context.Context, id string) (*scimmodels.User, error) {
	return mum.FindFunc(ctx, id)
}

func (mum *MockUserModule) Replace(ctx context.Context, id string, user scimmodels.ReplaceUser) (*scimmodels.User, error) {
	return mum.ReplaceFunc(ctx, id, user)
}

func (mum *MockUserModule) Update(ctx context.Context, id string, updateUser scimmodels.UpdateUser) (bool, error) {
	return mum.UpdateFunc(ctx, id, updateUser)
}

func (mum *MockUserModule) Delete(ctx context.Context, id string) (bool, error) {
	return mum.DeleteFunc(ctx, id)
}

type MockGroupModule struct {
	CreateFunc                 func(context.Context, scimmodels.CreateGroupBody) (*scimmodels.Group, error)
	ListFunc                   func(context.Context, *scimmodels.PaginationOptions) scimmodels.Iterator[scimmodels.Group]
	FindFunc                   func(context.Context, string) (*scimmodels.Group, error)
	ReplaceFunc                func(context.Context, string, scimmodels.ReplaceGroupBody) (*scimmodels.Group, error)
	UpdateAddMembersFunc       func(context.Context, string, []scimmodels.GroupMember) (bool, error)
	UpdateReplaceMembersFunc   func(context.Context, string, []scimmodels.GroupMember) (bool, error)
	UpdateReplaceNameFunc      func(context.Context, string, scimmodels.UpdateGroupReplaceName) (bool, error)
	UpdateRemoveMemberByIDFunc func(context.Context, string, string) (bool, error)
	DeleteFunc                 func(context.Context, string) (bool, error)
}

func NewMockGroupModule() scimsdk.GroupModule {
	return &MockGroupModule{}
}

func (mgm *MockGroupModule) Create(ctx context.Context, group scimmodels.CreateGroupBody) (*scimmodels.Group, error) {
	return mgm.CreateFunc(ctx, group)
}

func (mgm *MockGroupModule) List(ctx context.Context, paginationOptions *scimmodels.PaginationOptions) scimmodels.Iterator[scimmodels.Group] {
	return mgm.ListFunc(ctx, paginationOptions)
}

func (mgm *MockGroupModule) Find(ctx context.Context, id string) (*scimmodels.Group, error) {
	return mgm.FindFunc(ctx, id)
}

func (mgm *MockGroupModule) Replace(ctx context.Context, id string, group scimmodels.ReplaceGroupBody) (*scimmodels.Group, error) {
	return mgm.ReplaceFunc(ctx, id, group)
}

func (mgm *MockGroupModule) UpdateAddMembers(ctx context.Context, id string, members []scimmodels.GroupMember) (bool, error) {
	return mgm.UpdateAddMembersFunc(ctx, id, members)
}

func (mgm *MockGroupModule) UpdateReplaceMembers(ctx context.Context, id string, members []scimmodels.GroupMember) (bool, error) {
	return mgm.UpdateReplaceMembersFunc(ctx, id, members)
}

func (mgm *MockGroupModule) UpdateReplaceName(ctx context.Context, id string, replaceName scimmodels.UpdateGroupReplaceName) (bool, error) {
	return mgm.UpdateReplaceNameFunc(ctx, id, replaceName)
}

func (mgm *MockGroupModule) UpdateRemoveMemberByID(ctx context.Context, id string, memberID string) (bool, error) {
	return mgm.UpdateRemoveMemberByIDFunc(ctx, id, memberID)
}

func (mgm *MockGroupModule) Delete(ctx context.Context, id string) (bool, error) {
	return mgm.DeleteFunc(ctx, id)
}

type MockSDMSCIMClient struct {
	GetProvidedURLFunc func() string
	groupModule        scimsdk.GroupModule
	userModule         scimsdk.UserModule
}

func NewMockSDMSCIMClient(groupModule scimsdk.GroupModule, userModule scimsdk.UserModule) scimsdk.Client {
	mock := MockSDMSCIMClient{}
	mock.GetProvidedURLFunc = getProvidedURL
	mock.groupModule = groupModule
	mock.userModule = userModule
	return mock
}

func (mock MockSDMSCIMClient) Users() scimsdk.UserModule {
	return mock.userModule
}

func (mock MockSDMSCIMClient) Groups() scimsdk.GroupModule {
	return mock.groupModule
}

func (mock MockSDMSCIMClient) GetProvidedURL() string {
	return mock.GetProvidedURLFunc()
}

func getProvidedURL() string {
	return ""
}

func NewMockSDMSCIM(groupModule scimsdk.GroupModule, userModule scimsdk.UserModule) *SinkSDMSCIMImpl {
	mockClient := NewMockSDMSCIMClient(groupModule, userModule)
	mock := SinkSDMSCIMImpl{mockClient}
	return &mock
}
