package codec

import "encoding/binary"

type pcmF32Encoder struct{}

func (e *pcmF32Encoder) Encode(sound []float32) ([]byte, error) {
	return binary.Append(nil, order, sound)
}

type pcmF32Decoder struct{}

func (e *pcmF32Decoder) Decode(buf []byte) ([]float32, error) {
	data := make([]float32, len(buf)/4)
	_, err := binary.Decode(buf, order, data)
	return data, err
}
