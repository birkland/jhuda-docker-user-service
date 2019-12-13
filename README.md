# JHUDA user service

[![Build Status](https://travis-ci.com/jhu-sheridan-libraries/jhuda-docker-user-service.svg?branch=master)](https://travis-ci.com/jhu-sheridan-libraries/jhuda-docker-user-service)

Contains the JHUDA user service, which provides an HTTP API for finding information about the
current shibboleth logged-in user.  Presumes the incoming requests are protected by shibboleth, and
shibboleth releases attributes about the current user.

## Usage

    jhuda-docker-user-service serve

## Configuration

For cli flags, see `jhuda-docker-user-service help`

Environment variables are as follows:

* `USER_SERVICE_PORT` - Port to serve the user service on (default `8091`)
* `USER_SERVICE_JSONLD_CONTEXT` - JSONLD-context for User JSON responses (optional)
* `USER_SERVICE_USER_BASEURL` - BaseURL for user IDs (optional, e.g. `http://archive.local/fcrepo/rest/users`)

Shibboleth headers can be controlled by headers as well, if the defaults don't work out

* `SHIB_HEADER_EPPN` Name of the Eppn header (Default `Eppn`)
* `SHIB_HEADER_DISPLAYNAME`: Name of the Displayname header (Default `Displayname`)
* `SHIB_HEADER_EMAIL`: Name of the e-mail header (default `Mail`)
* `SHIB_HEADER_GIVEN_NAME`: Name of the "given name" header (default `Givenname`)
* `SHIB_HEADER_LAST_NAME`: Name of the "last name" header (default: `Sn`)
* `SHIB_HEADER_LOCATORS`: Comma-separated list of all headers to use as locators (default `Employeenumber,unique-id,Eppn`).  
  * Values may be colon separated `headerName:locatorName` for using a different values for header and locatorId.  For example, `Ajp_unique-id:unique-id` could result in extracting the value of header `Ajp_unique-id: foo` to formulate the locator `johnshopkins.edu:unique-id:foo`
