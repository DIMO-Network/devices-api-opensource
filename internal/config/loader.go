package config

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
)

// LoadConfig fills in all the values in the Settings from local yml file (for dev) and env vars (for deployments)
func LoadConfig(filePath string) (*Settings, error) {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "could not read file: " + filePath)
	}
	settings, err := loadFromYaml(b)
	loadFromEnvVars(settings) // override with any env vars found

	return settings, nil
}

func loadFromYaml(yamlFile []byte) (*Settings, error) {
	var settings Settings
	err := yaml.Unmarshal(yamlFile, &settings)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshall yaml file to settings")
	}
	return &settings, nil
}

func loadFromEnvVars(settings *Settings) {
	valueOfConfig := reflect.ValueOf(settings).Elem()
	typeOfT := valueOfConfig.Type()

	// iterate over all struct fields
	for i := 0; i < valueOfConfig.NumField(); i++ {
		field := valueOfConfig.Field(i)
		fieldYamlName := typeOfT.Field(i).Tag.Get("yaml")

		// check if env var with same field yaml name exists, if so, set the value from the env var
		if env, exists := os.LookupEnv(fieldYamlName); exists {
			var val interface{}
			switch field.Kind() {
			case reflect.String:
				val = env
			case reflect.Bool:
				val, _ = strconv.ParseBool(env)
			case reflect.Int:
				val, _ = strconv.Atoi(env)
			case reflect.Int64:
				val, _ = strconv.ParseInt(env, 10,64)
			}
			// now set the field with the val
			if val != nil {
				field.Set(reflect.ValueOf(val))
			}
		}
	}
}
