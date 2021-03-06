# INTERSTELLAR
HTTP server that exec code and use redis to host/url mapping. You can use middleware feature as chain of functions. Tested with: golang, gravity, rust, nodejs, bash and php.     
It's an experiment to define a flexible microservices proxy. Microservices run as commands and are in files, or in redis. You can use flexible language, no needs hot reload, share middleware from different project, easy deploy.      

### Features
- multiple languages backend
- middlewares (different languages too)
- http proxy
- code stored in redis
- not reload for new codes
- status health check
- maintenance mode for single route
- custom response content type
- trigger that exec commands
- basic authentication for single route
- optional gzip success response
- rate limiter on route based on ip or header
- restrictions on route based on ip or header
- body and querystring validation
- detachable jobs
- process only if status is setted ready
- ping health check
- minimal requirements

### Requirements
Node.JS v8 and redis

### Install
`git clone git@github.com:anddimario/interstellar.git`    
`npm install`     
`cp .env.example .env`    
`node app.js`     

### Initial status
When an Interstellar instance starts, it sets a redis variable in the form: 
`interstellar:instances:HOSTNAME`, initialize this variable with `INITIAL_STATUS` and set a ttl (60 sec), useful to check instance health status.    
In .env the `INITIAL_STATUS` is used to manage the possibility to process the request, default is `ready` so, when started, the instance could serve the request immediately, but you can set this status as you want. For example, a scenario where you mount your compiled files from a shared disk and you want wait this mount and, only then, serve the request. You can set `INITIAL_STATUS` as `waiting`, so, when the istance starts, the instance doesn't serve request and then, when you mount the disk, you could set the status as `ready` in redis with:     
`set interstellar:instances:HOSTNAME ready`     
Another scenario is when you deploy commit that affect more files and you want disable routing for single instance, indifferently from your `INITIAL_STATUS`, you can set status as `waiting` in redis with:     
`set interstellar:instances:HOSTNAME waiting`    
When done, allow request with:     
`set interstellar:instances:HOSTNAME ready`     

### Add host routing
Must add redis key in hash with this format: interstellar:vhost:HOSTNAME:URL     
Example:      
`redis-cli`     
`hset interstellar:vhost:localhost:3000:/ commands "command1,command2,..."`      
`hset interstellar:vhost:localhost:3000:/ method GET`      

### Gravity basic example
- Follow the [gravity install guide](https://marcobambini.github.io/gravity/getting-started.html) for start
- After you have cloned and make gravity (suppose that we clone it in /home/myuser/gravity), create an example file (mytest.gravity) in the gravity directory with your code
- Add rules in redis     
`hset interstellar:vhost:localhost:3000:/ method GET`      
`hset interstellar:vhost:localhost:3000:/ commands "cd /home/myuser/gravity && ./gravity mytest.gravity"`      
- Test with curl    
`curl http://localhost:3000/`

### Arguments passed to commands (middlewares, body, querystring and headers)
In this example in golang with a put request there's how interstellar pass arguments to code with useful informations. 
- Create a middleware mid.go with the code:
```
package main

import (
        "encoding/json"
	"fmt"
	"os"
)

func main() {
	// get last argument (the interstellar argument json)
	last := len(os.Args) - 1
        // Get and decode the json body passed by arguments
        byt := []byte(os.Args[last])
        var dat map[string]map[string]interface{}
        if err := json.Unmarshal(byt, &dat); err != nil {
                panic(err)
        }
	// get id from querystring and compare
	if dat["querystring"]["id"] == "foo" {
		// return
		fmt.Print("Validation done")
	} else {
		fmt.Print("MiddlewareFailedValidation Failed")
	}
}

```
and main.go:
```
package main

import (
        "encoding/json"
	"fmt"
	"os"
)

func main() {
	// get last argument (the interstellar argument json)
	last := len(os.Args) - 1
        // Get and decode the json body passed by arguments
        byt := []byte(os.Args[last])
        var dat map[string]map[string]interface{}
        if err := json.Unmarshal(byt, &dat); err != nil {
                panic(err)
        }
	// return
	fmt.Print(dat["headers"]["host"])
	fmt.Print(" had ")
	fmt.Print(dat["middlewares"]["1"])
	fmt.Print(" for ")
	fmt.Print(dat["body"]["field"])
}
```
**MUST KNOW** Exec is blocked if middleware return the `MIDDLEWARE_OUTPUT_FAILED` env variable value in stdout, or return no empty stderr, or there's an error. If middleware return the text setted in env variable `MIDDLEWARE_OUTPUT_SKIP`, the middleware output is not stored on arguments.
- Compile them
- Then we can add a route for this, with:     
```
hset interstellar:vhost:localhost:3000:/test method PUT      
hset interstellar:vhost:localhost:3000:/test commands "cd /home/myuser/mybinary && ./mid,cd /home/myuser/mybinary && ./main"
```      
**Note** The function order is important because are executed in this order, and use commas to separate them
- Test with curl    
`curl -XPUT -d "field=mytest" http://localhost:3000/test?id=foo`   

You should see in the response the body, the middleware response ("Validation done") and the host in this form: "localhost:3000 had Validation done for mytest".     
Try now with: `curl -d "field=mytest" http://localhost:3000/test?id=bar` and see what happen. The middleware gave back an error from this line:    
`fmt.Print("MiddlewareFailedValidation Failed")`     
As you can see, `MiddlewareFailed` is the key used to recognize the error and it is set in .env, `Validation Failed` is the output from middleware.    
**MUST KNOW** the arguments are passed in a stringify json in the last position of the commands and there're:    
**headers**: all headers defined for env in `ARGUMENT_HEADERS` (listed as comma separeted string with header name)       
**body**: the request body as json (optional, if is passed from the request)    
**querystring**: the request querystring as json (optional, if is passed from the request)   
**middlewares**: an object with all middleware output not skipped (optional)

### Code from redis (example with node.js)
In redis run this commands:
```
hset interstellar:vhost:localhost:3000:/redis commands "node -e CUSTOM_CODE"
hset interstellar:vhost:localhost:3000:/redis method GET
hset interstellar:vhost:localhost:3000:/redis code0 "const test='INTERSTELLAR.VARIABLES';const parsed=JSON.parse(decodeURIComponent(test));console.log(parsed.headers.host);"
```
Then with curl: `curl localhost:3000/redis`    
Try php:
```
hset interstellar:vhost:localhost:3000:/redis commands "php -r CUSTOM_CODE"
hset interstellar:vhost:localhost:3000:/redis code0 "\\$test='INTERSTELLAR.VARIABLES';\\$decoded=urldecode(\\$test);echo \\$decoded;"
```
Try both:   
```
hset interstellar:vhost:localhost:3000:/redis commands "node -e CUSTOM_CODE,php -r CUSTOM_CODE"
hset interstellar:vhost:localhost:3000:/redis code0 "const test='INTERSTELLAR.VARIABLES';const parsed=JSON.parse(decodeURIComponent(test));console.log(parsed.headers.host);"
hset interstellar:vhost:localhost:3000:/redis code1 "\\$test='INTERSTELLAR.VARIABLES';\\$decoded=urldecode(\\$test);echo \\$decoded;"
```
__IMP__ Note that `CUSTOM_CODE` in commands is where interstellare replace your code in commands and `INTERSTELLAR.VARIABLES` is replaced from interstellar variables object, with headers, body, querystring and middleware response. The `INTERSTELLAR.VARIABLES` is encoded with `encodeURIComponent()` this allow a simple escaped, but need `decodeURIComponent()`, as you can see in code, to get the object.    
Code are stored with index as commands' references, for example in "both" example, `code0` is used by `node -e CUSTOM_CODE` first element in commands.

### HTTP PROXY (example with [MICRO](https://github.com/zeit/micro))
Proxy to other listen service, for example, add in redis:
```
hset interstellar:vhost:localhost:8000:/micro method GET
hset interstellar:vhost:localhost:8000:/micro commands "http://localhost:3000"
hset interstellar:vhost:localhost:8000:/micro type http
```
Then create and run micro, as example in https://github.com/zeit/micro.
Try with curl: `curl http://localhost:8000/micro`     
**IMP** Attention to ports settings
**NOTE** As POST microservice example, you can use this: [MICRO POST](https://github.com/zeit/micro/tree/master/examples/urlencoded-body-parsing)

### Detachable jobs
They are long running jobs that get back a success message, but running in background, for example, add in redis:
```
hset interstellar:vhost:localhost:3000:/job method GET
hset interstellar:vhost:localhost:3000:/job commands "bash /path/test.sh"
hset interstellar:vhost:localhost:3000:/job type job
```
Then create test.sh with:
```
#!/usr/bin/env bash

sleep 180
touch /tmp/myjobcompleted
```
Try with curl: `curl http://localhost:3000/job`     
**LIMITS**
- not working with middleware system
- the command should be in the form: `command arg1 arg2 ...` last argument will be the interstellar informations 

### Validation (using go example above) (optional)
The validation use [https://github.com/chriso/validator.js](https://github.com/chriso/validator.js)    
For validations, in redis object must be validateBody (if you need to validate body) and validateQuery (if you need to validate querystring), the value is a JSON stringify array of object that could have different form:
- `{field:"name",validator:"isEmail"}`: basic validator, in this example isEmail
- `{field:"name",validator:"isIn",compare:[...]}`: `isIn` and maybe other validator has a different form, in this case compare is an array of strings, used for check
- `{field:"name",validator:"isLength",options:{....}}`: some validators, and one is `isLength`, could have options as object

Now add for example: 
```
hset interstellar:vhost:localhost:3000:/test validateBody '[{"field":"field","validator":"isIn","compare":["mytest","test"],"message":"Not in"}]'
hset interstellar:vhost:localhost:3000:/test validateQuery '[{"field":"id","validator":"isLength","options":{"min":"2"},"message":"Wrong length"}}]'
```
And try this curl:
```
curl -XPUT -d "field=mytest" http://localhost:3000/test?id=foo
curl -XPUT -d "field=myst" http://localhost:3000/test?id=foo
curl -XPUT -d "field=mytest" http://localhost:3000/test?id=f
```

### Setup response content type header (optional)
You can setup response content type header with this redis hset:    
`hset interstellar:vhost:localhost:3000:/ciao content_type application/json`    

### Status health check (optional)
Add in .env:
```
HEALTH_CHECK=true
HEALTH_CHECK_TYPE=
HEALTH_CHECK_MATCH=
```
Where `HEALTH_CHECK_TYPE` could be:
- *path*: if the reference is request.url
- *user-agent* if the reference is the client agent    

Example with user-agent (aws elb health check):
```
HEALTH_CHECK=true
HEALTH_CHECK_TYPE=user-agent
HEALTH_CHECK_MATCH=ELB-HealthChecker/1.0
```
Example with path:
```
HEALTH_CHECK=true
HEALTH_CHECK_TYPE=path
HEALTH_CHECK_MATCH=/interstellar/status
```

### Route maintenance mode (optional)
You can set a maintenance mode for route if necessary, add in redis for route:    
`hset interstellar:vhost:localhost:3000:/ maintenance true`    
Disable maintenance with:   
`hdel interstellar:vhost:localhost:3000:/ maintenance`   
To set multiple routes on maintenance, there's the scripts/maintenance. It can enable/disable the mode based on vhost, or if commands contains the string (useful if you have code shared in multiple routes that you will change), example:    
- Enable maintenance for all routes for example.com vhost: `node scripts/maintenance vhost example.com enable`   
- Enable maintenance for all routes that use `mid`: `node scripts/maintenance command mid enable`   
- Disable mode for previous routes: `node scripts/maintenance command mid disable`   

### Custom system messages (optional)
Define a custom content type response in .env with:    
`CUSTOM_RESPONSE_TYPE=`     
Then there are this defaults messages that you can customize with a variable in .env:    
- *redis error*: `MESSAGES_REDIS_ERROR` back when a redis error occurs   
- *not ready*: `MESSAGES_NOT_READY_ERROR` back when the application state is not ready   
- *UP*: `MESSAGES_HEALTH_OK` back from interstellar health check    
- *not found*: `MESSAGES_NOT_FOUND` back when route not found    
- *maintenance*: `MESSAGES_MAINTENANCE_ACTIVE` back if maintenance is active for route

### Basic authentication (optional)
You can set for a route if it's protected from basic auth with the redis key in the form `interstellar:vhost:HOST:PATH`, for example:     
`hset interstellar:vhost:localhost:3000:/ basicAuth true`     
Basic auth users are stored in redis and for each host, in the form: `interstellar:basic:auth:HOSTNAME:USER`, for example:    
`set interstellar:basic:auth:localhost:3000:john secret`   
You can test it on browser, or from curl: `curl http://john:secrset@localhost:3000`

### GZIP (optional)
In .env add: `GZIP=true`

### Rate Limiter (optional)
Set a limit for single route with, for example:    
`hset interstellar:vhost:localhost:3000:/ ratelimit "60,ip,5"`    
in the form: "time,reference,requests", where:
- time: interval in seconds where requests are checked
- reference: could be `ip`, to base limit on client ip, or an header name
- requests: number of requests allowed

### Restricted (optional)
Set a limit for single route with, for example:    
`hset interstellar:vhost:localhost:3000:/ restricted "x-interstellar:mykey1,mykey2"`   
in the form: "reference:value", where:
- reference: could be `ip`, to base restriction on client ip, or an header name
- value: is a comma separated list of allowed value    

Try the restriction above with: `curl localhost:3000 -H "x-interstellar:mykey1"`

### Errors logs
They are stored in redis in the form: `interstellar:logs:INSTANCE:TIMESTAMP string`

### Stats (optional)
To enable stats add in .env: `STATS=true`. You can see them in redis with: `KEYS interstellar:stats:*`     
Stats are for status code, instance, sites and in general.

### Triggers (optional)
To enable stats add in .env: `TRIGGERS=true`, stats are required. 
You can set a trigger that exec a commands when the thresold is reached, for example create a trigger with this command in redis:    
`hset interstellar:triggers my_great_trigger '{"min":5,"key":"interstellar:variables:triggers:test","thresold":5,"command":"touch /tmp/alert","global":true}'`     
In this way you have defined a trigger that exec `touch /tmp/alert` if there are 5 global request in 5 minutes.    
Options:
- `min`: the interval in minutes used to reference for trigger count that is refreshed when the time expire
- `key`: the temporary key used to store the count for this trigger
- `thresold`: when this count is reached the command is fired
- `command`: the command to fire when thresold is reached
- `global`: if it is setted as true, it's a global trigger that refer to all requests in all instances and for all sites (optional)
- `instance`: in the form: "instance":"hostname", it is used to watch requests for a specific instance (optional)
- `status`: for example: "status":200, it is used to fire trigger when the thresold based on request status is reached (optional)
- `site`: for example: "site":"example.com", watch the requests for a specific site

### Security
For security reason you can run commands using containers, or try: [nsjail](https://github.com/google/nsjail)    
Example: `nsjail -Mo --chroot / -q -- /path/to/your/file args`

### Example application
- [Microblog](https://github.com/anddimario/interstellar-microblog)  

### Benchmark
```
Intel(R) Celeron(R) CPU  N2840  @ 2.16GHz 32bits 2GB(RAM)
Ubuntu 16.04
Node v4.2.6
Redis 3.0.6
Siege 3.0.8
```
Backend [Microblog](https://github.com/anddimario/interstellar-microblog) 
```
siege -c50 -b -t1M http://localhost:3000/?title=prova
Transactions:		        5240 hits
Availability:		      100.00 %
Elapsed time:		       59.26 secs
Data transferred:	        0.02 MB
Response time:		        0.56 secs
Transaction rate:	       88.42 trans/sec
Throughput:		        0.00 MB/sec
Concurrency:		       49.73
Successful transactions:        5240
Failed transactions:	           0
Longest transaction:	        0.73
Shortest transaction:	        0.29
```
**NOTE** This test is done with all functionality (stats and triggers too), here a test without triggers and stats:
```
Transactions:		        7633 hits
Availability:		      100.00 %
Elapsed time:		       59.14 secs
Data transferred:	        0.04 MB
Response time:		        0.39 secs
Transaction rate:	      129.07 trans/sec
Throughput:		        0.00 MB/sec
Concurrency:		       49.80
Successful transactions:        7633
Failed transactions:	           0
Longest transaction:	        0.54
Shortest transaction:	        0.19
```
With micro (get data from sqlite)
```
Transactions:		       14116 hits
Availability:		      100.00 %
Elapsed time:		       59.58 secs
Data transferred:	        0.23 MB
Response time:		        0.21 secs
Transaction rate:	      236.93 trans/sec
Throughput:		        0.00 MB/sec
Concurrency:		       49.87
Successful transactions:       14116
Failed transactions:	           0
Longest transaction:	        0.47
Shortest transaction:	        0.09
```


### License
MIT
