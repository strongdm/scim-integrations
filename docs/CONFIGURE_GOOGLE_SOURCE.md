# Configure Google IdP

To use Google Directory as an IdP, you must perform the following steps:

1. Go to [Google Cloud Console](http://console.cloud.google.com) and select your project

![img1](https://user-images.githubusercontent.com/49597325/163180915-38358261-1063-4b76-b8f3-eb63ed83f44d.png)

2. Go to "IAM & Admin" -> "Service Accounts"

![img2](https://user-images.githubusercontent.com/49597325/163180947-addbec27-c6b9-4575-87a6-084090f89d08.png)

3. Click on the button "Create Service Account" and create your Service Account

![img3](https://user-images.githubusercontent.com/49597325/163180978-73f41679-e0cb-4947-998b-7aed9b60c27c.png)

4. Click on the created Service Account

![img4](https://user-images.githubusercontent.com/49597325/163180998-6f697700-1023-4365-ae6f-ae73bd57d4d6.png)

5. Then click on the "Keys" tab

![img5](https://user-images.githubusercontent.com/49597325/163181027-2c491377-8496-4177-85bf-fe6959ceab9c.png)

6. Click on "Add Key" and select "Create new key"

![img6](https://user-images.githubusercontent.com/49597325/163181039-2ccdb98e-95c5-49a4-ab6d-ae318e945367.png)

7. Select "JSON" and create the key (then it will download your Service Account Key - you'll use this key to authenticate in the application)

![img7](https://user-images.githubusercontent.com/49597325/163181052-ba2c55ac-003a-407c-ace6-8db21435ab5b.png)

8. Go back to the "Details" tab and copy your "Unique ID"

![img8](https://user-images.githubusercontent.com/49597325/163181064-f01be75d-7a3f-48f0-85c5-df736e46254a.png)

9. Go to the [Google Admin Console](https://admin.google.com) and go to "Security" -> "Access and data Control" -> "API Controls"

![img9](https://user-images.githubusercontent.com/49597325/163181081-05b1833f-13af-4f67-8bca-6b6df6dafcdc.png)

10. Click on "MANAGE DOMAIN WIDE DELEGATION"

![img10](https://user-images.githubusercontent.com/49597325/163181095-93e1a944-cddc-4600-a30e-46453af9cdab.png)

11. Click on "Add new" to add a new API Client

![img11](https://user-images.githubusercontent.com/49597325/163181113-26017685-1017-40bf-b968-42e612f42c0a.png)

12. Then fill the form with the Service Account Unique ID that you copied and the OAuth Scope "https://www.googleapis.com/auth/admin.directory.user"

![img12](https://user-images.githubusercontent.com/49597325/163181123-22e50c19-7a3b-432f-873c-c2c7372920be.png)

And your Service Account is configured. You just to set the path of your downloaded Service Account JSON file path into the `SDM_SCIM_IDP_KEY` env var

**A user can only be assigned to one OrgUnit at a time**

### Additional flags

To add a filter to the IdP search, you can use the `-query` flag refering to the [Google Users Search Documentation](https://developers.google.com/admin-sdk/directory/v1/guides/search-users)
