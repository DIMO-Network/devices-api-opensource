package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// errorResponseHandler is deprecated. it doesn't log. We prefer to return an err and have the ErrorHandler in api.go handle stuff.
func errorResponseHandler(c *fiber.Ctx, err error, status int) error {
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	return c.Status(status).JSON(fiber.Map{
		"errorMessage": msg,
	})
}

func getUserID(c *fiber.Ctx) string {
	token := c.Locals("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	userID := claims["sub"].(string)
	return userID
}

// CreateResponse is a generic response with an ID of the created entity
type CreateResponse struct {
	ID string `json:"id"`
}

// grpcErrorToFiber useful anywhere calling a grpc underlying service and wanting to augment the error for fiber from grpc status codes
// meant to play nicely with the ErrorHandler in api.go that this would hand off errors to.
// msgAppend appends to error string, to eg. help if this gets logged
func grpcErrorToFiber(err error, msgAppend string) error {
	if err == nil {
		return nil
	}
	// pull out grpc error status to then convert to fiber http equivalent
	errStatus, _ := status.FromError(err)

	switch errStatus.Code() {
	case codes.InvalidArgument:
		return fiber.NewError(fiber.StatusBadRequest, errStatus.Message()+". "+msgAppend)
	case codes.NotFound:
		return fiber.NewError(fiber.StatusNotFound, errStatus.Message()+". "+msgAppend)
	case codes.Aborted:
		return fiber.NewError(fiber.StatusConflict, errStatus.Message()+". "+msgAppend)
	case codes.Internal:
		return fiber.NewError(fiber.StatusInternalServerError, errStatus.Message()+". "+msgAppend)
	}
	return errors.Wrap(err, msgAppend)
}
