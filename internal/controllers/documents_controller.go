package controllers

import (
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

type DocumentsController struct {
	settings *config.Settings
	dbs      func() *database.DBReaderWriter
	log      *zerolog.Logger
}

// NewDocumentsController constructor
func NewDocumentsController(settings *config.Settings, dbs func() *database.DBReaderWriter, logger *zerolog.Logger) DocumentsController {
	return DocumentsController{
		settings: settings,
		dbs:      dbs,
		log:      logger,
	}
}

// GetDocuments godoc
// @Description  gets all documents associated with current user - pulled from token
// @Tags           documents
// @Produce      json
// @Accept       json
// @Success      200  {object}  []controllers.DocumentResponse
// @Security     BearerAuth
// @Router       /documents [get]
func (udc *DocumentsController) GetDocuments(c *fiber.Ctx) error {
	//userID := getUserID(c)
	//udc.settings.AWSBucketName
	documents := []DocumentResponse{
		{
			Name: "Driver's license",
		},
	}

	return c.JSON(documents)
}

// GetDocumentByID godoc
// @Description  get document by id associated with current user - pulled from token
// @Tags           documents
// @Produce      json
// @Accept       json
// @Param        id   path      int  true  "Document ID"
// @Success      200  {object}  controllers.DocumentResponse
// @Security     BearerAuth
// @Router       /documents/{id} [get]
func (udc *DocumentsController) GetDocumentByID(c *fiber.Ctx) error {
	document := DocumentResponse{Name: "Driver's license"}

	return c.JSON(document)
}

// PostDocument godoc
// @Description  post document by id associated with current user - pulled from token
// @Tags           documents
// @Produce      json
// @Accept       json
// @Param        user_device  body  controllers.DocumentRequest  true  "add document to user. either name and type are required"
// @Success      201  {object}  controllers.DocumentResponse
// @Security     BearerAuth
// @Router       /documents [post]
func (udc *DocumentsController) PostDocument(c *fiber.Ctx) error {

	file, err := c.FormFile("file")

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errorMessage": "invalid file document",
		})
	}

	// Get Buffer from file
	buffer, err := file.Open()

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errorMessage": "invalid file document",
		})
	}
	defer buffer.Close()

	document := new(DocumentRequest)

	if err := c.BodyParser(&document); err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(document)

	// // The session the S3 Uploader will use
	// sess := session.Must(session.NewSession())

	// // Create an uploader with the session and default options
	// uploader := s3manager.NewUploader(sess)

	// // Upload the file to S3.
	// result, err := uploader.Upload(&s3manager.UploadInput{
	// 	Bucket: aws.String(bucketName),
	// 	Key:    aws.String("myString"),
	// 	Body:   nil,
	// })

	// if err != nil {
	// 	return c.Status(fiber.StatusConflict).JSON(fiber.Map{
	// 		"errorMessage": "the file cannot be upload",
	// 	})
	// }

	// location := result.Location
	// return c.SendString(fmt.Sprintf("Post :=> %s", location))
}

// DeleteDocument godoc
// @Description  delete document associated with current user - pulled from token
// @Tags           documents
// @Produce      json
// @Accept       json
// @Success      204
// @Security     BearerAuth
// @Router       /documents/{id} [get]
func (udc *DocumentsController) DeleteDocument(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

type DocumentResponse struct {
	Name string
	URL  string
}

type DocumentRequest struct {
	Name string           `json:"name"`
	Type DocumentTypeEnum `json:"type"`
}

// AutoPiSubStatusEnum integration sub-status
type DocumentTypeEnum string

const (
	DriversLicense DocumentTypeEnum = "DriversLicense"
)

func (r DocumentTypeEnum) String() string {
	return string(r)
}
