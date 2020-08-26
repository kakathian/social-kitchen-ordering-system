package utility

import (
	"encoding/json"
	"io/ioutil"
)

func ReadFile(name string, content interface{}) error {
	jsonFile, err := ioutil.ReadFile(name)

	if err != nil {
		return err
	}

	json.Unmarshal(jsonFile, &content)
	return nil
}