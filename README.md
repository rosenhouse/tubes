# tubes
something to do with the internet


## TODO
Need to have:
- generate unique IAM credentials for use by BOSH director -- don't expose user's IAM creds
- bosh-init from jump box

Want to have:
- idempotent upsert from state directory
- BOSH director manifest doesn't use external IP at all (NATs messages stay inside VPC)
- Fake AWS for integration tests
- Fake SSH endpoint for integration tests
- Create CNAME for bosh director

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
