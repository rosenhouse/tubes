# tubes
[![Build Status](https://api.travis-ci.org/rosenhouse/tubes.png?branch=master)](http://travis-ci.org/rosenhouse/tubes)

**work in progress + side project == be careful**

## Huh?
A CLI tool that simplifies the creation of CF development environments on AWS

## Goals
- Automate as much as possible
- Minimize required human configuration
- Ease common existing workflows, don't try to replace them
- Encourage disposability of environments

## Contributing
Pull requests are welcome.  Here's how to get started:

1. Get the source
 ```bash
 [ -d $GOPATH ] && cd $GOPATH
 mkdir -p src/github.com/rosenhouse && cd src/github.com/rosenhouse
 git clone git://github.com/rosenhouse/tubes
 cd tubes
 ```
 
2. Get dependencies
 ```bash
 ./scripts/get-dependencies
 ```

3. Build the binary (optional)
 ```bash
 go build
 ```
 
4. Run the offline test suite
 ```bash
 ./scripts/test-offline
 ```
 
5. Run the online test suite (optional)

 Requires AWS account region & credentials in your environment.
 
 ```bash
 ./scripts/test-full  # WARNING: this uses a REAL AWS account and will cost you real money.
 ```


## What it does today
Here's a brief walkthrough.  Run with `-h` flag to see all options.  There are still several manual steps.  Automating those is a high priority.

1. Install for easy access
 ```bash
 go install github.com/rosenhouse/tubes
 ```
 
2. Set your AWS environment variables
 ```bash
 AWS_DEFAULT_REGION=us-west-2
 AWS_ACCESS_KEY_ID=some-key
 AWS_SECRET_ACCESS_KEY=some-secret
 ```

3. Boot a new environment named `my-environment`
 ```bash
 tubes -n my-environment up
 
 ```
 This boots 2 CloudFormation stacks, a "base" stack to support a BOSH director, and a "Concourse" stack with dedicated subnet and Elastic LoadBalancer.  It generates deployment manifests in `$PWD/environments/my-environment`
 
4. Manually `bosh-init` the director
 ```bash
 bosh-init deploy environments/my-environment/director.yml
 export BOSH_USER=admin
 export BOSH_PASSWORD="$(tubes -n my-environment show --bosh-password)"
 export BOSH_TARGET="$(tubes -n my-environment show --bosh-ip)"
 bosh -t $BOSH_TARGET status --uuid
 ```

5. Manually edit the partially-generated Concourse deployment manifest
 ```bash
 vim environments/my-environment/concourse.yml  # add the UUID at the top
 ```
 
5. Manually lookup the latest versions of Concourse & Garden Linux release, upload to the director, and deploy Concourse.


## What's next (maybe)
- Automate more of the Concourse deployment workflow
- Refactor manifest generation code, there's lots of incidental complexity in there at the moment
- Refactor the integration tests for better readability
- Idempotent upsert, using data in state directory (see below)
- Optional hosted zone: DNS for everything
- Add SSL for Concourse, maybe with Let's Encrypt?
- Feature to rotate credentials?
- Deploy CF, somehow?
- Keep a log somewhere, for auditing?
- Generate a pipeline that idempotently deploys a CF on AWS
- Separate binaries for separate steps (package some as Concourse resources?)
  - CloudFormation resource supporting both `in` and `out`
  - Credential-generation
- For newbies: no ruby required, instead `ssh` to the NAT box and uses it as a bastion to run `bosh-init deploy` and `bosh deploy`
- For the paranoid: No external IP for the BOSH director, all access via bastion.

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
