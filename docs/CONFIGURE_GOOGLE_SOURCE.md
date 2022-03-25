# Configure Google Source

To use Google Directory as a Source, you must perform the following steps:

1. Enable OAuth Consent: https://console.cloud.google.com/apis/credentials/consent (Internal is OK)
2. Create credentials for a Desktop App: https://console.cloud.google.com/apis/credentials
   1. Download your credentials file and save it in the project root folder
3. Enable Admin SDK API: https://console.cloud.google.com/apis/api/admin.googleapis.com/overview
4. Administrate Users and OrgUnits: https://admin.google.com/u/2/ac/users

**A user can only be assigned to one OrgUnit at a time**

### Optional Environment

- `SDM_SCIM_GOOGLE_ORGANIZATIONS_FILTER` - filter Google Directory users with a specific organization unit. e.g.: "engineering support"
