# SCIM Integrations

SDM SCIM SDK with IdP Integrations

## Table of Contents

- [Installation](#installation)
- [Getting Started](#getting-started)
- [Contributing](#contributing)
- [Support](#support)

## Installation

You will need to download this script with the following command:

```
git clone https://github.com/strongdm/scim-integrations.git
```

## Getting Started

To use the SDM SCIM Integration dependencies you'll need the SCIM Token. You can get one in the SCIM settings section if you have an organization with overhaul permissions. If you're not, please, contact strongDM support.

Once you have the SCIM API Token, you can use exporting as an environment var:

```
$ export SDM_SCIM_TOKEN=<YOUR ADMIN TOKEN>
```

### Configuring Source:

- If you want to use Google Directory as a source refer to [CONFIGURE_GOOGLE_SOURCE.md](docs/CONFIGURE_GOOGLE_SOURCE.md)

### Samples

- Help:

```
$ go run main.go -help
  -add
        enable the visualization of the planned data for the create operation
  -all
        enable the visualization of the planned data for all operations (create, update and delete)
  -apply
        apply the planned changes
  -delete
        enable the visualization of the planned data for the delete operation
  -enable-rate-limiter
        synchronize the planned data with a requester rate limiter, limiting with a limit set as 1000 requests per 30 seconds
  -idp string
        use Google as an IdP
  -idp-query string
        define a query according to the available query syntax for the selected Identity Provider
  -update
        enable the visualization of the planned data for the update operation
```

- Running Google IdP:

```
$ go run main.go -idp google -all -apply

Collecting data...

Groups to create:

        + Display Name: /engineering

Users to create:

        + ID: 123456789123456789123
                + User Name: catherine@domain.com
                + Display Name: Catherine Cazares
                + Given Name: Catherine
                + Family Name: Cazares
                + Active: true
                + Groups:
                        + /engineering

Groups to update:

         ~ ID: r-1a2b3c4d5e6f
         ~ Display Name: /support
         ~ Members:
                 ~ E-mail: shannon@domain.com

Users to update:

        ~ ID: 123456789123456789123
                ~ User Name: maria@domain.com
                ~ Display Name: Maria New
                ~ Given Name: Maria
                ~ Family Name: New
                ~ Active: true
                ~ Groups:
                        ~ /support

Groups to delete:

        - ID: r-www
                - Display Name: Removeme

Users to delete:

        - ID: a-zzz
                - Display Name: Norman Jordan
                - User Name: norman@domain.com

Synchronizing users...
+ User created: catherine@domain.com
~ User updated: maria@domain.com
- User deleted: norman@domain.com

Synchronizing groups...
+ Group created: engineering
        + Members:
                + catherine@domain.com
~ Group updated: support
        ~ Members:
                ~ shannon@domain.com
~ Group deleted: Removeme
```

**NOTES**:

- If you just want to see the plan without applying, run the command above without the `-apply` flag.
- If you want to set a rate limiter when synchronizing the planned data, just add the `-enable-rate-limiter` flag. The limit was defined to 1000 requests per 30s.

## Running with Docker

When running with docker, you need to follow these steps:

- Create a file called `env-file` using the content of the `env-file.example` file and fill the following variables:
  - `SDM_SCIM_TOKEN` - SDM SCIM Token
  - `SDM_SCIM_IDP` - defines the IdP that you want to synchronize
- Then you can run `docker-compose up`

**NOTES**:

- the project was designed to handle orgs with max of 100,000 users and ~50 groups. If your use case is above this numbers, please reach out to support.
- if you want to run this application with `Prometheus` and `Grafana`, you can run `docker-compose` with the [docker-compose-prometheus.yml](./docker-compose-prometheus.yml) file to see an example running with a proper setup that you can follow. For more details, please refer to [CONFIGURE_PROMETHEUS.md](./docs/CONFIGURE_PROMETHEUS.md).

For more details, please refer to [CONFIGURE_DOCKER](./docs/CONFIGURE_DOCKER.md)

## Contributing

Refer to the [contributing](CONTRIBUTING.md) guidelines or dump part of the information here.

## Support

Refer to the [support](SUPPORT.md) guidelines or dump part of the information here.
