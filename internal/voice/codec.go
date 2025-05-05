package voice

import (
	"fmt"

	"github.com/konfa-chat/voice-relay/internal/codec"
	voicev1 "github.com/konfa-chat/voice-relay/internal/proto/gen/konfa/voice_relay/v1"
)

func mapCodec(c voicev1.AudioCodec) (codec.Codec, error) {
	switch c {
	case voicev1.AudioCodec_AUDIO_CODEC_PCM_F32:
		return codec.CodecPCMf32, nil
	case voicev1.AudioCodec_AUDIO_CODEC_OPUS:
		return codec.CodecOpus, nil
	default:
		return 0, fmt.Errorf("unsupported codec: %v", c)
	}
}
