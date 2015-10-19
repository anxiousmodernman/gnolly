package models

import "errors"

// Item is a basic unit of knowledge
type Item struct {
	Key   string
	Value string
}

// NewItem is the basic factory function for Item types
func NewItem(key string, value string) (Item, error) {
	if key == "" {
		return Item{}, EmptyKey
	}

	if value == "" {
		return Item{}, EmptyValue
	}

	return Item{key, value}, nil
}

var EmptyKey = errors.New("Empty string")
var EmptyValue = errors.New("Empty string")
