# snptx

![Go](https://github.com/tullo/snptx/workflows/Go/badge.svg)
![CodeQL](https://github.com/tullo/snptx/workflows/CodeQL/badge.svg)
[![codecov](https://codecov.io/gh/tullo/snptx/branch/master/graph/badge.svg?token=R891ZHOLF6)](https://codecov.io/gh/tullo/snptx)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/tullo/snptx)
[![Total alerts](https://img.shields.io/lgtm/alerts/g/tullo/snptx.svg?logo=lgtm&logoWidth=18)](https://lgtm.com/projects/g/tullo/snptx/alerts/)
[![Language grade: JavaScript](https://img.shields.io/lgtm/grade/javascript/g/tullo/snptx.svg?logo=lgtm&logoWidth=18)](https://lgtm.com/projects/g/tullo/snptx/context:javascript)

## Webapp build & start

To launch the db & app with seed data (deps: docker, docker-compose) run:

1. `make run`
2. https://snptx.127.0.0.1.nip.io:4200/

## Topics covered

- Project structure and organization
- Debug mode
- Configuration and error handling
  - Managing configuration settings
  - Leveled logging
  - Dependency injection
  - Centralized error handling
  - Isolating application routes
- Database-driven responses
  - Database setup
  - Database driver installation
  - Database connection pool creation
  - Database model design
  - SQL statements execution
  - Single-record SQL queries
  - Multiple-record SQL queries
  - Transactions and other details
- HTML templating and inheritance
  - Dynamic HTML templates
  - Serving static files
  - Displaying dynamic data
  - Template actions and functions
  - Caching templates
  - Catching runtime rrrors
  - Common dynamic data
  - Custom template functions
- Middleware
  - Composable middleware chains
  - Setting security headers
  - Request logging
  - Panic recovery
- RESTful routing
  - Router installation
  - RESTful routes implemention
- Processing Forms
  - Setting up forms
  - Parsing form data
  - Data validation
  - Scaling data validation
- Stateful HTTP
  - Session manager installation
  - Session manager setup
  - Working with session data
- Security improvements
  - HTTPS server setup
  - HTTPS settings configuation 
  - Connection timeouts
- User authN
  - Routes setup
  - Users model creation
  - User signup and password encryption
  - User login
  - User logout
  - User authZ
  - CSRF protection
- Request context for authN/authZ
- Testing
  - Unit testing and sub-tests
  - Testing HTTP handlers
  - End-To-End testing
  - Mocking dependencies
  - Testing HTML forms
  - Integration testing
  - Profiling test coverage
- Pages
  - About
  - Change Password
  - Home
  - Ping (status/uptime monitoring)
  - User login
  - User logout
  - User profile
  - User signup
  - Snippet (form|display)
