#!/bin/bash

source /etc/environment
export SDM_SCIM_CRON=$SDM_SCIM_CRON

# setup cron
echo "$SDM_SCIM_CRON /bin/bash /exec.sh" > /etc/crontabs/root

crond -l 2 -f
