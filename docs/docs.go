// Package docs GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag
package docs

import (
	"bytes"
	"encoding/json"
	"strings"
	"text/template"

	"github.com/swaggo/swag"
)

var doc = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/": {
            "get": {
                "description": "get the status of server.",
                "consumes": [
                    "*/*"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "root"
                ],
                "summary": "Show the status of server.",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    }
                }
            }
        },
        "/device-definitions": {
            "get": {
                "description": "gets a specific device definition by make model and year",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "device-definitions"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "description": "make eg TESLA",
                        "name": "make",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "model eg MODEL Y",
                        "name": "model",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "year eg 2021",
                        "name": "year",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/services.DeviceDefinition"
                        }
                    }
                }
            }
        },
        "/device-definitions/all": {
            "get": {
                "description": "returns a json tree of Makes, models, and years",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "device-definitions"
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/controllers.DeviceMMYRoot"
                            }
                        }
                    }
                }
            }
        },
        "/device-definitions/{id}": {
            "get": {
                "description": "gets a specific device definition by id",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "device-definitions"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "description": "device definition id, KSUID format",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/services.DeviceDefinition"
                        }
                    }
                }
            }
        },
        "/device-definitions/{id}/integrations": {
            "get": {
                "description": "gets all the available integrations for a device definition. Includes the capabilities of the device with the integration",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "device-definitions"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "description": "device definition id, KSUID format",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/services.DeviceCompatibility"
                            }
                        }
                    }
                }
            }
        },
        "/user/devices": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    },
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "adds a device to a user. can add with only device_definition_id or with MMY, which will create a device_definition on the fly",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user-devices"
                ],
                "parameters": [
                    {
                        "description": "add device to user. either MMY or id are required",
                        "name": "user_device",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/controllers.RegisterUserDevice"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/controllers.RegisterUserDeviceResponse"
                        }
                    }
                }
            }
        },
        "/user/devices/:userDeviceID": {
            "delete": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "delete the user device record (hard delete)",
                "tags": [
                    "user-devices"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "description": "user id",
                        "name": "userDeviceID",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": ""
                    }
                }
            }
        },
        "/user/devices/:userDeviceID/commands/refresh": {
            "post": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Starts the process of refreshing device status from Smartcar",
                "tags": [
                    "user-devices"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "description": "user device ID",
                        "name": "user_device_id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": ""
                    },
                    "429": {
                        "description": "rate limit hit for integration"
                    }
                }
            }
        },
        "/user/devices/:userDeviceID/country_code": {
            "patch": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "updates the CountryCode on the user device record",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user-devices"
                ],
                "parameters": [
                    {
                        "description": "Country code",
                        "name": "name",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/controllers.UpdateCountryCodeReq"
                        }
                    }
                ],
                "responses": {
                    "204": {
                        "description": ""
                    }
                }
            }
        },
        "/user/devices/:userDeviceID/integrations/:integrationID": {
            "get": {
                "description": "Receive status updates about a Smartcar integration",
                "tags": [
                    "user-devices"
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/controllers.GetUserDeviceIntegrationResponse"
                        }
                    }
                }
            },
            "post": {
                "description": "Use a Smartcar auth code to connect to Smartcar and obtain access and refresh\ntokens for use by the app.",
                "consumes": [
                    "application/json"
                ],
                "tags": [
                    "user-devices"
                ],
                "parameters": [
                    {
                        "description": "Authorization code from Smartcar",
                        "name": "userDeviceIntegrationRegistration",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/controllers.RegisterSmartcarRequest"
                        }
                    }
                ],
                "responses": {
                    "204": {
                        "description": ""
                    }
                }
            },
            "delete": {
                "description": "Remove an user device's integration",
                "tags": [
                    "user-devices"
                ],
                "responses": {
                    "204": {
                        "description": ""
                    }
                }
            }
        },
        "/user/devices/:userDeviceID/name": {
            "patch": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "updates the Name on the user device record",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user-devices"
                ],
                "parameters": [
                    {
                        "description": "Name",
                        "name": "name",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/controllers.UpdateNameReq"
                        }
                    },
                    {
                        "type": "string",
                        "description": "user id",
                        "name": "user_device_id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": ""
                    }
                }
            }
        },
        "/user/devices/:userDeviceID/status": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Returns the latest status update for the device. May return 404 if the\nuser does not have a device with the ID, or if no status updates have come",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user-devices"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "description": "user device ID",
                        "name": "user_device_id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": ""
                    }
                }
            }
        },
        "/user/devices/:userDeviceID/vin": {
            "patch": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "updates the VIN on the user device record",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user-devices"
                ],
                "parameters": [
                    {
                        "description": "VIN",
                        "name": "vin",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/controllers.UpdateVINReq"
                        }
                    },
                    {
                        "type": "string",
                        "description": "user id",
                        "name": "userDeviceID",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": ""
                    }
                }
            }
        },
        "/user/devices/me": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "gets all devices associated with current user - pulled from token",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user-devices"
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/controllers.UserDeviceFull"
                            }
                        }
                    }
                }
            }
        },
        "/user/geofences": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    },
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "gets all geofences for the current user",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "geofence"
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/controllers.GetGeofence"
                            }
                        }
                    }
                }
            },
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    },
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "adds a new geofence to the user's account, optionally attached to specific user_devices",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "geofence"
                ],
                "parameters": [
                    {
                        "description": "add geofence to user.",
                        "name": "geofence",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/controllers.CreateGeofence"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/controllers.CreateResponse"
                        }
                    }
                }
            }
        },
        "/user/geofences/:geofenceID": {
            "put": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    },
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "updates an existing geofence for the current user",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "geofence"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "description": "geofence id",
                        "name": "geofenceID",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "add geofence to user.",
                        "name": "geofence",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/controllers.CreateGeofence"
                        }
                    }
                ],
                "responses": {
                    "204": {
                        "description": ""
                    }
                }
            },
            "delete": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    },
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "hard deletes a geofence from db",
                "tags": [
                    "geofence"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "description": "geofence id",
                        "name": "geofenceID",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": ""
                    }
                }
            }
        }
    },
    "definitions": {
        "controllers.CreateGeofence": {
            "type": "object",
            "properties": {
                "h3Indexes": {
                    "description": "required: true",
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "name": {
                    "description": "required: true",
                    "type": "string"
                },
                "type": {
                    "description": "one of following: \"PrivacyFence\", \"TriggerEntry\", \"TriggerExit\"\nrequired: true",
                    "type": "string"
                },
                "userDeviceIds": {
                    "description": "Optionally link the geofence with a list of user device Id",
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                }
            }
        },
        "controllers.CreateResponse": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "string"
                }
            }
        },
        "controllers.DeviceMMYRoot": {
            "type": "object",
            "properties": {
                "make": {
                    "type": "string"
                },
                "models": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/controllers.DeviceModels"
                    }
                }
            }
        },
        "controllers.DeviceModelYear": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "string"
                },
                "year": {
                    "type": "integer"
                }
            }
        },
        "controllers.DeviceModels": {
            "type": "object",
            "properties": {
                "model": {
                    "type": "string"
                },
                "years": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/controllers.DeviceModelYear"
                    }
                }
            }
        },
        "controllers.GeoFenceUserDevice": {
            "type": "object",
            "properties": {
                "mmy": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "userDeviceId": {
                    "type": "string"
                }
            }
        },
        "controllers.GetGeofence": {
            "type": "object",
            "properties": {
                "h3Indexes": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "id": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "type": {
                    "type": "string"
                },
                "userDevices": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/controllers.GeoFenceUserDevice"
                    }
                }
            }
        },
        "controllers.GetUserDeviceIntegrationResponse": {
            "type": "object",
            "properties": {
                "externalId": {
                    "description": "ExternalID is the identifier used by the third party for the device. It may be absent if we\nhaven't authorized yet.",
                    "type": "string"
                },
                "status": {
                    "description": "Status is one of \"Pending\", \"PendingFirstData\", \"Active\"",
                    "type": "string"
                }
            }
        },
        "controllers.RegisterSmartcarRequest": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "string"
                },
                "redirectURI": {
                    "type": "string"
                }
            }
        },
        "controllers.RegisterUserDevice": {
            "type": "object",
            "properties": {
                "countryCode": {
                    "type": "string"
                },
                "deviceDefinitionId": {
                    "type": "string"
                },
                "make": {
                    "type": "string"
                },
                "model": {
                    "type": "string"
                },
                "year": {
                    "type": "integer"
                }
            }
        },
        "controllers.RegisterUserDeviceResponse": {
            "type": "object",
            "properties": {
                "deviceDefinitionId": {
                    "type": "string"
                },
                "integrationCapabilities": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/services.DeviceCompatibility"
                    }
                },
                "userDeviceId": {
                    "type": "string"
                }
            }
        },
        "controllers.UpdateCountryCodeReq": {
            "type": "object",
            "properties": {
                "countryCode": {
                    "type": "string"
                }
            }
        },
        "controllers.UpdateNameReq": {
            "type": "object",
            "properties": {
                "name": {
                    "type": "string"
                }
            }
        },
        "controllers.UpdateVINReq": {
            "type": "object",
            "properties": {
                "vin": {
                    "type": "string"
                }
            }
        },
        "controllers.UserDeviceFull": {
            "type": "object",
            "properties": {
                "countryCode": {
                    "type": "string"
                },
                "customImageUrl": {
                    "type": "string"
                },
                "deviceDefinition": {
                    "$ref": "#/definitions/services.DeviceDefinition"
                },
                "id": {
                    "type": "string"
                },
                "integrations": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/controllers.UserDeviceIntegrationStatus"
                    }
                },
                "name": {
                    "type": "string"
                },
                "vin": {
                    "type": "string"
                },
                "vinConfirmed": {
                    "type": "boolean"
                }
            }
        },
        "controllers.UserDeviceIntegrationStatus": {
            "type": "object",
            "properties": {
                "integrationId": {
                    "type": "string"
                },
                "status": {
                    "type": "string"
                }
            }
        },
        "services.DeviceCompatibility": {
            "type": "object",
            "properties": {
                "capabilities": {
                    "type": "string"
                },
                "country": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "style": {
                    "type": "string"
                },
                "type": {
                    "type": "string"
                },
                "vendor": {
                    "type": "string"
                }
            }
        },
        "services.DeviceDefinition": {
            "type": "object",
            "properties": {
                "compatibleIntegrations": {
                    "description": "CompatibleIntegrations has systems this vehicle can integrate with",
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/services.DeviceCompatibility"
                    }
                },
                "deviceDefinitionId": {
                    "type": "string"
                },
                "imageUrl": {
                    "type": "string"
                },
                "metadata": {},
                "name": {
                    "type": "string"
                },
                "type": {
                    "$ref": "#/definitions/services.DeviceType"
                },
                "vehicleData": {
                    "description": "VehicleInfo will be empty if not a vehicle type",
                    "$ref": "#/definitions/services.DeviceVehicleInfo"
                },
                "verified": {
                    "type": "boolean"
                }
            }
        },
        "services.DeviceType": {
            "type": "object",
            "properties": {
                "make": {
                    "type": "string"
                },
                "model": {
                    "type": "string"
                },
                "subModels": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "type": {
                    "description": "Type is eg. Vehicle, E-bike, roomba",
                    "type": "string"
                },
                "year": {
                    "type": "integer"
                }
            }
        },
        "services.DeviceVehicleInfo": {
            "type": "object",
            "properties": {
                "base_msrp": {
                    "type": "integer"
                },
                "driven_wheels": {
                    "type": "string"
                },
                "epa_class": {
                    "type": "string"
                },
                "fuel_type": {
                    "type": "string"
                },
                "mpg_city": {
                    "type": "string"
                },
                "mpg_highway": {
                    "type": "string"
                },
                "number_of_doors": {
                    "type": "string"
                },
                "vehicle_type": {
                    "description": "VehicleType PASSENGER CAR, from NHTSA",
                    "type": "string"
                }
            }
        }
    },
    "securityDefinitions": {
        "BearerAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = swaggerInfo{
	Version:     "2.0",
	Host:        "",
	BasePath:    "/v1",
	Schemes:     []string{},
	Title:       "DIMO Devices API",
	Description: "",
}

type s struct{}

func (s *s) ReadDoc() string {
	sInfo := SwaggerInfo
	sInfo.Description = strings.Replace(sInfo.Description, "\n", "\\n", -1)

	t, err := template.New("swagger_info").Funcs(template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
		"escape": func(v interface{}) string {
			// escape tabs
			str := strings.Replace(v.(string), "\t", "\\t", -1)
			// replace " with \", and if that results in \\", replace that with \\\"
			str = strings.Replace(str, "\"", "\\\"", -1)
			return strings.Replace(str, "\\\\\"", "\\\\\\\"", -1)
		},
	}).Parse(doc)
	if err != nil {
		return doc
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, sInfo); err != nil {
		return doc
	}

	return tpl.String()
}

func init() {
	swag.Register("swagger", &s{})
}
