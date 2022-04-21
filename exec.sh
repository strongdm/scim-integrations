#!/bin/bash

source /etc/environment
export SDM_SCIM_TOKEN=$SDM_SCIM_TOKEN
export SDM_SCIM_IDP_USER=$SDM_SCIM_IDP_USER

cmd_flags=("-idp ${SDM_SCIM_IDP}")

if [ "${SDM_SCIM_IDP_QUERY}" != "" ]; then
  cmd_flags+=("-query ${SDM_SCIM_IDP_QUERY}")
fi
if [ "${SDM_SCIM_ENABLE_RATE_LIMITER}" == "true" ]; then
  cmd_flags+=("-enable-rate-limiter")
fi
if [ "${SDM_SCIM_ENABLE_DELETE_MISSING_USERS}" == "true" ]; then
  cmd_flags+=("-delete-users-missing-in-idp")
fi
if [ "${SDM_SCIM_ENABLE_DELETE_MISSING_GROUPS}" == "true" ]; then
  cmd_flags+=("-delete-groups-missing-in-idp")
fi
if [ "${SDM_SCIM_ENABLE_APPLY}" == "true" ]; then
  cmd_flags+=("-apply")
fi

cd /scim
./scim $(echo ${cmd_flags[*]})
