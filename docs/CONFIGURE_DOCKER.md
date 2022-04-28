# CONFIGURE DOCKER

While configuring docker, you can set the following optional variables:

- `SDM_SCIM_IDP_QUERY` - defines the IdP search query
- `SDM_SCIM_ENABLE_RATE_LIMITER` - enables the rate limiter when executing the synchronize process. It's disabled by default
- `SDM_SCIM_APPLY` - enables the synchronize process about the planned output. It's enabled by default
- `SDM_SCIM_DELETE_MISSING_USERS` - enables the delete missing users behavior when executing the synchronize process. It's disabled by default
- `SDM_SCIM_DELETE_MISSING_GROUPS` - enables the delete missing groups behavior when executing the synchronize process. It's disabled by default
- `SDM_SCIM_CRON` - cron run time, enclosed in double quotes. It's running every fifteen minutes by default ("\*/15 \* \* \* \*")
