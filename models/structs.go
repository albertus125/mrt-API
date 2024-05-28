package models

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Schedule struct {
	ID          int    `json:"id"`
	StasiunID   int    `json:"station_id"`
	StasiunName string `json:"stasiun_name"`
	Arah        string `json:"arah"`
	Jadwal      string `json:"jadwal"`
}

type Stasiun struct {
	StasiunID   int    `json:"id"`
	StasiunName string `json:"stasiun_name"`
}

type User struct {
	ID       int    `json:"id" gorm:"primary_key"`
	Username string `json:"username" gorm:"unique"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type Claims struct {
	Username string `json:"username"`
	UserID   int    `json:"user_id"`
	Role     string `json:"role"`
	jwt.StandardClaims
}

type Review struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Rating    float64   `json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}
