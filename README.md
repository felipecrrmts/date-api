# date-app

## Packaging and running

Commands below must be executed from the project root

### Test

`make test`

### Build and Run

`make run` or `docker-compose up --build -d`

### Stop

`make stop` or `docker-compose down`


## Tech/Framework used

### Backend
<b>Scaffolding</b>
- [GO MOD INIT] (https://blog.golang.org/using-go-modules)

<b>Built with</b>
- [Go 1.22](https://golang.org/doc/go1.22)
- [Fiber](https://gofiber.io/)
- [MongoDB Go Driver](https://github.com/mongodb/mongo-go-driver)
- [Testify - Thou Shalt Write Tests](https://github.com/stretchr/testify)

## Assumptions

- discover 

  - throws an error if queries as wrong
  - by default is sorted by "distanceFromMe"
  - the attractiveness rank is uses if query "ranked" is provided with "true", rank sorts by:
    - most yes swiped gender
    - average of yes swiped age

- login
   - its using SigningMethodHS256
   - header Authorization: bearer token

- create user

  - [gofakeit](https://github.com/brianvoe/gofakeit/v7) is used to generate stub values

- tests

  - only unit tests in the service layer were created


### Endpoints

- POST /user/create
- POST /login
- GET /discover
- POST /swipe

### Postman collection
There is a Postman collection in the root to help with manual tests in the zip file `users.postman_collection.json`