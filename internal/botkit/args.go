package botkit

import "encoding/json"

func ParseJSON[T any](src string) (T, error) { // Функция для парсинга json объектов
	var args T

	if err := json.Unmarshal([]byte(src), &args); err != nil { // Парсим json в дженерик тип args
		return *(new(T)), err //*(new(T)) нил указатель на тип Т
	}

	return args, nil
}
