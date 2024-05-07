package botkit

import "encoding/json"

func ParseJSON[T any](src string) (T, error) {
	var args T

	err := json.Unmarshal([]byte(src), &args)
	if err != nil {
		return *(new(T)), err
	}

	return args, nil
}
