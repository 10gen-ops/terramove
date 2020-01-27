# terramove

This application helps in managing the migration of terraform resources from one state file to a new one. It does that by using a config file that has a mapping of "source" (from) and "destination" (to) resources to migrate. The original state file will be read from a file and terramove will extract the ids from the source resources to use them on the generated terraform import commands.

Terramove will produce `terraform import` commands in stdout that can later be applied one by one when in the directory of the new module.

## workflow

- cd into original module and pull the state file using `terraform state pull > /tmp/original.tfstate`
- list the resources with `terraform state list -state /tmp/original.tfstate` and identify what resources were moved to the new module. 
- Write the [config file](#config) using the original resources and the new ones.
- Run terramove using the original state file and the config: `terramove --config <path_to_config> --state-file /tmp/original.tfstate`. Copy the terraform import commands that are going to be moved to the new module.
- cd to the new module directory and run the required `terraform import` commands.

## config

Default config file location is: `~/.terramove.yaml`

``` yaml
migrations:
  - from: "aws_s3_bucket.test"
    to: "aws_s3_bucket.testing"
  - from: "aws_route53_zone.primary"
...
```

If one of the migrations don't have the `to:` key, it's understood that the resource definition didn't change and the `from:` key will be used as the new resource name too.

## output

``` sh
terraform import <resource1> <original_id1>
terraform import <resource2> <original_id2>
terraform import <resource3> <original_id3>
```
