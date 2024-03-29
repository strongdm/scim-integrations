# build stage
FROM golang:1.18 as BUILDER

WORKDIR /scim

# build binary executable
COPY go.mod ./
COPY go.sum ./
COPY main.go ./
COPY internal ./internal
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=1 /usr/local/go/bin/go build -o scim .
RUN rm -r main.go internal/ go.mod go.sum

# final stage
FROM debian:buster-slim AS RUNNER

WORKDIR /scim-integrations

# get build stage generated binary
COPY --from=BUILDER /scim/scim /scim-integrations/scim

# set the default environment variables
ENV PATH="/scim-integrations/:${PATH}"
ENV SDM_SCIM_IDP_QUERY=""
ENV SDM_SCIM_ENABLE_RATE_LIMITER="false"
ENV SDM_SCIM_APPLY="true"
ENV SDM_SCIM_ADD="false"
ENV SDM_SCIM_UPDATE="false"
ENV SDM_SCIM_DELETE="false"
ENV SDM_SCIM_ALL="true"
ENV SDM_SCIM_IDP_GOOGLE_KEY_PATH="/scim-integrations/keys/idp-key.json"
ENV SDM_SCIM_REPORTS_DATABASE_PATH="/reports.db"
ENV SDM_SCIM_CRON="*/15 * * * *"
ENV CGO_ENABLED="1"

# install dependencies
RUN apt-get update -y
RUN apt-get install -y cron ca-certificates

# copy project files
COPY exec.sh /exec.sh
COPY start.sh /start.sh

# add execution
RUN chmod +x /start.sh /exec.sh /scim-integrations/scim

CMD ["bash", "/start.sh"]
