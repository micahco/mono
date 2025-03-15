package uuid

import "github.com/gofrs/uuid/v5"

type UUID struct {
	uuid.UUID
}

var Nil = UUID{uuid.Nil}

func FromString(text string) (UUID, error) {
	u, err := uuid.FromString(text)
	if err != nil {
		return UUID{uuid.Nil}, nil
	}

	return UUID{u}, nil
}

func FromStringOrNil(input string) UUID {
	uuid, err := FromString(input)
	if err != nil {
		return Nil
	}
	return uuid
}
