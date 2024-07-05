package mavlink

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"reflect"
)

// Encode and send mavlink message to the vehicle
func SendMessage(message DecodedMessage, conn net.Conn) error {
	var err error
	var encodedMsg []byte
	encodedMsg, err = encodeMessage(message)
	fmt.Println("Encoded message: ", encodedMsg)
	if err != nil {
		fmt.Println("Error encoding message: ", err)
		return err
	}
	_, err = conn.Write(encodedMsg)
	if err != nil {
		fmt.Println("Error sending message: ", err)
		return err
	}
	return nil
}

func encodeMessage(message DecodedMessage) ([]byte, error) {
	// TODO - Implement the encoding of the message
	fmt.Println("Encoding message: ", message.GetMessageName())

	var buffer bytes.Buffer

	header := []byte{
		0,                            // length
		0,                            // sequence
		0,                            // system id
		0,                            // component id
		byte(message.GetMessageID()), // message id
	}
	buffer.Write(header)

	// encode payload
	payloadEncoded, err := encodePayload(message.MessageData())
	if err != nil {
		return nil, fmt.Errorf("error encoding payload: %v", err)
	}

	buffer.Bytes()[0] = byte(len(payloadEncoded)) // update length

	// TODO: calculate crc
	// crc := byte(0)
	buffer.Write(payloadEncoded)

	// example of a MISSION_ITEM_INT message:
	// https://mavlink.io/en/messages/common.html#MISSION_ITEM_INT
	return buffer.Bytes(), nil
}

func encodePayload(payload DecodedPayload) ([]byte, error) {
	var buffer bytes.Buffer

	for key, value := range payload {
		// Encode the key as a string
		if err := binary.Write(&buffer, binary.LittleEndian, int32(len(key))); err != nil {
			return nil, fmt.Errorf("failed to encode key length: %v", err)
		}
		buffer.WriteString(key)

		// Handle the value based on its type
		val := reflect.ValueOf(value)
		switch val.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if err := binary.Write(&buffer, binary.LittleEndian, val.Int()); err != nil {
				return nil, fmt.Errorf("failed to encode int: %v", err)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if err := binary.Write(&buffer, binary.LittleEndian, val.Uint()); err != nil {
				return nil, fmt.Errorf("failed to encode uint: %v", err)
			}
		case reflect.Float32, reflect.Float64:
			if err := binary.Write(&buffer, binary.LittleEndian, val.Float()); err != nil {
				return nil, fmt.Errorf("failed to encode float: %v", err)
			}
		case reflect.String:
			str := val.String()
			if err := binary.Write(&buffer, binary.LittleEndian, int32(len(str))); err != nil {
				return nil, fmt.Errorf("failed to encode string length: %v", err)
			}
			buffer.WriteString(str)
		default:
			return nil, fmt.Errorf("unsupported value type: %v", val.Kind())
		}
	}

	return buffer.Bytes(), nil
}
