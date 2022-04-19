package sdmscim

import (
	"context"
	"errors"
	"scim-integrations/internal/sink"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	scimmodels "github.com/strongdm/scimsdk/models"
)

func TestSDMSCIMFetchUsers(t *testing.T) {
	t.Run("should return a list of users with groups when executing the default flow", func(t *testing.T) {
		assertT := assert.New(t)
		groupModule := NewMockGroupModule()
		groupModule.(*MockGroupModule).ListFunc = mockedSCIMSDKGroupsList
		userModule := NewMockUserModule()
		userModule.(*MockUserModule).ListFunc = mockedSCIMSDKUsersList
		mock := NewMockSDMSCIM(groupModule, userModule)

		rows, err := mock.FetchUsers(context.Background())

		assertT.NotNil(rows)
		assertT.Nil(err)

		row := rows[0]

		assertT.NotNil(row)
		assertT.True(row.User.Active)
		assertT.NotEmpty(row.User.ID)
		assertT.NotEmpty(row.User.UserName)
		assertT.NotEmpty(row.User.DisplayName)
		assertT.NotEmpty(row.User.GivenName)
		assertT.NotEmpty(row.User.FamilyName)
		assertT.NotNil(row.Groups)

		group := row.Groups[0]

		assertT.NotEmpty(group.ID)
		assertT.NotEmpty(group.DisplayName)
		assertT.Len(group.Members, 1)
	})

	t.Run("should return an empty list of users when executing the default flow", func(t *testing.T) {
		groupModule := NewMockGroupModule()
		groupModule.(*MockGroupModule).ListFunc = mockedSCIMSDKEmptyGroupsList
		userModule := NewMockUserModule()
		userModule.(*MockUserModule).ListFunc = mockedSCIMSDKEmptyUsersList
		mock := NewMockSDMSCIM(groupModule, userModule)

		rows, err := mock.FetchUsers(context.Background())

		assert.Empty(t, rows)
		assert.Nil(t, err)
	})

	t.Run("should return a list of users without groups when executing the default flow", func(t *testing.T) {
		assertT := assert.New(t)
		groupModule := NewMockGroupModule()
		groupModule.(*MockGroupModule).ListFunc = mockedSCIMSDKEmptyGroupsList
		userModule := NewMockUserModule()
		userModule.(*MockUserModule).ListFunc = mockedSCIMSDKUsersList
		mock := NewMockSDMSCIM(groupModule, userModule)

		rows, err := mock.FetchUsers(context.Background())

		assertT.NotNil(rows)
		assertT.Nil(err)

		row := rows[0]

		assertT.NotNil(row)
		assertT.True(row.User.Active)
		assertT.NotEmpty(row.User.ID)
		assertT.NotEmpty(row.User.UserName)
		assertT.NotEmpty(row.User.DisplayName)
		assertT.NotEmpty(row.User.GivenName)
		assertT.NotEmpty(row.User.FamilyName)
		assertT.Empty(row.Groups)
	})

	t.Run("should return an error when using a context with timeout on fetch groups", func(t *testing.T) {
		assertT := assert.New(t)
		groupModule := NewMockGroupModule()
		groupModule.(*MockGroupModule).ListFunc = mockedSCIMSDKGroupListCTXTimeoutExceeded
		userModule := NewMockUserModule()
		userModule.(*MockUserModule).ListFunc = mockedSCIMSDKEmptyUsersList
		mock := NewMockSDMSCIM(groupModule, userModule)

		timeoutContext, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		users, err := mock.FetchUsers(timeoutContext)

		assertT.Empty(users)
		assertT.NotNil(err)
		assertT.NotNil(timeoutContext.Err())
		assertT.Contains(err.Error(), "deadline exceeded")
		assertT.Contains(timeoutContext.Err().Error(), "deadline exceeded")
	})

	t.Run("should return an error when using a context with timeout on list groups", func(t *testing.T) {
		assertT := assert.New(t)
		groupModule := NewMockGroupModule()
		groupModule.(*MockGroupModule).ListFunc = mockedSCIMSDKEmptyGroupsList
		userModule := NewMockUserModule()
		userModule.(*MockUserModule).ListFunc = mockedSCIMSDKUserListCTXTimeoutExceeded
		mock := NewMockSDMSCIM(groupModule, userModule)

		timeoutContext, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		users, err := mock.FetchUsers(timeoutContext)

		assertT.Nil(users)
		assertT.NotNil(err)
		assertT.NotNil(timeoutContext.Err())
		assertT.Contains(err.Error(), "deadline exceeded")
		assertT.Contains(timeoutContext.Err().Error(), "deadline exceeded")
	})
}

func TestSDMSCIMCreateUser(t *testing.T) {
	t.Run("should create an user when passed a valid IdP user", func(t *testing.T) {
		userModule := NewMockUserModule()
		userModule.(*MockUserModule).CreateFunc = mockedSCIMSDKUserCreate
		mock := NewMockSDMSCIM(nil, userModule)

		response, err := mock.CreateUser(context.Background(), getMockUserRow())

		assert.NotNil(t, response)
		assert.Nil(t, err)
	})

	t.Run("should return an error when the context timeout exceeds", func(t *testing.T) {
		assertT := assert.New(t)
		userModule := NewMockUserModule()
		userModule.(*MockUserModule).CreateFunc = mockedSCIMSDKUserCreateCTXTimeoutExceeded
		mock := NewMockSDMSCIM(nil, userModule)

		timeoutContext, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		response, err := mock.CreateUser(timeoutContext, getMockUserRow())

		assertT.Nil(response)
		assertT.NotNil(err)
		assertT.NotNil(timeoutContext.Err())
		assertT.Contains(err.Error(), "deadline exceeded")
		assertT.Contains(timeoutContext.Err().Error(), "deadline exceeded")
	})
}

func TestSDMSCIMReplaceUser(t *testing.T) {
	t.Run("should create an user when passed a valid IdP user", func(t *testing.T) {
		userModule := NewMockUserModule()
		userModule.(*MockUserModule).ReplaceFunc = mockedSCIMSDKUserReplace
		mock := NewMockSDMSCIM(nil, userModule)

		err := mock.ReplaceUser(context.Background(), *getMockUserRow())

		assert.Nil(t, err)
	})

	t.Run("should return an error when the context timeout exceeds", func(t *testing.T) {
		assertT := assert.New(t)
		userModule := NewMockUserModule()
		userModule.(*MockUserModule).ReplaceFunc = mockedSCIMSDKUserReplaceCTXTimeoutExceeded
		mock := NewMockSDMSCIM(nil, userModule)

		timeoutContext, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		err := mock.ReplaceUser(timeoutContext, *getMockUserRow())

		assertT.NotNil(err)
		assertT.NotNil(timeoutContext.Err())
		assertT.Contains(err.Error(), "deadline exceeded")
		assertT.Contains(timeoutContext.Err().Error(), "deadline exceeded")
	})
}

func TestSDMSCIMDeleteUser(t *testing.T) {
	t.Run("should delete an user when passed a valid user ID", func(t *testing.T) {
		userModule := NewMockUserModule()
		userModule.(*MockUserModule).DeleteFunc = mockedSCIMSDKUserDelete
		mock := NewMockSDMSCIM(nil, userModule)

		err := mock.DeleteUser(context.Background(), *getMockUserRow())

		assert.Nil(t, err)
	})

	t.Run("should return an error when the context timeout exceeds", func(t *testing.T) {
		assertT := assert.New(t)
		userModule := NewMockUserModule()
		userModule.(*MockUserModule).DeleteFunc = mockedSCIMSDKUserDeleteCTXTimeoutExceeded
		mock := NewMockSDMSCIM(nil, userModule)

		timeoutContext, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		err := mock.DeleteUser(timeoutContext, *getMockUserRow())

		assertT.NotNil(err)
		assertT.NotNil(timeoutContext.Err())
		assertT.Contains(err.Error(), "deadline exceeded")
		assertT.Contains(timeoutContext.Err().Error(), "deadline exceeded")
	})
}

func TestSDMSCIMFetchGroups(t *testing.T) {
	t.Run("should return a list of groups when executing the default flow", func(t *testing.T) {
		assertT := assert.New(t)
		groupModule := NewMockGroupModule()
		groupModule.(*MockGroupModule).ListFunc = mockedSCIMSDKGroupsList
		mock := NewMockSDMSCIM(groupModule, nil)

		groups, err := mock.FetchGroups(context.Background())

		assertT.NotNil(groups)
		assertT.Nil(err)

		group := groups[0]

		assertT.NotEmpty(group.ID)
		assertT.NotEmpty(group.DisplayName)
		assertT.NotNil(group.Members)

		member := group.Members[0]

		assertT.NotEmpty(member.ID)
		assertT.NotEmpty(member.Email)
	})

	t.Run("should return an empty list of groups when executing the default flow", func(t *testing.T) {
		groupModule := NewMockGroupModule()
		groupModule.(*MockGroupModule).ListFunc = mockedSCIMSDKEmptyGroupsList
		mock := NewMockSDMSCIM(groupModule, nil)

		groups, err := mock.FetchGroups(context.Background())

		assert.Empty(t, groups)
		assert.Nil(t, err)
	})

	t.Run("should return an error when the context timeout exceeds", func(t *testing.T) {
		assertT := assert.New(t)
		groupModule := NewMockGroupModule()
		groupModule.(*MockGroupModule).ListFunc = mockedSCIMSDKGroupListCTXTimeoutExceeded
		mock := NewMockSDMSCIM(groupModule, nil)

		timeoutContext, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		groups, err := mock.FetchGroups(timeoutContext)

		assertT.Nil(groups)
		assertT.NotNil(err)
		assertT.NotNil(timeoutContext.Err())
		assertT.Contains(err.Error(), "deadline exceeded")
		assertT.Contains(timeoutContext.Err().Error(), "deadline exceeded")
	})
}

func TestSDMSCIMCreateGroup(t *testing.T) {
	t.Run("should create a group when executing the default flow", func(t *testing.T) {
		groupModule := NewMockGroupModule()
		groupModule.(*MockGroupModule).CreateFunc = mockedSCIMSDKGroupCreate
		mock := NewMockSDMSCIM(groupModule, nil)

		response, err := mock.CreateGroup(context.Background(), getMockGroupRow())

		assert.NotNil(t, response)
		assert.Nil(t, err)
	})

	t.Run("should return an error when the context timeout exceeds", func(t *testing.T) {
		assertT := assert.New(t)
		groupModule := NewMockGroupModule()
		groupModule.(*MockGroupModule).CreateFunc = mockedSCIMSDKGroupCreateCTXTimeoutExceeded
		mock := NewMockSDMSCIM(groupModule, nil)

		timeoutContext, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		response, err := mock.CreateGroup(timeoutContext, getMockGroupRow())

		assertT.Nil(response)
		assertT.NotNil(err)
		assertT.NotNil(timeoutContext.Err())
		assertT.Contains(err.Error(), "deadline exceeded")
		assertT.Contains(timeoutContext.Err().Error(), "deadline exceeded")
	})
}

func TestSDMSCIMReplaceMembers(t *testing.T) {
	t.Run("should replace members when executing the default flow", func(t *testing.T) {
		groupModule := NewMockGroupModule()
		groupModule.(*MockGroupModule).UpdateReplaceMembersFunc = mockedSCIMSDKGroupReplaceMembers
		mock := NewMockSDMSCIM(groupModule, nil)

		err := mock.ReplaceGroupMembers(context.Background(), getMockGroupRow())

		assert.Nil(t, err)
	})

	t.Run("should return an error when the context timeout exceeds", func(t *testing.T) {
		assertT := assert.New(t)

		groupModule := NewMockGroupModule()
		groupModule.(*MockGroupModule).UpdateReplaceMembersFunc = mockedSCIMSDKGroupReplaceMembersCTXTimeoutExceeded
		mock := NewMockSDMSCIM(groupModule, nil)

		timeoutContext, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		err := mock.ReplaceGroupMembers(timeoutContext, getMockGroupRow())

		assertT.NotNil(err)
		assertT.NotNil(timeoutContext.Err())
		assertT.Contains(err.Error(), "deadline exceeded")
		assertT.Contains(timeoutContext.Err().Error(), "deadline exceeded")
	})
}

func TestSDMSCIMDeleteGroup(t *testing.T) {
	t.Run("should delete a group when passed a valid group ID", func(t *testing.T) {
		groupModule := NewMockGroupModule()
		groupModule.(*MockGroupModule).DeleteFunc = mockedSCIMSDKGroupDelete
		mock := NewMockSDMSCIM(groupModule, nil)

		err := mock.DeleteGroup(context.Background(), getMockGroupRow())

		assert.Nil(t, err)
	})

	t.Run("should return an error when the context timeout exceeds", func(t *testing.T) {
		assertT := assert.New(t)
		groupModule := NewMockGroupModule()
		groupModule.(*MockGroupModule).DeleteFunc = mockedSCIMSDKGroupDeleteCTXTimeoutExceeded
		mock := NewMockSDMSCIM(groupModule, nil)

		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		err := mock.DeleteGroup(ctx, getMockGroupRow())

		assertT.NotNil(err)
		assertT.NotNil(ctx.Err())
		assertT.Contains(err.Error(), "deadline exceeded")
		assertT.Contains(ctx.Err().Error(), "deadline exceeded")
	})
}

func mockedSCIMSDKUsersList(_ context.Context, _ *scimmodels.PaginationOptions) scimmodels.Iterator[scimmodels.User] {
	return &MockUserIterator{
		buffer: []*scimmodels.User{
			getMockSCIMSDKUser(),
			getMockSCIMSDKUser(),
		},
		index: -1,
	}
}

func mockedSCIMSDKEmptyUsersList(_ context.Context, _ *scimmodels.PaginationOptions) scimmodels.Iterator[scimmodels.User] {
	return &MockUserIterator{
		buffer: []*scimmodels.User{},
		index:  -1,
	}
}

func mockedSCIMSDKUserListCTXTimeoutExceeded(_ context.Context, _ *scimmodels.PaginationOptions) scimmodels.Iterator[scimmodels.User] {
	time.Sleep(time.Millisecond * 2)
	return &MockUserIterator{err: errors.New("context deadline exceeded")}
}

func mockedSCIMSDKUserCreate(_ context.Context, _ scimmodels.CreateUser) (*scimmodels.User, error) {
	return &scimmodels.User{
		Name: &scimmodels.UserName{},
	}, nil
}

func mockedSCIMSDKUserCreateCTXTimeoutExceeded(_ context.Context, _ scimmodels.CreateUser) (*scimmodels.User, error) {
	time.Sleep(time.Millisecond * 2)
	return nil, errors.New("context deadline exceeded")
}

func mockedSCIMSDKUserReplace(_ context.Context, _ string, _ scimmodels.ReplaceUser) (*scimmodels.User, error) {
	return &scimmodels.User{}, nil
}

func mockedSCIMSDKUserReplaceCTXTimeoutExceeded(_ context.Context, _ string, _ scimmodels.ReplaceUser) (*scimmodels.User, error) {
	time.Sleep(time.Millisecond * 2)
	return nil, errors.New("context deadline exceeded")
}

func mockedSCIMSDKUserDelete(_ context.Context, _ string) (bool, error) {
	return true, nil
}

func mockedSCIMSDKUserDeleteCTXTimeoutExceeded(_ context.Context, _ string) (bool, error) {
	time.Sleep(time.Millisecond * 2)
	return false, errors.New("context deadline exceeded")
}

func mockedSCIMSDKGroupsList(_ context.Context, _ *scimmodels.PaginationOptions) scimmodels.Iterator[scimmodels.Group] {
	return &MockGroupIterator{
		buffer: []*scimmodels.Group{
			getMockSCIMSDKGroup(),
			getMockSCIMSDKGroup(),
		},
		index: -1,
	}
}

func mockedSCIMSDKEmptyGroupsList(_ context.Context, _ *scimmodels.PaginationOptions) scimmodels.Iterator[scimmodels.Group] {
	return &MockGroupIterator{
		buffer: []*scimmodels.Group{},
		index:  -1,
	}
}

func mockedSCIMSDKGroupListCTXTimeoutExceeded(_ context.Context, _ *scimmodels.PaginationOptions) scimmodels.Iterator[scimmodels.Group] {
	time.Sleep(time.Millisecond * 2)
	return &MockGroupIterator{err: errors.New("context deadline exceeded")}
}

func mockedSCIMSDKGroupCreate(_ context.Context, _ scimmodels.CreateGroupBody) (*scimmodels.Group, error) {
	return &scimmodels.Group{}, nil
}

func mockedSCIMSDKGroupCreateCTXTimeoutExceeded(_ context.Context, _ scimmodels.CreateGroupBody) (*scimmodels.Group, error) {
	time.Sleep(time.Millisecond * 2)
	return nil, errors.New("context deadline exceeded")
}

func mockedSCIMSDKGroupReplaceMembers(_ context.Context, _ string, _ []scimmodels.GroupMember) (bool, error) {
	return true, nil
}

func mockedSCIMSDKGroupReplaceMembersCTXTimeoutExceeded(_ context.Context, _ string, _ []scimmodels.GroupMember) (bool, error) {
	time.Sleep(time.Millisecond * 2)
	return false, errors.New("context deadline exceeded")
}

func mockedSCIMSDKGroupDelete(_ context.Context, _ string) (bool, error) {
	return true, nil
}

func mockedSCIMSDKGroupDeleteCTXTimeoutExceeded(_ context.Context, _ string) (bool, error) {
	time.Sleep(time.Millisecond * 2)
	return false, errors.New("context deadline exceeded")
}

func getMockSCIMSDKUser() *scimmodels.User {
	return &scimmodels.User{
		ID:          "www",
		Active:      true,
		DisplayName: "xxx",
		Emails: []scimmodels.UserEmail{
			{
				Primary: true,
				Value:   "yyy",
			},
		},
		Groups: []scimmodels.UserGroupReference{
			{
				Value: "yyy",
				Ref:   "xxx",
			},
		},
		Name: &scimmodels.UserName{
			FamilyName: "ccc",
			Formatted:  "ddd",
			GivenName:  "eee",
		},
		UserName: "yyy",
		UserType: "zzz",
	}
}

func getMockSCIMSDKGroup() *scimmodels.Group {
	return &scimmodels.Group{
		ID:          "xxx",
		DisplayName: "yyy",
		Members: []*scimmodels.GroupMember{
			{
				ID:    "www",
				Email: "yyy",
			},
		},
		Meta: &scimmodels.GroupMetadata{
			ResourceType: "www",
			Location:     "zzz",
		},
	}
}

func getMockUserRow() *sink.UserRow {
	return &sink.UserRow{
		User: &sink.User{
			ID:       "xxx",
			UserName: "sdm@test.com",
		},
	}
}

func getMockGroupRow() *sink.GroupRow {
	return &sink.GroupRow{
		ID:          "xxx",
		DisplayName: "yyy",
		Members: []*sink.GroupMember{
			{
				ID:          "xxx",
				Email:       "yyy",
				SDMObjectID: "zzz",
			},
		},
	}
}
