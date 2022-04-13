# Configure Google IdP

To use Google Directory as an IdP, you must perform the following steps:

1. Go to [Google Cloud Console](http://console.cloud.google.com) and select your project

2. Go to "IAM & Admin" -> "Service Accounts"

3. Click on the button "Create Service Account" and create your Service Account

4. Click on the created Service Account

5. Then click on the "Keys" tab

6. Click in "Add Key" and select "Create new key"

7. Select "JSON" and create the key (then it will download your Service Account Key)

8. Go back to the "Details" tab and copy your `Unique ID`

9. Go to the [Google Admin Console](https://admin.google.com) and go to "Security" -> "Access and data Control" -> "API Controls"

10. Go to "MANAGE DOMAIN WIDE DELEGATION"

11. Add a new API Client

12. Then fill the form with the Service Account Unique ID that you copied and the scope "https://www.googleapis.com/auth/admin.directory.user"

And your Service Account is configured. You just need to pass the Service Account JSON file path to the `-key` parameter.

**A user can only be assigned to one OrgUnit at a time**

### Additional flags

To add a filter to the IdP search, you can use the `-query` flag refering to the [Google Users Search Documentation](https://developers.google.com/admin-sdk/directory/v1/guides/search-users)
