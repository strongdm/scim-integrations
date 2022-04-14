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
  -delete-groups-missing-in-idp
        delete groups present in SDM but not in the selected Identity Provider
  -delete-users-missing-in-idp
        delete users present in SDM but not in the selected Identity Provider
  -idp string
        use Google as an IdP
  -plan
        do not apply changes just show initial queries
  -query string
        pass a query according to the available query syntax for the selected Identity Provider
  -user string
        pass the user e-mail
  -verbose
        show the verbose report output
```

- Running Google IdP:

```
$ go run main.go -idp google -key ./key.json -user myuser@company.com

Collecting data...

Groups to create:

        + Display Name: /engineering

Users to create:

        + ID: 123456789123456789123
                + Display Name: Rodolfo Campos
                + Family Name: Campos
                + Given Name: Rodolfo
                + User Name: rodolfo+me3@strongdm.com
                + Active: true
                + Groups:
                        + /engineering

Groups to update:

         ~ ID: r-1a2b3c4d5e6f
         ~ Display Name: /engineering
         ~ Members:
                 ~ E-mail: rodolfo+me2@strongdm.com

Users to update:

        ~ ID: 123456789123456789123
                ~ Family Name: Campos
                ~ Given Name: Rodolfo
                ~ User Name: rodolfo+me2@strongdm.com
                ~ Active: true
                ~ SDMID: a-uuu

Groups to delete:

        - ID: r-www
                - Display Name: Removeme

Users to delete:

        - ID: a-xxx
                - Display Name: User 01
                - User Name: user+01@strongdm.com

        - ID: a-yyy
                - Display Name: User 02
                - User Name: user+02@strongdm.com

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
~ Group updated: engineering
         ~ Members:
                 ~ rodolfo+me2@strongdm.com
~ Group deleted: Removeme

Sync with google IdP finished
```

- For verbose reporting use `-plan` to your command.

- Currently, to not make an overhead of requests, we defined a limit of 1000 requests per 30s.

## Contributing

Refer to the [contributing](CONTRIBUTING.md) guidelines or dump part of the information here.

## Support

Refer to the [support](SUPPORT.md) guidelines or dump part of the information here.
