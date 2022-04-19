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
  -apply
        apply the planned changes
  -delete-groups-missing-in-idp
        delete groups present in SDM but not in the selected Identity Provider
  -delete-users-missing-in-idp
        delete users present in SDM but not in the selected Identity Provider
  -idp string
        use Google as an IdP
  -query string
        pass a query according to the available query syntax for the selected Identity Provider
  -rate-limiter
        synchronize the planned data with a requester rate limiter, limiting with a limit set as 1000 requests per 30 seconds
```

- Running Google IdP:

```
$ go run main.go -idp google -apply

Collecting data...

Groups to create:

        + Display Name: /engineering

Users to create:

        + ID: 123456789123456789123
                + User Name: rodolfo+me3@strongdm.com
                + Display Name: Rodolfo Campos
                + Given Name: Rodolfo
                + Family Name: Campos
                + Active: true
                + Groups:
                        + /engineering

Groups to update:

         ~ ID: r-1a2b3c4d5e6f
         ~ Display Name: /support
         ~ Members:
                 ~ E-mail: rodolfo+me2@strongdm.com

Users to update:

        ~ ID: 123456789123456789123
                ~ User Name: rodolfo+me2@strongdm.com
                ~ Display Name: Rodolfo Campos
                ~ Given Name: Rodolfo
                ~ Family Name: Campos
                ~ Active: true
                ~ Groups:
                        ~ /support

Groups to delete:

        - ID: r-www
                - Display Name: Removeme

Users to delete:

        - ID: a-zzz
                - Display Name: Rodolfo Campos
                - User Name: rodolfo+me@strongdm.com

Synchronizing users...
+ User created: rodolfo+me3@strongdm.com
~ User updated: rodolfo+me2@strongdm.com
- User deleted: rodolfo+me@strongdm.com

Synchronizing groups...
+ Group created: engineering
        + Members:
                + rodolfo+me3@strongdm.com
~ Group updated: support
        ~ Members:
                ~ rodolfo+me2@strongdm.com
~ Group deleted: Removeme

Sync with google IdP finished
```

- If you just want to see the plan without applying, run the command above without the `-apply` flag.

- If you want to set a rate limiter when synchronizing the planned data, just add the `-rate-limiter` flag. The limit was defined to 1000 requests per 30s.

## Contributing

Refer to the [contributing](CONTRIBUTING.md) guidelines or dump part of the information here.

## Support

Refer to the [support](SUPPORT.md) guidelines or dump part of the information here.
