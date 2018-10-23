# UpCloud Go API client library

[![Build Status](https://travis-ci.org/UpCloudLtd/upcloud-go-api.svg?branch=master)](https://travis-ci.org/UpCloudLtd/upcloud-go-api)
[![Go Report Card](https://goreportcard.com/badge/github.com/UpCloudLtd/upcloud-go-api)](https://goreportcard.com/report/github.com/UpCloudLtd/upcloud-go-api)
[![GoDoc](https://godoc.org/github.com/UpCloudLtd/upcloud-go-api?status.svg)](https://godoc.org/github.com/UpCloudLtd/upcloud-go-api)

This is the official client for interfacing with UpCloud's API using the Go programming language. The features allows for easy and quick development and integration when using Go.

## Installation and requirements

You'll need Go 1.7 or higher to use the client. You can use the following command to retrieve the client:

```
go get github.com/UpCloudLtd/upcloud-go-api
```

## Usage

The general usage of the client adheres to the following pattern:

* Authenticate by creating a `client.Client`
* Create a `service.Service` by passing the newly created `client` object to it
* Interface with the API using the various methods of the `service` object. Methods that take parameters wrap them in request objects.

We recommend setting up a separate subaccount for API usage to allow better access control and security. You can find out more about creating subaccounts at the following support article for [Server Tags and Group Accounts](https://www.upcloud.com/support/server-tags-and-group-accounts/). We strongly recommend limiting the connections to a specific address or address range for security purposes.

The examples here only deal with how to use the client itself. For information on how to use the API in general, please consult the [UpCloud API documentation](https://www.upcloud.com/api/).

### Creating the client and the service

```go
// Authenticate by passing your account login credentials to the client
c := client.New(user, password)

// It is generally a good idea to override the default timeout of the underlying HTTP client since some requests block for longer periods of time
c.SetTimeout(time.Second * 30)

// Create the service object
svc := service.New(c)
```

### Validating credentials

The easiest way to check whether the client credentials are correct is to issue a call to `GetAccount()`.

```go
username := "completely"
password := "invalid"

svc := service.New(client.New(username, password))

_, err := svc.GetAccount()

if err != nil {
	panic("Invalid credentials")
}
```

### Error handling

All `Service` methods return a result and an error object. You can differentiate between generic connection errors (like the API not being reachable) and service errors, which are errors returned in the response body by the API. This is useful for gracefully recovering from certain types of errors.

```go
username := "completely"
password := "invalid"

svc := service.New(client.New(username, password))

_, err := svc.GetAccount()

// Handle errors in general
if (err != nil) {
	// Handle service errors specifically
	if serviceError, ok := err.(*upcloud.Error); ok {
		fmt.Println(serviceError.ErrorCode)
		fmt.Println(serviceError.ErrorMessage)
	}
}
````

This snippet would print the following:

```
AUTHENTICATION_FAILED
Authentication failed using the given username and password.
```

The rest of these examples assume you already have a service object configured and named `svc`.

### Retrieving a list of servers

The following example will retrieve a list of servers the account has access to.

```go
// Retrieve the list of servers
servers, err := svc.GetServers()

if err != nil {
	panic(err)
}

// Print the UUID and hostname of each server
for _, server := range servers.Servers {
	fmt.Println(fmt.Sprintf("UUID: %s, hostname: %s", server.UUID, server.Hostname))
}
```

### Creating a new server

Since the request for creating a new server is asynchronous, the server will report its status as "maintenance" until the deployment has been fully completed.

```go
// Create the server
serverDetails, err := svc.CreateServer(&request.CreateServerRequest{
	Zone:             "fi-hel1",
	Title:            "My new server",
	Hostname:         "server.example.com",
	PasswordDelivery: request.PasswordDeliveryNone,
	StorageDevices: []request.CreateServerStorageDevice{
		{
			Action:  request.CreateStorageDeviceActionClone,
			Storage: "01000000-0000-4000-8000-000030060200",
			Title:   "disk1",
			Size:    30,
			Tier:    upcloud.StorageTierMaxIOPS,
		},
	},
	IPAddresses: []request.CreateServerIPAddress{
		{
			Access: upcloud.IPAddressAccessPrivate,
			Family: upcloud.IPAddressFamilyIPv4,
		},
		{
			Access: upcloud.IPAddressAccessPublic,
			Family: upcloud.IPAddressFamilyIPv4,
		},
		{
			Access: upcloud.IPAddressAccessPublic,
			Family: upcloud.IPAddressFamilyIPv6,
		},
	},
})

if err != nil {
	panic(err)
}

fmt.Println(fmt.Sprintf("Server %s with UUID %s created", serverDetails.Title, serverDetails.UUID))

// Block for up to five minutes until the server has entered the "started" state
err = svc.WaitForServerState(&request.WaitForServerStateRequest{
	UUID:         serverDetails.UUID,
	DesiredState: upcloud.ServerStateStarted,
	Timeout:      time.Minute * 5,
})

if err != nil {
	panic(err)
}

fmt.Println("Server is now started")
```

### Templatizing a server's storage device

In this example, we assume that there is a server represented by the variable `serverDetails` and that the server state is `stopped`. The next piece of code allows you to templatize the first storage device of the server.

```go
// Loop through the storage devices
for i, storage := range serverDetails.StorageDevices {
	// Find the first device
	if i == 0 {
		// Templatize the storage
		storageDetails, err := svc.TemplatizeStorage(&request.TemplatizeStorageRequest{
			UUID:  storage.UUID,
			Title: "Templatized storage",
		})

		if err != nil {
			panic(err)
		}
		
		fmt.Println(fmt.Sprintf("Storage templatized as %s", storageDetails.UUID))		
		break
	}
}
```

### Create a manual backup

In this example, we assume that there is a storage device represented by `storageDetails` and that if it is attached to any server, the server is stopped.

```go
backupDetails, err := svc.CreateBackup(&request.CreateBackupRequest{
	UUID:  storageDetails.UUID,
	Title: "Backup",
})

if err != nil {
    panic(err)
}

fmt.Println(fmt.Sprintf("Backup of %s created as %s", storageDetails.UUID, backupDetails.UUID))
```

### Create a new firewall rule

In this example, we assume that there is a server represented by the variable `serverDetails`.

```go
firewallRule, err := svc.CreateFirewallRule(&request.CreateFirewallRuleRequest{
	ServerUUID: serverDetails.UUID,
	FirewallRule: upcloud.FirewallRule{
		Direction: upcloud.FirewallRuleDirectionIn,
		Action:    upcloud.FirewallRuleActionAccept,
		Family:    upcloud.IPAddressFamilyIPv4,
		Protocol:  upcloud.FirewallRuleProtocolTCP,
		Position:  1,
		Comment:   "Accept all TCP input on IPv4",
	},
})

if err != nil {
    panic(err)
}
```

For more examples, please consult the service integration test suite (`upcloud/service/service_test.go`).

### Debugging the API using Postman

The repository contains a Postman collection which can be used to quickly perform requests against the API to see what 
it returns. Import the collection into Postman, then create an environment containing the following variables:

* `authorization` - the value of the `Authorization` HTTP header, e.g. `Basic <base64>`
* `serveruuid` - an existing server UUID
* `ipaddress` - an existing IP address that is assigned to one of your servers
* `storageuuid` - the UUID of a piece of storage (can be a template, doesn't have to be a disk)

## Testing

To be able to run the test suite you'll need to export the following environment variables with their corresponding 
values:

* `UPCLOUD_GO_API_TEST_USER` (the API username)
* `UPCLOUD_GO_API_TEST_PASSWORD` (the API password)
* `UPCLOUD_GO_API_TEST_DELETE_RESOURCES` (either `yes` or `no`)

To run the test suite, run `go test ./... -v -parallel 8`. If `UPCLOUD_GO_API_TEST_DELETE_RESOURCES` is set to `yes`, 
all resources will be stopped and/or deleted after the test suite has run. Be careful which account you use for 
testing so you don't accidentally delete or your production resources!

You can skip running the integration tests and just run the unit tests by passing `-short` to the test command.

## License

This client is distributed under the [MIT License](https://opensource.org/licenses/MIT), see LICENSE.txt for more information.
