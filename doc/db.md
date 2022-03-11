# Database origin
RedPlant can be used to map databases as origins, so that they can be queried as REST APIs.
Mind that this is not a general solution to expose DB-backed APIs, but a tool to be used mostly in development or testing.

## Setup
Configure a route as follows:
```yaml
"/db":
    origin: "mysql://root:example@localhost:3306/example2"
```
Where the value of `origin` the URI of your database.


## Support
Currently, the available drivers are:
* PostgreSQL
* MySQL

## Usage
At this point any HTTP call against `/db` will assume the request body contains a **SQL statement**. The response body
will contain a JSON version of the queried data.
All transformers and sidecars apply just like an HTTP origin.