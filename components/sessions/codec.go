package sessions

import (
	"encoding/json"
)

type data struct {
	Entries []dataEntry `json:"entries"`
}

type dataEntry struct {
	Key        string `json:"k"`
	ProtoValue []byte `json:"pv"`
}

func Encode(session *Session) ([]byte, error) {
	var data data
	for k, value := range session.values {
		data.Entries = append(data.Entries, dataEntry{
			Key:        k,
			ProtoValue: value.Data,
		})
	}

	// TODO: Use proto instead
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func Decode(sessionID string, b []byte) (*Session, error) {
	var data data
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, err
	}

	session := &Session{
		values: make(map[string]*sessionValue),
	}
	for _, entry := range data.Entries {
		session.values[entry.Key] = &sessionValue{Data: entry.ProtoValue}
	}
	session.ID = sessionID
	return session, nil
}
