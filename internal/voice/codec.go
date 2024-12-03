package voice

import (
	"fmt"

	"github.com/royalcat/konfa/internal/codec"
	voicev1 "github.com/royalcat/konfa/internal/proto/gen/konfa/voice/v1"
)

func mapCodec(c voicev1.Codec) (codec.Codec, error) {
	switch c {
	case voicev1.Codec_CODEC_PCM_F32:
		return codec.CodecPCMf32, nil
	case voicev1.Codec_CODEC_OPUS:
		return codec.CodecOpus, nil
	default:
		return 0, fmt.Errorf("unsupported codec: %v", c)
	}
}
