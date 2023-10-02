package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	_ "github.com/lib/pq"
)

var dbConn *sql.DB

func main() {
	dbConn = connectToDB()
	defer dbConn.Close()

	e := echo.New()

	e.Use(middleware.Logger())
	e.Logger.SetLevel(log.DEBUG)

	e.GET("/events", handleGetAllEvents)
	e.GET("/event/:id", handleGetEventById)
	e.DELETE("/event/:id", handleDeleteEventById)
	e.POST("/event", handleCreateEvent)
	e.PUT("/event/:id", handleUpdateEventById)

	e.Logger.Fatal(e.Start(":" + os.Getenv("APP_PORT")))
}

func handleGetAllEvents(c echo.Context) error {
	rows, err := dbConn.Query("SELECT * FROM Events")
	if err != nil {
		return err
	}

	defer rows.Close()

	events := make([]Event, 0)
	for rows.Next() {
		event, err := newEventFromRow(rows)
		if err != nil {
			c.Logger().Error(err)
		} else {
			events = append(events, event)
		}
	}

	response := EventsResponse{Events: events}
	return c.JSON(http.StatusOK, &response)
}

func handleGetEventById(c echo.Context) error {
	id := c.Param("id")
	rows, err := dbConn.Query("SELECT * FROM Events WHERE id = $1", id)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		return c.NoContent(http.StatusNotFound)
	}

	event, err := newEventFromRow(rows)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &event)
}

func handleDeleteEventById(c echo.Context) error {
	id := c.Param("id")
	_, err := dbConn.Exec("DELETE FROM Events WHERE id = $1", id)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func handleCreateEvent(c echo.Context) error {
	bodyReader := c.Request().Body
	bytes, err := io.ReadAll(bodyReader)
	if err != nil {
		return err
	}

	event := Event{}
	err = json.Unmarshal(bytes, &event)
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	id := uuid.New().String()
	_, err = dbConn.Exec("INSERT INTO Events (id, title, author, date) VALUES ($1, $2, $3, $4)", id, event.Title, event.Author, event.Date)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &EventIdResponse{Id: id})
}

func handleUpdateEventById(c echo.Context) error {
	id := c.Param("id")
	bodyReader := c.Request().Body
	bytes, err := io.ReadAll(bodyReader)
	if err != nil {
		return err
	}

	event := Event{}
	err = json.Unmarshal(bytes, &event)
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	result, err := dbConn.Exec("UPDATE Events SET title = $1, author = $2, date = $3 WHERE id = $4", event.Title, event.Author, event.Date, id)
	if err != nil {
		return err
	} else {
		affected, err := result.RowsAffected()
		if err == nil && affected == 0 {
			return c.NoContent(http.StatusNotFound)
		}
	}

	return c.NoContent(http.StatusOK)
}

func newEventFromRow(rows *sql.Rows) (Event, error) {
	event := Event{}
	err := rows.Scan(&event.Id, &event.Title, &event.Author, &event.Date)
	return event, err
}

func connectToDB() *sql.DB {
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s port=%s sslmode=disable host=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_HOST"),
	)
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	return conn
}

type Event struct {
	Id     string
	Title  string
	Author string
	Date   string
}

type EventsResponse struct {
	Events []Event `json:"events"`
}

type EventIdResponse struct {
	Id string `json:"id"`
}
