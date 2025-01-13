# Splitwise Clone (Backend Only)

This repository contains the backend code for a simplified Splitwise clone. The project focuses on implementing basic expense-sharing functionalities without a frontend. Below is an overview of the project and setup instructions.

## Features

- User management
- Group creation and management
- Expense creation and tracking
- Splitting expenses between group members
- Calculating balances owed or owed-to by users
- Caching for improved performance using Redis

## Tech Stack

- **Programming Language:** Go (Golang)
- **Web Framework:** Gin
- **Database:** PostgreSQL
- **Caching:** Redis
- **Dependency Management:** Go modules

## Setup Instructions

### Prerequisites

Make sure you have the following installed:

- [Go](https://golang.org/dl/) (v1.23 or higher recommended)
- [PostgreSQL](https://www.postgresql.org/)
- [Redis](https://redis.io/)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/kaankulac/gin-splitwise.git
   cd gin-splitwise
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Set up your configuration using the provided config/local.yaml file in the root directory. You can modify this file to include your specific settings for the database and Redis.
   ```env
    server:
    port: 8080
    db:
    dataSourceName: postgres://postgres:123123@localhost:5432/splitwise?sslmode=disable
    redis:
    host: localhost
    port: 6379
   ```
4. Start the server:
   ```bash
   go run cmd/server/main.go
   ```

The server will be running at `http://localhost:8080` by default.

## API Endpoints

Here is a brief overview of the main endpoints:

### Users

- `POST /users/signup`: Register
- `POST /users/login`: Login
- `GET /users/profile`: Get user profile
- `PUT /users/profile`: Update user profile
### Groups

- `POST /groups`: Create a new group
- `PUT /groups/:groupId`: Update group
- `DELETE /groups/:groupId`: Delete group
- `POST /groups/members`: Add a new group member
- `POST /groups/:groupId/leave`: Leave group

### Expenses

- `GET /expenses/:groupId`: Get expenses by Group ID
- `POST /expenses`: Create a new expense to a group
- `DELETE /expenses/:expenseId`: Delete expense
- `POST /expenses/:expenseId/pay`: Pay expense

## Notes

- Ensure that PostgreSQL and Redis are running before starting the server.

## License

This project is licensed under the MIT License. See the LICENSE file for details.

