#!/bin/bash

# export env vars without export definition
set -o allexport
source /etc/environment
set +o allexport

# setup cron
echo "$SDM_SCIM_CRON /bin/bash /exec.sh" > /etc/crontabs/root

# starts cron daemon
crond -f
