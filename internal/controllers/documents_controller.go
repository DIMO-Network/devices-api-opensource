package controllers

import (
	"fmt"
	"path/filepath"

	"bytes"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
)

type DocumentsController struct {
	settings *config.Settings
	s3Client *s3.Client
}

// NewDocumentsController constructor
func NewDocumentsController(settings *config.Settings, s3Client *s3.Client) DocumentsController {
	return DocumentsController{settings: settings, s3Client: s3Client}
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
	userID := getUserID(c)

	response, err := udc.s3Client.ListObjectsV2(c.Context(), &s3.ListObjectsV2Input{
		Bucket: aws.String(udc.settings.AWSDocumentsBucketName),
		Prefix: aws.String(userID),
	})

	documents := []DocumentResponse{}

	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "the bucket does not exist")
	}

	for _, item := range response.Contents {
		responseItem, err := udc.s3Client.GetObject(c.Context(),
			&s3.GetObjectInput{
				Bucket: aws.String(udc.settings.AWSDocumentsBucketName),
				Key:    aws.String(*item.Key),
			})

		if err != nil {
			return errorResponseHandler(c, err, fiber.StatusInternalServerError)
		}

		documents = append(documents, DocumentResponse{
			ID:   responseItem.Metadata[MetadataDocumentID],
			Name: responseItem.Metadata[MetadataDocumentName],
			URL:  fmt.Sprintf("%s/v1/documents/%s/download", udc.settings.DeploymentBaseURL, responseItem.Metadata[MetadataDocumentID]),
			Type: DocumentTypeEnum(responseItem.Metadata[MetadataDocumentType]),
		})

	}

	return c.JSON(documents)
}

// GetDocumentByID godoc
// @Description  get document by id associated with current user - pulled from token
// @Tags           documents
// @Produce      json
// @Accept       json
// @Param        id   path      string  true  "Document ID"
// @Success      200  {object}  controllers.DocumentResponse
// @Security     BearerAuth
// @Router       /documents/{id} [get]
func (udc *DocumentsController) GetDocumentByID(c *fiber.Ctx) error {
	userID := getUserID(c)
	fileID := c.Params("id")

	response, err := udc.s3Client.GetObject(c.Context(),
		&s3.GetObjectInput{
			Bucket: aws.String(udc.settings.AWSDocumentsBucketName),
			Key:    aws.String(fmt.Sprintf("%s/%s", userID, fileID)),
		})

	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("no document with id %s found", fileID))
	}

	document := DocumentResponse{
		ID:   response.Metadata[MetadataDocumentID],
		Name: response.Metadata[MetadataDocumentName],
		URL:  fmt.Sprintf("%s/v1/documents/%s/download", udc.settings.DeploymentBaseURL, response.Metadata[MetadataDocumentID]),
		Type: DocumentTypeEnum(response.Metadata[MetadataDocumentType]),
	}

	return c.JSON(document)
}

// PostDocument godoc
// @Description  post document by id associated with current user - pulled from token
// @Tags           documents
// @Produce      json
// @Accept       multipart/form-data
// @Param        file  formData  file  true  "The file to upload. file is required"
// @Param        name  formData  string  true  "The document name. name is required"
// @Param        type  formData  string  true  "The document type. type is required"
// @Success      201  {object}  controllers.DocumentResponse
// @Security     BearerAuth
// @Router       /documents [post]
func (udc *DocumentsController) PostDocument(c *fiber.Ctx) error {

	userID := getUserID(c)
	file, err := c.FormFile("file")
	documentName := c.FormValue("name")
	documentType := c.FormValue("type")

	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid file.")
	}

	if err := DocumentTypeEnum(documentType).IsValid(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid document type.")
	}
	// Get Buffer from file
	fileHeader, err := file.Open()

	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "document cannot be read.")
	}
	defer fileHeader.Close()

	// Validate file type
	filetype := file.Header.Get("content-type")

	if err := FileTypeAllowedEnum(filetype).IsValid(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "the provided file format is not allowed.")
	}

	// Create an uploader with the session and default options
	uploader := manager.NewUploader(udc.s3Client)

	// Unique Id
	id := ksuid.New().String()

	metadata := map[string]string{}
	metadata[MetadataDocumentID] = id
	metadata[MetadataDocumentName] = documentName
	metadata[MetadataDocumentFile] = file.Filename
	metadata[MetadataDocumentFileExtension] = filepath.Ext(file.Filename)
	metadata[MetadataDocumentType] = documentType

	fileID := fmt.Sprintf("%s%s", id, filepath.Ext(file.Filename))

	// Upload the file to S3.
	result, err := uploader.Upload(c.Context(), &s3.PutObjectInput{
		Bucket:             aws.String(udc.settings.AWSDocumentsBucketName),
		Key:                aws.String(fmt.Sprintf("%s/%s", userID, fileID)),
		Body:               fileHeader,
		ContentDisposition: aws.String("attachment"),
		Metadata:           metadata,
	})

	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	_ = result

	url := fmt.Sprintf("%s/v1/documents/%s/download", udc.settings.DeploymentBaseURL, id)
	return c.JSON(DocumentResponse{ID: id, Name: documentName, URL: url, Type: DocumentTypeEnum(documentType)})
}

// DeleteDocument godoc
// @Description  delete document associated with current user - pulled from token
// @Tags           documents
// @Produce      json
// @Accept       json
// @Param        id   path      string  true  "Document ID"
// @Success      204
// @Security     BearerAuth
// @Router       /documents/{id} [delete]
func (udc *DocumentsController) DeleteDocument(c *fiber.Ctx) error {
	userID := getUserID(c)
	fileID := c.Params("id")

	_, err := udc.s3Client.DeleteObject(c.Context(), &s3.DeleteObjectInput{
		Bucket: aws.String(udc.settings.AWSDocumentsBucketName),
		Key:    aws.String(fmt.Sprintf("%s/%s", userID, fileID)),
	})

	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// DownloadDocument godoc
// @Description  download document associated with current user - pulled from token
// @Tags           documents
// @Produce      octet-stream
// @Produce      png
// @Produce      jpeg
// @Param        id   path      string  true  "Document ID"
// @Success      200
// @Security     BearerAuth
// @Router       /documents/{id}/download [get]
func (udc *DocumentsController) DownloadDocument(c *fiber.Ctx) error {
	userID := getUserID(c)
	fileID := c.Params("id")

	buffer := manager.NewWriteAtBuffer([]byte{})

	downloader := manager.NewDownloader(udc.s3Client)

	numBytes, err := downloader.Download(c.Context(), buffer,
		&s3.GetObjectInput{
			Bucket: aws.String(udc.settings.AWSDocumentsBucketName),
			Key:    aws.String(fmt.Sprintf("%s/%s", userID, fileID)),
		})

	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	if numBytes == 0 {
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("no document with id %s found", fileID))
	}

	data := buffer.Bytes()

	return c.SendStream(bytes.NewReader(data))
}

type DocumentResponse struct {
	ID   string
	Name string
	URL  string
	Type DocumentTypeEnum
}

type FileTypeAllowedEnum string

const (
	jpeg FileTypeAllowedEnum = "image/jpeg"
	png  FileTypeAllowedEnum = "image/png"
	pdf  FileTypeAllowedEnum = "application/pdf"
)

func (r FileTypeAllowedEnum) IsValid() error {
	switch r {
	case jpeg, png, pdf:
		return nil
	}
	return errors.New("invalid document type")
}

type DocumentTypeEnum string

const (
	DriversLicense DocumentTypeEnum = "DriversLicense"
	Other          DocumentTypeEnum = "Other"
)

func (r DocumentTypeEnum) String() string {
	return string(r)
}

func (r DocumentTypeEnum) IsValid() error {
	switch r {
	case DriversLicense, Other:
		return nil
	}
	return errors.New("invalid document type")
}

const (
	MetadataDocumentID            = "document-id"
	MetadataDocumentName          = "document-name"
	MetadataDocumentFile          = "document-file"
	MetadataDocumentType          = "document-type"
	MetadataDocumentFileExtension = "document-file-ext"
)
