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

1. Get the source and dependencies
 ```bash
 [ -d $GOPATH ] && cd $GOPATH
 mkdir -p src/github.com/rosenhouse && cd src/github.com/rosenhouse
 git clone git://github.com/rosenhouse/tubes
 cd tubes
 git submodule update --init --recursive
 ```

2. Run the offline test suite
 ```bash
 ./scripts/test-offline
 ```

3. Run the online test suite (optional)

 Requires AWS account region & credentials in your environment.  Takes a while, as it creates real resources on AWS.

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

## Things you can do manually
*things to automate eventually ...*

4. Manually `bosh-init` the director and get the director UUID
 ```bash
 bosh-init deploy environments/my-environment/director.yml
 eval `tubes -n my-environment show --bosh-environment`
 bosh -t $BOSH_TARGET status --uuid
 ```

5. Manually edit the partially-generated Concourse deployment manifest
 ```bash
 vim environments/my-environment/concourse.yml  # add the UUID at the top
 ```

5. Manually lookup the latest versions of Concourse & Garden Linux release, upload to the director, and deploy Concourse.
