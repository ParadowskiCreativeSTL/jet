<p align="center"><img src="img/logo.png" alt="Jet Logo" /></p>

# jet

A super fast deployment tool for read-only, production WordPress sites.

## Why?

WordPress is a wonderful CMS, but it suffers from a few disadvantages. When you have multiple authors working on the same site, sometimes mistakes will be made and errors will be had. You don't want these to go public! A better solution is to have a staging server setup where you make all content edits and then deploy all of those edits to production at once when you have reviewed changes.

NOTE: This repository contains ONLY the content deployment tool. Your mileage may vary on when a production deploy is appropriate, so it's up to you to build a WordPress hook. This is just a simple executable.

## Compiling

As of right now, there is no automated install process, you'll have to do this manually (PRs welcome!). You'll need to have the latest version of [Go](https://golang.org/dl/) installed. You can then start by cloning this repository into your GOPATH.

Since jet needs to be compiled into an executable, you'll need to find out [what type of kernel you're using](https://unix.stackexchange.com/questions/88644/how-to-check-os-and-version-using-a-linux-command). Usually, this can be achieved by running:
```
$ uname -a
```

Once you've determined the type of kernel jet will be run on (the host machine, not your local environment), you'll need to compile for that architecture by running:
```
$ env GOOS=<TARGET_OS> GOARCH=<TARGET_ARCHITECTURE> go build ./
```
Read more: [Digital Ocean](https://www.digitalocean.com/community/tutorials/how-to-build-go-executables-for-multiple-platforms-on-ubuntu-16-04#step-4-%E2%80%94-building-executables-for-different-architectures)

Once you've got a nice, shiny, new executable, you can throw it up on your host machine in `/usr/local/bin/jet` to use it globally.

## Environment Setup

The jet content deployment tool uses AWS S3 buckets for WordPress uploads storage. So to set up your environment for use, start by creating a new bucket in S3. You'll want to give it public permissions since everybody should be able to submit `GET` requests for your assets, the rest of the settings are up to your preference.

This tool also assumes that you have the WP-CLI installed. Instructions on how that can be done can be found [here](https://wp-cli.org/).

## Configuration

Before using this tool, you'll need to ensure that you have AWS credentials set up in the user's home folder. TLDR: create a file named `credentials` in `~/.aws` and insert your proper IAM credentials. If those terms or this process don't make sense to you, the AWS documentation will be much better at explaining: [AWS CLI Configuration and Credential Files](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html)

In the root directory of each project should be the following files:
```
/config.json
/mysql.cnf
```

Sample `config.json`:
```
{
    "binary_paths": {
        "ssh": "/usr/bin/ssh",
        "mysql_admin": "/usr/bin/mysqladmin",
        "mysql_dump": "/usr/bin/mysqldump",
        "mysql": "/usr/bin/mysql",
        "scp": "/usr/bin/scp",
        "php": "/usr/bin/php",
        "wp": "/usr/local/bin/wp"
    },
    "s3": {
            "url": "s3://a-bucket-name",
            "region": "us-east-2",
            "bucket_prefix": "htdocs/wordpress/uploads"
    },
    "environments": {
        "production": {
            "user": "admin",
            "host": "this.is.a.server.com",
            "root_directory": "/var/www/example.com",
            "uploads_location": "htdocs/wordpress/uploads",
            "database": {
                "name": "example_com",
                "host": "localhost",
                "port": "3306",
                "username": "username",
                "password": "password",
                "table_prefix": "wp_"
            },
            "target_url_patterns": [
                    "staging\.website\.url",
                    "qa\.website\.url"
            ],
            "replacement_url": "example.com"
        },
        "staging": {
            "user": "admin",
            "host": "this.is.a.staging.server.com",
            "root_directory": "/var/www/example.com",
            "uploads_location": "htdocs/wordpress/uploads",
            "database": {
                "name": "example_com",
                "host": "localhost",
                "port": "3306",
                "username": "username",
                "password": "password",
                "table_prefix": "wp_"
            }
        }
    }
}
```

Sample `mysql.cnf`:
```
[client]
user=test
password=test
host=127.0.0.1
port=3306
```

## Running

To run the tool, you'll need to call:
```
$ jet --environment=staging
```
and the tool should take care of the rest! It will prepare the staging backup, and automatically call `$ jet --environment=production <BACKUP_NAME>` for you. This tool was designed specifically not to complete should it fail at any point along the way. It will produce logging output to stdout, so if you are having trouble debugging, you might want to start there. It is recommended that you save all this logging information to a file. You can achieve this by running `$ jet --environment=staging 2>> deployment.log`.

## Questions, Comments, Concerns, Feature/Enhancements?

Open an issue!