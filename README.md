# awstk

Command line toolkit for AWS. Provides utilities for EC2, RDS, Aurora and more.

## Listing commands

```
awstk ec2 ls [-S stack-name]
```
Lists EC2 instances in the current region or within the given CloudFormation stack.

```
awstk rds ls [-S stack-name]
```
Lists RDS instances in the current region or from the specified stack.

```
awstk aurora ls [-S stack-name]
```
Lists Aurora DB clusters with optional stack filtering.
