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
  -delete-unmatching-groups
    delete groups present in SDM but not in selected IdP data
  -delete-unmatching-users
    delete users present in SDM but not in the selected IdP data
  -google
    use Google as IdP
  -json
    dump a JSON report for debugging
  -plan
    do not apply changes just show initial queries
````

- Running Google IdP:
```
$ go run main.go -google
5 IdP users, 3 strongDM users in IdP, 3 strongDM roles in Idp
```

- For verbose reporting use `-plan` to your command.


## Contributing

Refer to the [contributing](CONTRIBUTING.md) guidelines or dump part of the information here.

## Support

Refer to the [support](SUPPORT.md) guidelines or dump part of the information here.
