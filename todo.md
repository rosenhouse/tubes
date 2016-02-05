# todos

## small tasks
- generate a log

## medium tasks
- refactor manifest generation code, there's lots of incidental complexity in there at the moment
- automatically SSH to the NAT box and bosh-init the director from there
- discover the UUID of the director, and store that
- for the paranoid: No external IP for the BOSH director, all access via NAT box / bastion.
- automatically deploy concourse (via director API or SSH to NAT box)

## bigger tasks
- Automate more of the Concourse deployment workflow
- Optional hosted zone: DNS for everything
- Add SSL for Concourse, maybe with Let's Encrypt?
- Feature to rotate credentials?
- Deploy CF, somehow?
- Generate a pipeline that idempotently deploys a CF on AWS
- Separate binaries for separate steps (package some as Concourse resources?)
  - CloudFormation resource supporting both `in` and `out`
  - Credential-generation
- Idempotent upsert, using data in state directory (see below)

### Idempotency user stories

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
