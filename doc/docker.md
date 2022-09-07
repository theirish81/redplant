# Running on Docker
RedPlant is already packaged into Alpine container images, available on DockerHub.

Once you have a proper `etc` directory with the necessary configuration in it, you can issue (*NIX command):
```shell
docker run -v "$(pwd)/etc:/usr/local/redplant/etc" -p 9001:9001 theirish81/redplant
```
And you should be set to go.

**Reminder:** configuration changes do not live update, and you'll need to restart the server to make configuration
changes effective.