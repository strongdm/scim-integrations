package synchronizer

import (
	"scim-integrations/internal/flags"
	"scim-integrations/internal/sink"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReport(t *testing.T) {
	t.Run("should return true on HaveUsersSyncContent when have any user to add or edit or delete", func(t *testing.T) {
		report := newReport()
		report.IdPUsersToAdd = []*sink.UserRow{{}}
		assert.True(t, report.HaveUsersSyncContent())

		report.IdPUsersToAdd = []*sink.UserRow{}
		report.IdPUsersToUpdate = []*sink.UserRow{{}}
		assert.True(t, report.HaveUsersSyncContent())

		*flags.DeleteUsersNotInIdPFlag = true
		report.IdPUsersToUpdate = []*sink.UserRow{}
		report.SinkUsersNotInIdP = []*sink.UserRow{{}}
		assert.True(t, report.HaveUsersSyncContent())
	})

	t.Run("should return false on HaveUsersSyncContent when haven't any user to add or edit or delete", func(t *testing.T) {
		report := newReport()
		assert.False(t, report.HaveUsersSyncContent())
	})

	t.Run("should return true on HaveGroupsSyncContent when have any group to add or edit or delete", func(t *testing.T) {
		report := newReport()
		report.IdPUserGroupsToAdd = []*sink.GroupRow{{}}
		assert.True(t, report.HaveGroupsSyncContent())

		report.IdPUserGroupsToAdd = []*sink.GroupRow{}
		report.IdPUserGroupsToUpdate = []*sink.GroupRow{{}}
		assert.True(t, report.HaveGroupsSyncContent())

		*flags.DeleteGroupsNotInIdPFlag = true
		report.IdPUserGroupsToUpdate = []*sink.GroupRow{}
		report.SinkGroupsNotInIdP = []*sink.GroupRow{{}}
		assert.True(t, report.HaveGroupsSyncContent())
	})

	t.Run("should return false on HaveGroupsSyncContent when haven't any group to add or edit or delete", func(t *testing.T) {
		report := newReport()
		assert.False(t, report.HaveGroupsSyncContent())
	})
}
