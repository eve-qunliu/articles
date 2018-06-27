## Steps to run the program
#### 1. make deps

This is to install dependencies

#### 2. make migrateUp

This is to migrate database schema

#### 3. make start

This will start http server listening on port 8080

#### 4. Testing endpoints
Create one article
```
curl -XPOST "http://localhost:8080/articles" -d'{"title":"z1","body":"body","date":"2018-06-12","tags":["sports","Music","music"]}'
```

Create another article
```
curl -XPOST "http://localhost:8080/articles" -d'{"title":"z2","body":"body","date":"2018-06-12","tags":["drama","sports","Music","music"]}'
```

Get the first article
```
curl -XGET "http://localhost:8080/articles/1"
```

Get the second article
```
curl -XGET "http://localhost:8080/articles/2"
```

Get tag on specific date
```
curl -XGET "http://localhost:8080/tag/sports/20180612"
```

#### 5. make test
This will run testing cases

### Next Step
#### 1. Use cache to improve performance.

#### 2. When error happens, use meaning response body.
