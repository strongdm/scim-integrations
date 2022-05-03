#!/bin/bash

SUPPORTED_IDPS="google"

get_idp() {
  conf=""
  for s in $SUPPORTED_IDPS; do
    contains=$(echo $SDM_SCIM_IDP | grep -wq $s; echo $?)
    if [ $contains -eq 0 ]; then
      echo "$s"
      return
    fi
  done
  if [ "$conf" == "" ]; then
    echo "google"
  fi
}

cmd_flags=("-idp $(get_idp)")

if [ "$SDM_SCIM_IDP_QUERY" != "" ]; then
  cmd_flags+=("-query $SDM_SCIM_IDP_QUERY")
fi
if [ "$SDM_SCIM_ENABLE_RATE_LIMITER" == "true" ]; then
  cmd_flags+=("-enable-rate-limiter")
fi
if [ "$SDM_SCIM_DELETE_MISSING_USERS" == "true" ]; then
  cmd_flags+=("-delete-users-missing-in-idp")
fi
if [ "$SDM_SCIM_DELETE_MISSING_GROUPS" == "true" ]; then
  cmd_flags+=("-delete-groups-missing-in-idp")
fi
if [ "$SDM_SCIM_APPLY" == "true" ]; then
  cmd_flags+=("-apply")
fi

/scim $(echo ${cmd_flags[*]})
