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

if [ "$SDM_SCIM_ENABLE_RATE_LIMITER" == "true" ]; then
  cmd_flags+=("-enable-rate-limiter")
fi
if [ "$SDM_SCIM_APPLY" == "true" ]; then
  cmd_flags+=("-apply")
fi
if [ "$SDM_SCIM_ALL" == "true" ]; then
  cmd_flags+=("-all")
fi
if [ "$SDM_SCIM_ADD" == "true" ]; then
  cmd_flags+=("-add")
fi
if [ "$SDM_SCIM_UPDATE" == "true" ]; then
  cmd_flags+=("-update")
fi
if [ "$SDM_SCIM_DELETE" == "true" ]; then
  cmd_flags+=("-delete")
fi
if [ "$SDM_SCIM_IDP_QUERY" != "" ]; then
  cmd_flags+=("-idp-query '$SDM_SCIM_IDP_QUERY'")
fi

# We need this line to properly build the command and be able to run it
echo "scim $(echo ${cmd_flags[*]})" > /run_scim.sh
chmod +x /run_scim.sh
/run_scim.sh
