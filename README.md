# MRT-API

MRT-API is a RESTful API service for managing and retrieving Jakarta MRT schedules, user authentication, and user reviews.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [API Endpoints](#api-endpoints)
- [License](#license)

## Features

- User authentication with JWT.
- Fetch MRT schedules by station and direction.
- User reviews for the MRT system.
- Caching for optimized data retrieval.
- Admin and user roles with different token expiration times.

## Installation

1. Clone the repository:

    ```sh
    git clone https://github.com/albertus125/mrt-API.git
    cd mrt-API
    ```

2. Install the dependencies:

    ```sh
    go mod tidy
    ```

3. Set up the PostgreSQL database and apply the necessary migrations.

## Configuration

1. Create a `.env` file in the config folder and add the following environment variables:

    ```env
    DB_HOST=your_database_host
    DB_PORT=your_database_port
    DB_USER=your_database_user
    DB_PASSWORD=your_database_password
    DB_NAME=your_database_name
    ```

2. Update the `config` package to read from the `.env` file.

## Usage

1. Run the server:

    ```sh
    go run main.go
    ```

2. The API will be available at `http://localhost:8080`.

### API Availability
- **Local Environment**: The API will be available at `http://localhost:8080`.
- **Production Environment**: The API is also available at `https://mrt-api-production.up.railway.app/api`.
## API Endpoints
To access these API endpoints, you need to create an account and use the token provided in the Authorization header. Follow these steps:
### Authentication

- **Register User**

    ```http
    POST /api/register
    ```

    Request body:

    ```json
    {
      "username": "user1",
      "password": "password"
    }
    ```

    Response:

    ```json
    {
    "message": "User registered"
    }
    ```
- **Login User**

    ```http
    POST /api/login
    ```

    Request body:

    ```json
    {
      "username": "user1",
      "password": "password"
    }
    ```

    Response:

    ```json
    {
      "token": "your_jwt_token"
    }
    ```
  - **Get All review**
    `http
    GET /api/review
    ```
   Response:

    ```json
      {
        "id": 1,
        "user_id": 0,
        "rating": 4.5,
        "comment": "Aplikasi sangat membantu!",
        "created_at": "2024-05-29T13:18:08.209673Z"
      },
      ...
    ```
Include the token in the Authorization header for all subsequent API requests.
Example request:
  ```http
  GET /api/schedules/:id/:arah
  Authorization: Bearer your-jwt-token
  ```
### Schedules
- **Get All Schedules**
    ```http
    GET /api/schedules/
    ```
   Response: 
   ```json
    [
      {
        "id": 48185,
        "station_id": 20,
        "stasiun_name": "Stasiun Lebak Bulus Grab",
        "arah": "Arah Bundaran HI",
        "jadwal": "05:00"
      },
      ...
    ]
    ```
- **Get Schedules by Station ID**

    Path parameters:

    - `id`: Station ID
    ```http
    GET /api/schedules/:id
    ```
   Response: 
   ```json
    [
      {
        "id": 48327,
        "station_id": 21,
        "stasiun_name": "Stasiun Fatmawati Indomaret",
        "arah": "Arah Lebak Bulus",
        "jadwal": "05:32"
      },
      ...
    ]
    ```
- **Get Schedules by Station ID and Direction**
    ```http
    GET /api/schedules/:id/:arah
    ```

    Path parameters:

    - `id`: Station ID
    - `arah`: Direction(must be either "Arah Bundaran HI" or "Arah Lebak Bulus")

    Response:

    ```json
    [
      {
        "id": 48185,
        "station_id": 20,
        "stasiun_name": "Stasiun Lebak Bulus Grab",
        "arah": "Arah Bundaran HI",
        "jadwal": "05:00"
      },
      ...
    ]
    ```
### Stasiun
- **Get All Stasiun**
    ```http
    GET /api/stasiun/
    ```
   Response: 
   ```json
    [
      {
        "id": 20,
        "stasiun_name": "Lebak Bulus Grab"
    },
      ...
    ]

### Reviews

- **Add Review**

    ```http
    POST /api/review
    ```

    Request body:

    ```json
    {
        "rating": 4.5,
         "comment": "Aplikasi sangat membantu!"
    }
    ```

    Response:

    ```json
    {
      "id": 1,
    "user_id": 1,
    "rating": 4.5,
    "comment": "Aplikasi sangat membantu!",
    "created_at": "2024-05-29T13:18:08.209673Z"
    }
    ```

### Caching

- **Get All Stations with Caching**

    ```http
    GET /api/stations
    ```

    Response:

    ```json
    [
      {
        "id": 1,
        "name": "Station 1"
      },
      ...
    ]
    ```
- **Get All Schedules with Caching**
    ```http
    GET /api/schedules/
    ```
   Response: 
   ```json
    [
      {
        "id": 48185,
        "station_id": 20,
        "stasiun_name": "Stasiun Lebak Bulus Grab",
        "arah": "Arah Bundaran HI",
        "jadwal": "05:00"
      },
      ...
    ]
    ```
    - **Get Schedules by Station ID with Caching**

    Path parameters:

    - `id`: Station ID
    ```http
    GET /api/schedules/:id
    ```
   Response: 
   ```json
    [
      {
        "id": 48327,
        "station_id": 21,
        "stasiun_name": "Stasiun Fatmawati Indomaret",
        "arah": "Arah Lebak Bulus",
        "jadwal": "05:32"
      },
      ...
    ]
    ```
    - **Get Schedules by Station ID and Direction with Caching**
    ```http
    GET /api/schedules/:id/:arah
    ```

    Path parameters:

    - `id`: Station ID
    - `arah`: Direction(must be either "Arah Bundaran HI" or "Arah Lebak Bulus")

    Response:

    ```json
    [
      {
        "id": 48185,
        "station_id": 20,
        "stasiun_name": "Stasiun Lebak Bulus Grab",
        "arah": "Arah Bundaran HI",
        "jadwal": "05:00"
      },
      ...
    ]
    ```
## License

This project is licensed under the MIT License.
