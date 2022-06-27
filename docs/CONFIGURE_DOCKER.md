# CONFIGURE DOCKER

While configuring docker, you can set the following optional variables:

- `SDM_SCIM_IDP_QUERY` - define the IdP search query
- `SDM_SCIM_ENABLE_RATE_LIMITER` - enable the rate limiter when executing the synchronize process. It's disabled by default
- `SDM_SCIM_APPLY` - enable the synchronize process about the planned output. It's enabled by default
- `SDM_SCIM_CRON` - cron run time, enclosed in double quotes. It's running every fifteen minutes by default ("\*/15 \* \* \* \*")
- `SDM_SCIM_ALL` - enable the plan and the sync process for all operations. It's disabled by default
- `SDM_SCIM_ADD` - enable the plan and the sync process for the create operation for users and groups. It's disabled by default
- `SDM_SCIM_UPDATE` - enable the plan and the sync process for the update operation for users and groups. It's disabled by default
- `SDM_SCIM_DELETE` - enable the plan and the sync process for the delete operation for users and groups. It's disabled by default
- `SDM_SCIM_REPORTS_DATABASE_PATH` - define the absolute path of the sqlite3 reports database. It's disabled by default

**NOTE**: when using `SDM_SCIM_REPORTS_DATABASE_PATH` and running the application outside docker you must set the environment variable `CGO_ENABLED=1`
