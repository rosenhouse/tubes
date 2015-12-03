# tubes

**work in progress + side project == be careful**

## Huh?
A CLI tool that simplifies the creation of CF development environments on AWS

## Goals
- Automate as much as possible
- Minimize required human configuration
- Ease common existing workflows, don't try to replace them
- Encourage disposability of environments

## What it does today

*Given* I've set my `AWS_*` environment variables on an empty AWS account

*When* I run 
 
 ```bash
 tubes -n my-environment up && bosh-init deploy environments/my-environment/director.yml
 ```

*Then* I get a [fully-operational](https://www.google.com/search?q=fully+operational&safe=active&source=lnms&tbm=isch) BOSH director on AWS

## What's next (maybe)
- generate a Concourse deployment manifest
- idempotent upsert, using data in state directory (see below)
- Fake SSH endpoint for integration tests
- Optional hosted zone: CNAMEs for everything
- Write a log of resources created
- Optional noob-feature: no ruby required; use the NAT box as a bastion; run `bosh-init deploy` and `bosh deploy` from an SSH session on the NAT box
- Related optional tin-foil-hat feature: No external IP for the BOSH director, all access via bastion.

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
