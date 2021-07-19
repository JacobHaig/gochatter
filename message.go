package main

import (
	"encoding/json"
	"fmt"
)

type Message struct {
	Username string
	Content  string
	//Time time.Time
}

func (m Message) ToString() string {
	return fmt.Sprintf("%s : %s", m.Username, m.Content)
}

func (m Message) ToBytes() *[]byte {
	bytes, err := json.Marshal(m)
	if err != nil {
		println(err)
	}

	return &bytes
}

func FromBytes(b []byte) *Message {
	var message Message
	err := json.Unmarshal(b[:], &message)
	if err != nil {
		println(err)
	}

	return &message
}
