package sdmscim

import (
	"context"
	"errors"
	"scim-integrations/internal/sink"
	"scim-integrations/internal/source"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"github.com/strongdm/scimsdk/scimsdk"
)

func TestSDMSCIMFetchUsers(t *testing.T) {
	t.Run("should return a empty list of users when executing the default flow", func(t *testing.T) {
		defer monkey.UnpatchAll()
		monkey.Patch(internalSCIMSDKGroupsList, mockedInternalSCIMSDKEmptyGroupsList)
		monkey.Patch(internalSCIMSDKUsersList, mockedInternalSCIMSDKEmptyUsersList)

		users, err := FetchUsers(context.Background())

		assert.Empty(t, users)
		assert.Nil(t, err)
	})

	t.Run("should return a list of users when executing the default flow", func(t *testing.T) {
		defer monkey.UnpatchAll()
		monkey.Patch(internalSCIMSDKGroupsList, mockedInternalSCIMSDKGroupsList)
		monkey.Patch(internalSCIMSDKUsersList, mockedInternalSCIMSDKUsersList)

		assertT := assert.New(t)

		rows, err := FetchUsers(context.Background())

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

	t.Run("should return a list of users without groups when executing the default flow", func(t *testing.T) {
		defer monkey.UnpatchAll()
		monkey.Patch(internalSCIMSDKGroupsList, mockedInternalSCIMSDKEmptyGroupsList)
		monkey.Patch(internalSCIMSDKUsersList, mockedInternalSCIMSDKUsersList)

		assertT := assert.New(t)

		rows, err := FetchUsers(context.Background())

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

	t.Run("should return a empty list of users when using a context with timeout", func(t *testing.T) {
		defer monkey.UnpatchAll()
		monkey.Patch(internalSCIMSDKGroupsList, mockedInternalSCIMSDKEmptyGroupsList)
		monkey.Patch(internalSCIMSDKUsersList, mockedInternalSCIMSDKEmptyUsersList)

		timeoutContext, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		users, err := FetchUsers(timeoutContext)

		assert.Empty(t, users)
		assert.Nil(t, err)
		assert.Nil(t, timeoutContext.Err())
	})

	t.Run("should return an error when the context timeout exceeds", func(t *testing.T) {
		defer monkey.UnpatchAll()
		monkey.Patch(internalSCIMSDKGroupsList, mockedInternalSCIMSDKEmptyGroupsList)
		monkey.Patch(internalSCIMSDKUsersList, mockedSCIMSDKUserListCTXTimeoutExceeded)

		timeoutContext, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		users, err := FetchUsers(timeoutContext)

		assert.Nil(t, users)
		assert.NotNil(t, err)
		assert.NotNil(t, timeoutContext.Err())
		assert.Contains(t, err.Error(), "deadline exceeded")
		assert.Contains(t, timeoutContext.Err().Error(), "deadline exceeded")
	})

	t.Run("should return an error when the context timeout exceeds in fetch groups", func(t *testing.T) {
		defer monkey.UnpatchAll()
		monkey.Patch(internalSCIMSDKGroupsList, mockedInternalSCIMSDKGroupListCTXTimeoutExceeded)
		monkey.Patch(internalSCIMSDKUsersList, mockedInternalSCIMSDKEmptyUsersList)

		timeoutContext, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		users, err := FetchUsers(timeoutContext)

		assert.Nil(t, users)
		assert.NotNil(t, err)
		assert.NotNil(t, timeoutContext.Err())
		assert.Contains(t, err.Error(), "deadline exceeded")
		assert.Contains(t, timeoutContext.Err().Error(), "deadline exceeded")
	})
}

func TestSDMSCIMCreateUser(t *testing.T) {
	t.Run("should create an user when passed a valid IdP user", func(t *testing.T) {
		defer monkey.UnpatchAll()
		monkey.Patch(internalSCIMSDKUsersCreate, mockedSCIMSDKUserCreate)

		response, err := CreateUser(context.Background(), &source.User{})

		assert.NotNil(t, response)
		assert.Nil(t, err)
	})

	t.Run("should create an user when using a context with timeout", func(t *testing.T) {
		defer monkey.UnpatchAll()
		monkey.Patch(internalSCIMSDKUsersCreate, mockedSCIMSDKUserCreate)

		timeoutContext, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		response, err := CreateUser(timeoutContext, &source.User{})

		assert.NotNil(t, response)
		assert.Nil(t, err)
		assert.Nil(t, timeoutContext.Err())
	})

	t.Run("should return an error when the context timeout exceeds", func(t *testing.T) {
		defer monkey.UnpatchAll()
		monkey.Patch(internalSCIMSDKUsersCreate, mockedSCIMSDKUserCreateCTXTimeoutExceeded)

		timeoutContext, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		response, err := CreateUser(timeoutContext, &source.User{})

		assert.Nil(t, response)
		assert.NotNil(t, err)
		assert.NotNil(t, timeoutContext.Err())
		assert.Contains(t, err.Error(), "deadline exceeded")
		assert.Contains(t, timeoutContext.Err().Error(), "deadline exceeded")
	})
}

func TestSDMSCIMDeleteUser(t *testing.T) {
	t.Run("should delete an user when passed a valid user ID", func(t *testing.T) {
		defer monkey.UnpatchAll()
		monkey.Patch(internalSCIMSDKUsersDelete, mockedSCIMSDKUserDelete)

		err := DeleteUser(context.Background(), *getMockUserRow())

		assert.Nil(t, err)
	})

	t.Run("should delete an user when passed a valid user ID and using a context with timeout", func(t *testing.T) {
		defer monkey.UnpatchAll()
		monkey.Patch(internalSCIMSDKUsersDelete, mockedSCIMSDKUserDelete)

		timeoutContext, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		err := DeleteUser(timeoutContext, *getMockUserRow())

		assert.Nil(t, err)
		assert.Nil(t, timeoutContext.Err())
	})

	t.Run("should return an error when the context timeout exceeds", func(t *testing.T) {
		defer monkey.UnpatchAll()
		monkey.Patch(internalSCIMSDKUsersDelete, mockedSCIMSDKUserDeleteCTXTimeoutExceeded)

		timeoutContext, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		err := DeleteUser(timeoutContext, *getMockUserRow())

		assert.NotNil(t, err)
		assert.NotNil(t, timeoutContext.Err())
		assert.Contains(t, err.Error(), "deadline exceeded")
		assert.Contains(t, timeoutContext.Err().Error(), "deadline exceeded")
	})
}

func TestSDMSCIMFetchGroups(t *testing.T) {
	t.Run("should return a empty list of groups when executing the default flow", func(t *testing.T) {
		defer monkey.UnpatchAll()
		monkey.Patch(internalSCIMSDKGroupsList, mockedInternalSCIMSDKEmptyGroupsList)

		groups, err := FetchGroups(context.Background())

		assert.Empty(t, groups)
		assert.Nil(t, err)
	})

	t.Run("should return a list of groups when executing the default flow", func(t *testing.T) {
		defer monkey.UnpatchAll()
		monkey.Patch(internalSCIMSDKGroupsList, mockedInternalSCIMSDKGroupsList)

		assertT := assert.New(t)

		groups, err := FetchGroups(context.Background())

		assert.NotNil(t, groups)
		assert.Nil(t, err)

		group := groups[0]

		assertT.NotEmpty(group.ID)
		assertT.NotEmpty(group.DisplayName)
		assertT.NotNil(group.Members)

		member := group.Members[0]

		assertT.NotEmpty(member.ID)
		assertT.NotEmpty(member.Email)
	})

	t.Run("should return a empty list of groups when using a context with timeout", func(t *testing.T) {
		defer monkey.UnpatchAll()
		monkey.Patch(internalSCIMSDKGroupsList, mockedInternalSCIMSDKEmptyGroupsList)

		timeoutContext, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		users, err := FetchGroups(timeoutContext)

		assert.Empty(t, users)
		assert.Nil(t, err)
		assert.Nil(t, timeoutContext.Err())
	})

	t.Run("should return an error when the context timeout exceeds", func(t *testing.T) {
		defer monkey.UnpatchAll()
		monkey.Patch(internalSCIMSDKGroupsList, mockedInternalSCIMSDKGroupListCTXTimeoutExceeded)

		timeoutContext, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		groups, err := FetchGroups(timeoutContext)

		assert.Nil(t, groups)
		assert.NotNil(t, err)
		assert.NotNil(t, timeoutContext.Err())
		assert.Contains(t, err.Error(), "deadline exceeded")
		assert.Contains(t, timeoutContext.Err().Error(), "deadline exceeded")
	})
}

func TestSDMSCIMCreateGroup(t *testing.T) {
	t.Run("should create a group when passed a valid IdP group", func(t *testing.T) {
		defer monkey.UnpatchAll()
		monkey.Patch(internalSCIMSDKGroupsCreate, mockedSCIMSDKGroupCreate)

		response, err := CreateGroup(context.Background(), &source.UserGroup{})

		assert.NotNil(t, response)
		assert.Nil(t, err)
	})

	t.Run("should create a group when using a context with timeout", func(t *testing.T) {
		defer monkey.UnpatchAll()
		monkey.Patch(internalSCIMSDKGroupsCreate, mockedSCIMSDKGroupCreate)

		timeoutContext, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		response, err := CreateGroup(timeoutContext, &source.UserGroup{})

		assert.Empty(t, response)
		assert.Nil(t, err)
		assert.Nil(t, timeoutContext.Err())
	})

	t.Run("should return an error when the context timeout exceeds", func(t *testing.T) {
		defer monkey.UnpatchAll()
		monkey.Patch(internalSCIMSDKGroupsCreate, mockedSCIMSDKGroupCreateCTXTimeoutExceeded)

		timeoutContext, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		response, err := CreateGroup(timeoutContext, &source.UserGroup{})

		assert.Nil(t, response)
		assert.NotNil(t, err)
		assert.NotNil(t, timeoutContext.Err())
		assert.Contains(t, err.Error(), "deadline exceeded")
		assert.Contains(t, timeoutContext.Err().Error(), "deadline exceeded")
	})
}

func TestSDMSCIMUpdateReplaceMembers(t *testing.T) {
	t.Run("should update members in a SDM Group when executing the default flow", func(t *testing.T) {
		defer monkey.UnpatchAll()
		monkey.Patch(internalSCIMSDKGroupsUpdateReplaceMembers, mockedSCIMSDKGroupUpdateReplaceMembers)

		err := ReplaceGroupMembers(context.Background(), &source.UserGroup{})

		assert.Nil(t, err)
	})

	t.Run("should update members in a SDM Group when using a context with timeout", func(t *testing.T) {
		defer monkey.UnpatchAll()
		monkey.Patch(internalSCIMSDKGroupsUpdateReplaceMembers, mockedSCIMSDKGroupUpdateReplaceMembers)

		timeoutContext, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		err := ReplaceGroupMembers(timeoutContext, &source.UserGroup{})

		assert.Nil(t, err)
		assert.Nil(t, timeoutContext.Err())
	})

	t.Run("should return an error when the context timeout exceeds", func(t *testing.T) {
		defer monkey.UnpatchAll()
		monkey.Patch(internalSCIMSDKGroupsUpdateReplaceMembers, mockedSCIMSDKGroupUpdateReplaceMembersCTXTimeoutExceeded)

		timeoutContext, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		err := ReplaceGroupMembers(timeoutContext, &source.UserGroup{})

		assert.NotNil(t, err)
		assert.NotNil(t, timeoutContext.Err())
		assert.Contains(t, err.Error(), "deadline exceeded")
		assert.Contains(t, timeoutContext.Err().Error(), "deadline exceeded")
	})
}

func TestSDMSCIMDeleteGroup(t *testing.T) {
	t.Run("should delete a group when passed a valid group ID", func(t *testing.T) {
		defer monkey.UnpatchAll()
		monkey.Patch(internalSCIMSDKGroupsDelete, mockedSCIMSDKGroupDelete)

		err := DeleteGroup(context.Background(), getMockGroupRow())

		assert.Nil(t, err)
	})

	t.Run("should delete a group when using a context with timeout", func(t *testing.T) {
		defer monkey.UnpatchAll()
		monkey.Patch(internalSCIMSDKGroupsDelete, mockedSCIMSDKGroupDelete)

		timeoutContext, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		err := DeleteGroup(timeoutContext, getMockGroupRow())

		assert.Nil(t, err)
		assert.Nil(t, timeoutContext.Err())
	})

	t.Run("should return an error when the context timeout exceeds", func(t *testing.T) {
		defer monkey.UnpatchAll()
		monkey.Patch(internalSCIMSDKGroupsDelete, mockedSCIMSDKGroupDeleteCTXTimeoutExceeded)

		timeoutContext, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		err := DeleteGroup(timeoutContext, getMockGroupRow())

		assert.NotNil(t, err)
		assert.NotNil(t, timeoutContext.Err())
		assert.Contains(t, err.Error(), "deadline exceeded")
		assert.Contains(t, timeoutContext.Err().Error(), "deadline exceeded")
	})
}

func mockedSCIMSDKUserListCTXTimeoutExceeded(_ context.Context, _ *scimsdk.PaginationOptions) scimsdk.UserIterator {
	time.Sleep(time.Millisecond * 2)
	return &MockUserIterator{err: errors.New("context deadline exceeded")}
}

func mockedSCIMSDKUserCreate(_ context.Context, _ scimsdk.CreateUser) (*scimsdk.User, error) {
	return &scimsdk.User{
		Name: &scimsdk.UserName{},
	}, nil
}

func mockedSCIMSDKUserCreateCTXTimeoutExceeded(_ context.Context, _ scimsdk.CreateUser) (*scimsdk.User, error) {
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

func mockedSCIMSDKGroupCreate(_ context.Context, _ scimsdk.CreateGroupBody) (*scimsdk.Group, error) {
	return &scimsdk.Group{}, nil
}

func mockedSCIMSDKGroupCreateCTXTimeoutExceeded(_ context.Context, _ scimsdk.CreateGroupBody) (*scimsdk.Group, error) {
	time.Sleep(time.Millisecond * 2)
	return nil, errors.New("context deadline exceeded")
}

func mockedSCIMSDKGroupUpdateReplaceMembers(_ context.Context, _ string, _ []scimsdk.GroupMember) (bool, error) {
	return true, nil
}

func mockedSCIMSDKGroupUpdateReplaceMembersCTXTimeoutExceeded(_ context.Context, _ string, _ []scimsdk.GroupMember) (bool, error) {
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

func mockedInternalSCIMSDKUsersList(_ context.Context, _ *scimsdk.PaginationOptions) scimsdk.UserIterator {
	return &MockUserIterator{
		buffer: []*scimsdk.User{
			getMockSCIMSDKUser(),
			getMockSCIMSDKUser(),
		},
		index: -1,
	}
}

func mockedInternalSCIMSDKEmptyUsersList(_ context.Context, _ *scimsdk.PaginationOptions) scimsdk.UserIterator {
	return &MockUserIterator{
		buffer: []*scimsdk.User{},
		index:  -1,
	}
}

func mockedInternalSCIMSDKGroupsList(_ context.Context, _ *scimsdk.PaginationOptions) scimsdk.GroupIterator {
	return &MockGroupIterator{
		buffer: []*scimsdk.Group{
			getMockSCIMSDKGroup(),
			getMockSCIMSDKGroup(),
		},
		index: -1,
	}
}

func mockedInternalSCIMSDKEmptyGroupsList(_ context.Context, _ *scimsdk.PaginationOptions) scimsdk.GroupIterator {
	return &MockGroupIterator{
		buffer: []*scimsdk.Group{},
		index:  -1,
	}
}

func mockedInternalSCIMSDKGroupListCTXTimeoutExceeded(_ context.Context, _ *scimsdk.PaginationOptions) scimsdk.GroupIterator {
	time.Sleep(time.Millisecond * 2)
	return &MockGroupIterator{err: errors.New("context deadline exceeded")}
}

func getMockSCIMSDKUser() *scimsdk.User {
	return &scimsdk.User{
		ID:          "www",
		Active:      true,
		DisplayName: "xxx",
		Emails: []scimsdk.UserEmail{
			{
				Primary: true,
				Value:   "yyy",
			},
		},
		Groups: []scimsdk.UserGroupReference{
			{
				Value: "yyy",
				Ref:   "xxx",
			},
		},
		Name: &scimsdk.UserName{
			FamilyName: "ccc",
			Formatted:  "ddd",
			GivenName:  "eee",
		},
		UserName: "yyy",
		UserType: "zzz",
	}
}

func getMockSCIMSDKGroup() *scimsdk.Group {
	return &scimsdk.Group{
		ID:          "xxx",
		DisplayName: "yyy",
		Members: []*scimsdk.GroupMember{
			{
				ID:    "www",
				Email: "yyy",
			},
		},
		Meta: &scimsdk.GroupMetadata{
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
		Members:     []*sink.GroupMember{},
	}
}
