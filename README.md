# MRT-api

An API for retrieving the Jakarta MRT schedule. This API is primarily utilized by the [website](https://cek-mrt.vercel.app/).

### How does it works?
This API uses a daily cron job, executed at midnight,  to scrape the Jakarta MRT schedule from the official PT.MRT JAKARTA website using Go library package [gocolly](github.com/gocolly/colly"). Subsequently, the data is processed, stored in a PostgreSQL database and cached using [go-cache](github.com/patrickmn/go-cache).

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
    git clone https://github.com/albertus125/mrt-api/.git
    cd mrt-api
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

2. The api will be available at `http://localhost:8080`.

### API Availability
- **Local Environment**: The api will be available at `http://localhost:8080`.
- **Production Environment**: The api is also available at `https://mrt-api/-production.up.railway.app/api/v1`.
## API Endpoints
To access these api endpoints, you need to create an account and use the token provided in the Authorization header. Follow these steps:
### Authentication

- **Register User**

    ```http
    POST /api/v1/register
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
    "data": {
        "id": 10,
        "username": "user1",
        "password": "your_jwt_token",
        "role": ""
    },
    "message": "User berhasil registrasi",
    "success": true
    }
    ```
- **Login User**

    ```http
    POST /api/v1/login
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
    "message": "login success",
    "success": true,
    "token": "your_jwt_token"
    }
    ```
  - **Get All review**
    ```http
    GET /api/v1/review
    ```
   Response:

    ```json
    [
      "data": [
        {
            "id": 1,
            "user_id": 0,
            "rating": 4.5,
            "comment": "Aplikasi sangat membantu!",
            "created_at": "2024-05-29T13:18:08.209673Z"
        },
        ...
      ],
      "message": "berhasil mengambil seluruh data review",
      "succes": true
      
    ]
    ```
Include the token in the Authorization header for all subsequent api requests.
Example request:
  ```http
  GET /api/v1/schedules/:id/:arah
  Authorization: Bearer your-jwt-token
  ```
### Schedules
- **Get All Schedules**
    ```http
    GET /api/v1/schedules/
    ```
   Response: 
   ```json
    [
      "message": "Sukses mengambil seluruh data schedules",
      "success": true,
      "data": [
        {
          "id": 48185,
          "station_id": 20,
          "stasiun_name": "Stasiun Lebak Bulus Grab",
          "arah": "Arah Bundaran HI",
          "jadwal": "05:00"
        },
      ...
      ]
    ]
    ```
- **Get Schedules by Station ID**

    Path parameters:

    - `id`: Station ID
    ```http
    GET /api/v1/schedules/:id
    ```
   Response: 
   ```json
    [
      {
      "message": "Data schedule dengan stasiun id21berhasil diambil",
      "success": true,
      "data": [
        {
            "id": 54350,
            "station_id": 21,
            "stasiun_name": "Stasiun Fatmawati Indomaret",
            "arah": "Arah Lebak Bulus",
            "jadwal": "05:32"
        },
      ...
      ]
       }
    ]
    ```
- **Get Schedules by Station ID and Direction**
    ```http
    GET /api/v1/schedules/:id/:arah
    ```

    Path parameters:

    - `id`: Station ID
    - `arah`: Direction(must be either "Arah Bundaran HI" or "Arah Lebak Bulus")

    Response:

    ```json
    [
      {
      "message": "Data schedule dengan stasiun ID: 21 dan arah: Arah Bundaran HI berhasil diambil",
      "success": true,
      "data": [
        {
            "id": 54493,
            "station_id": 21,
            "stasiun_name": "Stasiun Fatmawati Indomaret",
            "arah": "Arah Bundaran HI",
            "jadwal": "05:03"
        },
      ...
      ]
       }
    ]
    ```
### Stasiun
- **Get All Stasiun**
    ```http
    GET /api/v1/stasiun/
    ```
   Response: 
   ```json
    [
      {
      "message": "Berhasil mengambil semua data stasiun",
      "success": true,
      "data": [
          {
            "id": 20,
            "stasiun_name": "Lebak Bulus Grab"
          },
          ...
        ]
       }
       
    ]

### Reviews

- **Add Review**

    ```http
    POST /api/v1/review
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
    "data": [
        {
            "id": 1,
            "user_id": 0,
            "rating": 4.5,
            "comment": "Aplikasi sangat membantu!",
            "created_at": "2024-05-29T13:18:08.209673Z"
        }
    ],
    "message": "berhasil mengambil seluruh data review",
    "succes": true
    ```
    
## License

This project is licensed under the MIT License.
