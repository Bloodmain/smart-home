package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"homework/internal/domain"
	"homework/internal/gateways/http/models"
	"homework/internal/usecase"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-openapi/strfmt"

	"github.com/gin-gonic/gin"
	"github.com/jeanfric/goembed/countingwriter"
)

func setupRouter(r *gin.Engine, uc UseCases, ws *WebSocketHandler) {
	r.HandleMethodNotAllowed = true

	r.POST("/events", setupPostEventHandler(uc))
	r.OPTIONS("/events", setupOptionsEventHandler())
	r.GET("/sensors", setupGetSensorHandler(uc))
	r.HEAD("/sensors", setupHeadSensorHandler(uc))
	r.POST("/sensors", setupPostSensorHandler(uc))
	r.OPTIONS("/sensors", setupOptionsSensorHandler())
	r.GET("/sensors/:sensor_id", setupGetSensorIdHandler(uc))
	r.HEAD("/sensors/:sensor_id", setupHeadSensorIdHandler(uc))
	r.OPTIONS("/sensors/:sensor_id", setupOptionsSensorIdHandler())
	r.OPTIONS("/users", setupOptionsUserHandler())
	r.POST("/users", setupPostUserHandler(uc))
	r.POST("/users/:user_id/sensors", setupPostUserIdHandler(uc))
	r.HEAD("/users/:user_id/sensors", setupHeadUserIdHandler(uc))
	r.OPTIONS("/users/:user_id/sensors", setupOptionsUserIdHandler())
	r.GET("/users/:user_id/sensors", setupGetUserIdHandler(uc))
	r.GET("/sensors/:sensor_id/events", setupGetSensorEventHandler(ws))
}

func setupGetSensorEventHandler(ws *WebSocketHandler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("sensor_id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnprocessableEntity)
			return
		}

		if err := ws.Handle(ctx, id); err != nil {
			if errors.Is(err, usecase.ErrSensorNotFound) {
				ctx.AbortWithStatus(http.StatusNotFound)
			} else {
				ctx.AbortWithStatus(http.StatusInternalServerError)
			}
		}
	}
}

func checkAccept(ctx *gin.Context) bool {
	if ctx.GetHeader("Accept") != "application/json" {
		ctx.AbortWithStatus(http.StatusNotAcceptable)
		return false
	}
	return true
}

func checkContentType(ctx *gin.Context) bool {
	if ctx.GetHeader("Content-Type") != "application/json" {
		ctx.AbortWithStatus(http.StatusUnsupportedMediaType)
		return false
	}
	return true
}

func getSensorsDto(items ...domain.Sensor) []models.Sensor {
	itemsDto := make([]models.Sensor, len(items))

	for i, it := range items {
		item := it
		name := string(item.Type)
		itemsDto[i] = models.Sensor{
			CurrentState: &item.CurrentState,
			Description:  &item.Description,
			ID:           &item.ID,
			IsActive:     &item.IsActive,
			LastActivity: (*strfmt.DateTime)(&item.LastActivity),
			RegisteredAt: (*strfmt.DateTime)(&item.RegisteredAt),
			SerialNumber: &item.SerialNumber,
			Type:         &name,
		}
	}
	return itemsDto
}

func setupGetSensorHandler(uc UseCases) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !checkAccept(ctx) {
			return
		}
		items, err := uc.Sensor.GetSensors(ctx)
		if err != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		ctx.JSON(http.StatusOK, getSensorsDto(items...))
	}
}

func setupGetSensorIdHandler(uc UseCases) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !checkAccept(ctx) {
			return
		}
		id, err := strconv.ParseInt(ctx.Param("sensor_id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnprocessableEntity)
			return
		}
		s, err := uc.Sensor.GetSensorByID(ctx, id)
		if err != nil {
			if errors.Is(err, usecase.ErrSensorNotFound) {
				ctx.AbortWithStatus(http.StatusNotFound)
			} else {
				ctx.AbortWithStatus(http.StatusInternalServerError)
			}
			return
		}
		ctx.JSON(http.StatusOK, getSensorsDto(*s))
	}
}

func setupGetUserIdHandler(uc UseCases) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !checkAccept(ctx) {
			return
		}
		id, err := strconv.ParseInt(ctx.Param("user_id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnprocessableEntity)
			return
		}
		items, err := uc.User.GetUserSensors(ctx, id)
		if err != nil {
			if errors.Is(err, usecase.ErrUserNotFound) {
				ctx.AbortWithStatus(http.StatusNotFound)
			} else {
				ctx.AbortWithStatus(http.StatusInternalServerError)
			}
			return
		}
		ctx.JSON(http.StatusOK, getSensorsDto(items...))
	}
}

func setContentLength(ctx *gin.Context, items ...domain.Sensor) {
	cW := countingwriter.New(io.Discard)
	for _, item := range getSensorsDto(items...) {
		e := json.NewEncoder(cW).Encode(item)
		if e != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}

	ctx.Header("Content-Length", fmt.Sprintf("%d", cW.BytesWritten()))
	ctx.Status(http.StatusOK)
}

func setupHeadSensorHandler(uc UseCases) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !checkAccept(ctx) {
			return
		}
		items, err := uc.Sensor.GetSensors(ctx)
		if err != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		setContentLength(ctx, items...)
	}
}

func setupHeadSensorIdHandler(uc UseCases) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !checkAccept(ctx) {
			return
		}
		id, err := strconv.ParseInt(ctx.Param("sensor_id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnprocessableEntity)
			return
		}
		s, err := uc.Sensor.GetSensorByID(ctx, id)
		if err != nil {
			if errors.Is(err, usecase.ErrSensorNotFound) {
				ctx.AbortWithStatus(http.StatusNotFound)
			} else {
				ctx.AbortWithStatus(http.StatusInternalServerError)
			}
			return
		}
		setContentLength(ctx, *s)
	}
}

func setupHeadUserIdHandler(uc UseCases) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !checkAccept(ctx) {
			return
		}
		id, err := strconv.ParseInt(ctx.Param("user_id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnprocessableEntity)
			return
		}
		items, err := uc.User.GetUserSensors(ctx, id)
		if err != nil {
			if errors.Is(err, usecase.ErrUserNotFound) {
				ctx.AbortWithStatus(http.StatusNotFound)
			} else {
				ctx.AbortWithStatus(http.StatusInternalServerError)
			}
			return
		}
		setContentLength(ctx, items...)
	}
}

type validatable interface {
	*models.SensorEvent | *models.SensorToCreate | *models.UserToCreate | *models.SensorToUserBinding
	Validate(formats strfmt.Registry) error
}

func bindAndValidate[T validatable](ctx *gin.Context, item T) bool {
	if err := ctx.ShouldBindJSON(item); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return false
	}

	if err := item.Validate(nil); err != nil {
		ctx.AbortWithStatus(http.StatusUnprocessableEntity)
		return false
	}
	return true
}

func setupPostEventHandler(uc UseCases) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !checkContentType(ctx) {
			return
		}
		e := models.SensorEvent{}
		if !bindAndValidate(ctx, &e) {
			return
		}

		newEvent := domain.Event{SensorSerialNumber: *e.SensorSerialNumber, Payload: *e.Payload, Timestamp: time.Now()}
		if err := uc.Event.ReceiveEvent(ctx, &newEvent); err != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
		} else {
			ctx.Status(http.StatusCreated)
		}
	}
}

func setupPostSensorHandler(uc UseCases) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !checkContentType(ctx) {
			return
		}
		e := models.SensorToCreate{}
		if !bindAndValidate(ctx, &e) {
			return
		}
		newItem := domain.Sensor{
			SerialNumber: *e.SerialNumber, Description: *e.Description,
			IsActive: *e.IsActive, Type: domain.SensorType(*e.Type),
		}
		if item, err := uc.Sensor.RegisterSensor(ctx, &newItem); err != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
		} else {
			ctx.JSON(http.StatusOK, getSensorsDto(*item))
		}
	}
}

func setupPostUserHandler(uc UseCases) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !checkContentType(ctx) {
			return
		}
		e := models.UserToCreate{}
		if !bindAndValidate(ctx, &e) {
			return
		}
		newItem := domain.User{Name: *e.Name}
		if u, err := uc.User.RegisterUser(ctx, &newItem); err != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
		} else {
			ctx.JSON(http.StatusOK, models.User{
				ID:   &u.ID,
				Name: &u.Name,
			})
		}
	}
}

func setupPostUserIdHandler(uc UseCases) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !checkContentType(ctx) {
			return
		}
		userId, err := strconv.ParseInt(ctx.Param("user_id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnprocessableEntity)
			return
		}
		e := models.SensorToUserBinding{}
		if !bindAndValidate(ctx, &e) {
			return
		}

		err = uc.User.AttachSensorToUser(ctx, userId, *e.SensorID)
		if err != nil {
			if errors.Is(err, usecase.ErrUserNotFound) || errors.Is(err, usecase.ErrSensorNotFound) {
				ctx.AbortWithStatus(http.StatusNotFound)
			} else {
				ctx.AbortWithStatus(http.StatusInternalServerError)
			}
			return
		}
		ctx.Status(http.StatusCreated)
	}
}

func setupOptionsEventHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Header("Allow", strings.Join([]string{http.MethodOptions, http.MethodPost}, ","))
		ctx.Status(http.StatusNoContent)
	}
}

func setupOptionsSensorHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Header("Allow", strings.Join([]string{http.MethodOptions, http.MethodPost, http.MethodGet, http.MethodHead}, ","))
		ctx.Status(http.StatusNoContent)
	}
}

func setupOptionsSensorIdHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Header("Allow", strings.Join([]string{http.MethodOptions, http.MethodGet, http.MethodHead}, ","))
		ctx.Status(http.StatusNoContent)
	}
}

func setupOptionsUserHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Header("Allow", strings.Join([]string{http.MethodOptions, http.MethodPost}, ","))
		ctx.Status(http.StatusNoContent)
	}
}

func setupOptionsUserIdHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Header("Allow", strings.Join([]string{http.MethodOptions, http.MethodPost, http.MethodGet, http.MethodHead}, ","))
		ctx.Status(http.StatusNoContent)
	}
}
