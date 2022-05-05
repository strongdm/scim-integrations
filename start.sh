#!/bin/bash

# setup cron
echo "$(export)" >> /etc/profile
echo "$SDM_SCIM_CRON root BASH_ENV=/etc/profile /bin/bash -l /exec.sh > /proc/1/fd/1 2>/proc/1/fd/2" > /etc/crontab

# start prometheus metrics endpoint
/scim expose-metrics &

# starts cron daemon
cron -f
