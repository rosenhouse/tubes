# tubes
something to do with the internet

# draft workflow

## starting config

### required
- aws region & iam user creds
- human-readable name of environment, unique within account

### optional
- name of hosted zone in the account

## stage 1: base infrastructure
- discover latest nat ami
- discover or generate ssh key
- allow additional config
  - network cidrs
  - nat instance size
  - nat instance ssh reachability
- generate cloudformation parameter set
- create/update cloudformation stack
  - hard-coded template
  - parameter set
- if generated ssh key, then store in s3

## stage 2: bosh director
- discover infrastructure IDs from cloudformation stack
- discover latest releases & stemcells from bosh.io
- discover or generate bosh director credentials
- allow additional config
  - bosh instance size
  - bosh instance reachability: default is 0.0.0.0/0
- generate bosh-init manifest
- bosh-init deploy from jumpbox
- store bosh-init manifest & state to S3

## stage 3: ???

## TODO
### Mutable environments
Variables: in-cloud vs on-filesystem

```
- Given the state directory is empty
- and there are no cloud resources
- When I run `up`
- Then I get a new stack and the state directory is updated

- Given the state directory is empty
- and there are no cloud resources
- When I run any other command
- Then I get an error

- Given the state directory is empty
- and there are some cloud resources
- When I run any command
- Then I get an error

- Given the state directory has content
- and there are no cloud resources
- When I run `up`
- Then the cloud resources get re-created and the state directory updated
- updated ips and ids are saved

- Given the state directory has content
- and there are no cloud resources
- When I run any other command
- Then I get an error

- Given the state directory has content
- And there are some cloud resources
- And there are no mismatches between them
- When I run any command
- Then it succeeds idempotently, updating both
```
