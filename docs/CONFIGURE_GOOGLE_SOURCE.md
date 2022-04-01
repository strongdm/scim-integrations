# Configure Google Source

To use Google Directory as a Source, you must perform the following steps:

1. Enable OAuth Consent: https://console.cloud.google.com/apis/credentials/consent (Internal is OK)
2. Create credentials for a Desktop App: https://console.cloud.google.com/apis/credentials
   1. Download your credentials file and save it in the project root folder
3. Enable Admin SDK API: https://console.cloud.google.com/apis/api/admin.googleapis.com/overview
4. Administrate Users and OrgUnits: https://admin.google.com/u/2/ac/users
5. Execute the script [auth.go](../tools/google/auth.go) to generate the `token.json` file. This file will be used to authenticate in Google Admin SDK

**A user can only be assigned to one OrgUnit at a time**

### Additional flags

To add a filter to the IdP search, you can use the `-query` flag refering to the [Google Users Search Documentation](https://developers.google.com/admin-sdk/directory/v1/guides/search-users)
