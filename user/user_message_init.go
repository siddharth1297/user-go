package user

func AllocateMessage(message_name string) *Message {
	return NewMessage(message_name)
}

func AllocateDeserialiser(message_name string, buf *[]byte, buf_len uint64, buf_start uint64) *Deserialiser {
	return NewDeserialiser(message_name, buf, buf_len, buf_start)
}
