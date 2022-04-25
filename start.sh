#!/bin/bash

# setup cron
echo "$SDM_SCIM_CRON /bin/bash /exec.sh" > /etc/crontabs/root

# starts cron daemon
crond -f
