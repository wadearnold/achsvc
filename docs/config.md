## Configuration

PayGate uses a file based config for modifying the default way it operates. Various settings can be changed with this file which is written in [YAML format](https://en.wikipedia.org/wiki/YAML). Key names are case-insensitive.

Use the command-line flag `-config <filename>` for specifying where to read this file from.

Generic placeholders are defined as follows, but typically real-world examples are used. Brackets indicate that a parameter is optional. For non-list parameters the value is set to the specified default.

* `<address>`: a with scheme, host and port that parses as a URL
* `<base64-string>`: a Base64 encoded string
* `<boolean>`: a boolean that can take the values `true` or `false`
* `<duration>`: a duration matching the regular expression `[0-9]+(ms|[smhdwy])`
* `<filename>`: a valid path in the current working directory
* `<host>`: a valid string consisting of hostname or IP followed by an optional port number
* `<number>`: a 32-bit number, but typically only positive values
* `<scheme>`: a string that can take the values `http` or `https`
* `<string>`: a regular string
* `<secret>`: a regular string that is a secret, such as a password
* `<tmpl-string>`: a string which is template-expanded before usage

### Logging

```yaml
logging:
  # Which format to print logs as.
  # Options: plain or json
  [ format: <string> | default = "plain" ]
```

### HTTP

```yaml
http:
  # Address for paygate to bind its HTTP server on.
  [ bindAddress: <string> | default = ":8082" ]
```

### Admin

```yaml
admin:
  # Address for paygate to bind its admin HTTP server on.
  [ bindAddress: <strong> | default = ":9092" ]
  [ disableConfigEndpoint: <boolean> | default = false ]
```

### Customers

Right now we support the [Moov Customers](https://github.com/moov-io/customers) service for reading Customer and Account information. We are looking to support additional services.

```yaml
customers:
  # A DNS record responsible for routing us to a Moov Customers instance.
  endpoint: <address>
  accounts:
    decryptor:
      symmetric:
        # Base64 encoded URI for encryption key to use
        # Example: base64key://<base64-string>
        keyURI: <string>
  [ debug: <boolean> | default = false ]
```

### Organization

An organization is a value used to separate models from each other. This can be used to store Transfers for individual users or Transfers for different purposes.

```yaml
organization:
  # HTTP header name to lookup in each request
  [ header: <string> | default = "X-Organization" ]
  # Default value to be used for all requests. The header property will override
  # this value if it's found in a HTTP request.
  [ default: <string> ]
```

### Database

In production deployments we recommend deploying a replicated and secure MySQL cluster.

```yaml
database:
  # Setup a SQLite connection for the database. If using this config all fields are required.
  sqlite:
    [ path: <filename> ]
  # Setup a MySQL connection for the database. If using this config all fields are required.
  mysql:
    [ address: <address> ]
    [ username: <string> ]
    [ password: <secret> ]
    [ database: <string> ]
```

### ODFI

```yaml
# ODFI holds all the configuration for sending and retrieving ACH files with
# a financial institution to originate files.
odfi:
  # The ABA routing number to use and run PayGate under.
  # Example: 987654320
  routingNumber: <string>

  # Gateway holds FileHeader information which many ODFI require is always set
  # to certain values all files uploaded.
  # More Details: https://moov-io.github.io/paygate/file-details.html#file-header
  gateway:
    [ origin: <string> ]
    [ originName: <string> ]
    [ destination: <string> ]
    [ destinationName: <string> ]

  cutoffs:
    # An IANA Timezone used to determine when to upload ACH files to the ODFI.
    # Time fields in ACH files are created in this timezone.
    # Example: America/New_York
    timezone: <string>
    # Array of 24-hour and minute timestamps when to initiate cutoff processing.
    # Example: 16:15
    windows:
      - <string>

  # These paths point to directories on the remote FTP/SFTP server.
  inboundPath: <filename>
  outboundPath: <filename>
  returnPath: <filename>

  # Comma separated list of IP addresses and CIDR ranges where connections
  # are allowed. If this value is non-empty remote servers not within these
  # ranges will not be connected to.
  [ allowedIPs: <string> ]

  # Go template string of filenames for the remote server.
  [ outboundFilenameTemplate: <tmpl-string> ]

  # Configuration for using a remote File Transfer Protocol server
  # for ACH file uploads.
  ftp:
    hostname: <host>
    username: <string>
    [ password: <secret> ]
    [ caFile: <filename> ]
    [ dialTimeout: <duration> | default = 10s ]
    # Offer EPSV to be used if the FTP server supports it.
    [ disabledEPSV: <boolean> | default = false ]

  # Configuration for using a remote SSH File Transfer Protocol server
  # for ACH file uploads
  sftp:
    hostname: <host>
    username: <string>
    [ password: <secret> ]
    [ clientPrivateKey: <filename> ]
    [ hostPublicKey: <filename> ]
    [ dialTimeout: <duration> | default = 10s ]
    [ maxConnectionsPerFile: <number> | default = 1 ]
    # Sets the maximum size of the payload, measured in bytes.
    # Try lowering this on "failed to send packet header: EOF" errors.
    [ maxPacketSize: <number> | default = 20480 ]

  inbound:
    # How often PayGate should scan Inbound and Return directories for files to process.
    [ interval: <duration> ]

  fileConfig:
    batchHeader:
      # CompanyIdentification is a required field that is written to the Batch Header
      # field of the same name.
      companyIdentification: "MoovZZZZZZ"
    # Create an offsetting record for each debit and credit created from a Transfer.
    # Often FI's require this to understand and perform accounting operations.
    [ balanceEntries: <boolean> | default = false ]
    addendum:
      [ create05: <boolean> | default = false ]

  storage:
    # Should we delete the local temporary directory after inbound processing is finished.
    # Leaving these files around helps debugging, but also exposes customer information.
    # Empty directories are deleted and if no files are downloaded the entire temporary
    # directory is removed.
    [ cleanupLocalDirectory: <boolean> | default = false ]

    # Should we delete the remote file on an ODFI's server after downloading and processing of each file.
    [ keepRemoteFiles: <boolean> | default = false ]
    # Time to wait before deleting a zero byte file
    [ removeZeroByteFilesAfter: <time.Duration> ]

    local:
      [ directory: <filename> ]
```

### Transfers

```yaml
transfers:
  limits:
    # Fixed limits reject or force a manual review of Transfers if their amount exceeds
    # the configured hard or soft limit respectively.
    fixed:
      # Transfer whose amount exceeds this value are created with the REVIEWABLE status
      # for manual approval prior to upload.
      # Example: 250000
      softLimit: <number>

      # No Transfer amount is allowed to exceed this value when specified.
      # Example: 1000000
      [ hardLimit: <number> ]
```
### Pipeline

```yaml
pipeline:
  preUpload:
    gpg:
      [ keyFile: <filename> ]
      signer:
        [ keyFile: <filename> ]
        # Optional password to decrypt this private key.
        # It can also be set with PIPELINE_SIGNING_KEY_PASSWORD as an environment variable
        [ keyPassword: <secret> ]
  output:
    # Which encoding to use when writing ACH files to the remote.
    # Options: base64, encrypted-bytes, nacha
    [ format: <string> | default = "nacha" ]
  merging:
    [ directory: <filename> ]
    [ flattenBatches: <object> ]
  auditTrail:
    # BucketURI is a URI used to connect to a remote storage layer for saving
    # ACH files uploaded to the ODFI as part of records retention.
    # See the provider docs for more information: https://gocloud.dev/howto/blob/
    #
    # Example: gs://my-bucket
    bucketURI: <string>
    gpg:
      # Optional filepath used for encrypting ACH files when they're saved for auditing
      [ keyFile: <filename> ]
  stream:
    inmem:
      [ url: <address> ]
    kafka:
      brokers:
        - [ <address> ]
      group: [ <string> ]
      topic: [ <string> ]
  notifications:
    email:
      from: <string>
      to:
        - <string>
      # ConnectionURI is a URI used to connect with a remote SFTP server.
      # This config typically needs to contain enough values to successfully
      # authenticate with the server.
      #  - insecure_skip_verify is an optional parameter for disabling certificate verification
      #
      # Example: smtps://user:pass@localhost:1025/?insecure_skip_verify=true
      connectionURI: <string>
      [ template: <tmpl-string> ]
      companyName: <string>
    pagerduty:
      [ apiKey: <secret> ]
      [ from: <string> ]
      [ serviceKey: <string> ]
    slack:
      [ webhookURL: <secret> ]
```

### Validation

```yaml
# In order to validate Accounts and Customers to transfer money PayGate must ensure the accounts
# are valid, customers have access to them and are legally allowed in the US to transfer funds.
#
# Currently micro-deposits (two small credits and a debit of their sum) is the only allowed method
# of account validation.
validation:
  microDeposits:
    [ sameyDay: <boolean> ]
    source:
      # ID from the Customers service for the source of micro-deposit funds
      customerID: <string>
      accountID: <string>
      organization: <string>
    # Description is the default for what appears in the Online Banking
    # system for end-users of PayGate. Per NACHA limits this is restricted
    # to 10 characters.
    [ description: <string> ]
```

## Getting Help

 channel | info
 ------- | -------
 [Project Documentation](https://moov-io.github.io/paygate/) | Our project documentation available online.
 Twitter [@moov_io](https://twitter.com/moov_io)	| You can follow Moov.IO's Twitter feed to get updates on our project(s). You can also tweet us questions or just share blogs or stories.
 [GitHub Issue](https://github.com/moov-io) | If you are able to reproduce a problem please open a GitHub Issue under the specific project that caused the error.
 [moov-io slack](https://slack.moov.io/) | Join our slack channel (`#paygate`) to have an interactive discussion about the development of the project.
