package main

import (
	"fmt"
	"net/url"
)

type urls []url.URL

func (us *urls) String() string {
	return fmt.Sprint(*us)
}

func (us *urls) Set(value string) error {
	u, err := url.Parse(value)
	if err != nil {
		return fmt.Errorf("invalid input URL format: %v", err)
	}

	*us = append(*us, *u)
	return nil
}
