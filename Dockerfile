FROM golang:1.18-alpine

WORKDIR /scim

# set the default environment variables
ENV SDM_SCIM_IDP_QUERY=""
ENV SDM_SCIM_ENABLE_RATE_LIMITER="false"
ENV SDM_SCIM_ENABLE_APPLY="true"
ENV SDM_SCIM_ENABLE_DELETE_MISSING_USERS="false"
ENV SDM_SCIM_ENABLE_DELETE_MISSING_GROUPS="false"
ENV SDM_SCIM_IDP_KEY_PATH="/scim/keys/idp-key.json"
ENV SDM_SCIM_CRON="*/15 * * * *"

# build binary executable
COPY go.mod ./
COPY go.sum ./
COPY main.go ./
COPY internal ./internal
RUN /usr/local/go/bin/go build .
RUN mv scim-integrations scim
RUN rm -r main.go internal/ go.mod go.sum

# install dependencies
RUN apk add bash

# copy project files
COPY exec.sh /exec.sh
COPY start.sh /start.sh
COPY env-file /env-file
RUN echo "$(export)" > /etc/environment
RUN echo "$(cat /env-file)" >> /etc/environment

RUN chmod +x /start.sh /exec.sh /scim/scim

CMD ["sh", "/start.sh"]
