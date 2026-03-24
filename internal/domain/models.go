package domain

import "time"

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

type BookingStatus string

const (
	BookingStatusActive    BookingStatus = "active"
	BookingStatusCancelled BookingStatus = "cancelled"
)

type User struct {
	ID           string     `json:"id"`
	Email        string     `json:"email"`
	Role         Role       `json:"role"`
	PasswordHash *string    `json:"-"`
	CreatedAt    *time.Time `json:"createdAt,omitempty"`
}

type Room struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	Capacity    *int       `json:"capacity,omitempty"`
	CreatedAt   *time.Time `json:"createdAt,omitempty"`
}

type Schedule struct {
	ID         string     `json:"id,omitempty"`
	RoomID     string     `json:"roomId"`
	DaysOfWeek []int      `json:"daysOfWeek"`
	StartTime  string     `json:"startTime"`
	EndTime    string     `json:"endTime"`
	CreatedAt  *time.Time `json:"createdAt,omitempty"`
}

type Slot struct {
	ID        string    `json:"id"`
	RoomID    string    `json:"roomId"`
	Start     time.Time `json:"start"`
	End       time.Time `json:"end"`
	CreatedAt time.Time `json:"-"`
}

type Booking struct {
	ID             string        `json:"id"`
	SlotID         string        `json:"slotId"`
	UserID         string        `json:"userId"`
	Status         BookingStatus `json:"status"`
	ConferenceLink *string       `json:"conferenceLink,omitempty"`
	CreatedAt      *time.Time    `json:"createdAt,omitempty"`
	SlotStart      *time.Time    `json:"-"`
}

type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
}
