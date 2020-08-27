package pkg

import (
	"encoding/json"
	"io/ioutil"
)

func ReadFile(name string, content interface{}) error {
	jsonFile, err := ioutil.ReadFile(name)

	if err != nil {
		return err
	}

	json.Unmarshal([]byte(jsonFile), &content)
	return nil
}
