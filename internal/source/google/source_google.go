package google

import (
	"context"
	"errors"
	"os"
	"scim-integrations/internal/flags"
	"scim-integrations/internal/sink"
	"scim-integrations/internal/source"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
)

const fetchPageSize = 500

// DefaultGoogleCustomer refer to customer field in: https://developers.google.com/admin-sdk/directory/reference/rest/v1/users/list
const DefaultGoogleCustomer = "my_customer"

type SourceGoogle interface {
	FetchUsers(context.Context) ([]*source.User, error)
	ExtractGroupsFromUsers([]*source.User) []*source.UserGroup
	InternalGoogleFetchUsers(*admin.Service, string) (*admin.Users, error)
	GetGoogleAdminService(ctx context.Context) (*admin.Service, error)
	GetGoogleTokenSource(ctx context.Context) (oauth2.TokenSource, error)
}

type sourceGoogleImpl struct{}

func NewGoogleSource() SourceGoogle {
	return &sourceGoogleImpl{}
}

func (g *sourceGoogleImpl) FetchUsers(ctx context.Context) ([]*source.User, error) {
	return internalFetchUsers(ctx, g)
}

func internalFetchUsers(ctx context.Context, src SourceGoogle) ([]*source.User, error) {
	svc, err := src.GetGoogleAdminService(ctx)
	if err != nil {
		return nil, err
	}
	var nextPageToken string
	var users []*source.User
	for {
		response, err := src.InternalGoogleFetchUsers(svc, nextPageToken)
		if err != nil {
			return nil, err
		}
		users = append(users, googleUsersToSCIMUser(response.Users)...)
		if response.NextPageToken == "" {
			break
		}
		nextPageToken = response.NextPageToken
	}
	return users, nil
}

func (*sourceGoogleImpl) ExtractGroupsFromUsers(users []*source.User) []*source.UserGroup {
	var groups []*source.UserGroup
	mappedGroupMembers := map[string][]*sink.GroupMember{}
	for _, user := range users {
		for _, userGroup := range user.Groups {
			if _, ok := mappedGroupMembers[userGroup.Path]; !ok {
				mappedGroupMembers[userGroup.Path] = []*sink.GroupMember{
					{
						ID:    user.ID,
						Email: user.UserName,
					},
				}
			} else {
				mappedGroupMembers[userGroup.Path] = append(mappedGroupMembers[userGroup.Path], &sink.GroupMember{
					ID:    user.ID,
					Email: user.UserName,
				})
			}
		}
	}
	for orgUnitPath, members := range mappedGroupMembers {
		groups = append(groups, &source.UserGroup{DisplayName: formatGroupNameToSDM(orgUnitPath), Path: orgUnitPath, Members: members})
	}
	return groups
}

func (*sourceGoogleImpl) InternalGoogleFetchUsers(service *admin.Service, nextPageToken string) (*admin.Users, error) {
	return service.Users.List().Query(*flags.IdPQueryFlag).Customer(DefaultGoogleCustomer).PageToken(nextPageToken).MaxResults(fetchPageSize).Do()
}

func googleUsersToSCIMUser(googleUsers []*admin.User) []*source.User {
	var users []*source.User
	for _, googleUser := range googleUsers {
		users = append(
			users,
			&source.User{
				ID:         googleUser.Id,
				UserName:   googleUser.PrimaryEmail,
				GivenName:  googleUser.Name.GivenName,
				FamilyName: googleUser.Name.FamilyName,
				Active:     !googleUser.Suspended,
				Groups: []source.UserGroup{
					{
						DisplayName: formatGroupNameToSDM(googleUser.OrgUnitPath),
						Path:        googleUser.OrgUnitPath,
					},
				},
			},
		)
	}
	return users
}

func (g *sourceGoogleImpl) GetGoogleAdminService(ctx context.Context) (*admin.Service, error) {
	ts, err := g.GetGoogleTokenSource(ctx)
	if err != nil {
		return nil, err
	}
	svc, err := admin.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, err
	}
	return svc, nil
}

func (*sourceGoogleImpl) GetGoogleTokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	jsonCredentials, err := os.ReadFile(os.Getenv("SDM_SCIM_IDP_GOOGLE_KEY_PATH"))
	if err != nil {
		return nil, errors.New("Unable to read service account key file: " + err.Error())
	}
	config, err := google.JWTConfigFromJSON(jsonCredentials, admin.AdminDirectoryUserScope)
	if err != nil {
		return nil, err
	}
	config.Subject = os.Getenv("SDM_SCIM_IDP_GOOGLE_SUBJECT_USER")

	ts := config.TokenSource(ctx)
	return ts, nil
}

func formatGroupNameToSDM(orgUnitPath string) string {
	if orgUnitPath == "/" {
		alias := os.Getenv("SDM_SCIM_IDP_GOOGLE_ROOT_ORG_UNIT_ALIAS")
		if alias != "" {
			return alias
		}
		return "root"
	}
	orgUnits := strings.Split(orgUnitPath, "/")
	if len(orgUnits) == 0 {
		return ""
	}
	if len(orgUnits) == 1 {
		return orgUnits[0]
	}
	groupName := strings.Join(orgUnits[1:], "_")
	return groupName
}
