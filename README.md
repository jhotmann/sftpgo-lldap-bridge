# SFTPGo to LLDAP Bridge (ALPHA)
This is an external auth hook for SFTPGo that connects to LLDAP.

Features:  
- Map LLDAP groups to SFTPGo groups
- Set a group that is required for auth into SFTPGo
- Optional default SFTPGo group

In my personal setup, I have a group with the default settings configured (S3, some virtual folders, etc) and all users have that default group set as the primary group in SFTPGo so no user-specific configuration is necessary other than username and password. I then have other groups that are mapped to users via LLDAP as secondary groups for access to network shares and other data pools as virtual folders in SFTPGo. This has been tested under those conditions but since there are a miriad of ways to configure SFTPGo it would be good to test in other configurations before moving out of alpha.

## Instructions
Since this is still an alpha product, it should be tested in an environment similar to your production setup first.

1. Copy `config-example.yml` to `config.yml`
1. Configure any group mappings, default group, required group, etc (see comments in config file for more information)
1. Launch all 3 containers: `docker-compose up -d`
1. Configure users and groups in LLDAP `http://localhost:17170` (credentials in `docker-compose.yml`)
1. Configure SFTPGo groups `http://localhost:8080/web/admin` (credentials in `docker-compose.yml`)
1. Test logging into SFTPGo with your various users and groups
1. If you're happy with the results, add it to your production setup
