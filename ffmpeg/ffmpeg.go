package ffmpeg

// #include <libavcodec/avcodec.h>
// #include <libswscale/swscale.h>
//
// // ... yes. Don't ask.
// typedef struct SwsContext SwsContext;
//
// #ifndef PIX_FMT_RGB0
// #define PIX_FMT_RGB0 PIX_FMT_RGB32
// #endif
//
// #cgo pkg-config: libavdevice libavformat libavfilter libavcodec libswscale libavutil
import "C"

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"unsafe"
)

const (
	CODEC_ID_H264 = C.AV_CODEC_ID_H264
)

type Encoder struct {
	//codec uint32
	//im            image.Image
	//underlying_im image.Image
	Output io.Writer

	codec   *C.AVCodec
	context *C.AVCodecContext
	//_swscontext *C.SwsContext
	frame  *C.AVFrame
	pkt    *C.AVPacket
	outbuf []byte
}

func init() {
	C.avcodec_register_all()
}

func ptr(buf []byte) *C.uint8_t {
	h := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	return (*C.uint8_t)(unsafe.Pointer(h.Data))
}

/*
type EncoderOptions struct {
    BitRate uint32
    W, H int
    TimeBase
} */

/*
var DefaultEncoderOptions = EncoderOptions{
    BitRate:400000,
    W: 0, H: 0,
    c.time_base = C.AVRational{1,25}
    c.gop_size = 10
    c.max_b_frames = 1
    c.pix_fmt = C.PIX_FMT_RGB
} */

func NewEncoder(fileName string) (*Encoder, error) {
	f, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}

	codec := C.avcodec_find_encoder(C.AV_CODEC_ID_H264)
	if codec == nil {
		return nil, fmt.Errorf("could not find codec")
	}

	c := C.avcodec_alloc_context3(codec)
	if c == nil {
		return nil, fmt.Errorf("сould not allocate video codec context")
	}

	pkt := C.av_packet_alloc()
	if pkt == nil {
		return nil, fmt.Errorf("сould not allocate packet")
	}

	/* put sample parameters */
	c.bit_rate = 400000
	/* resolution must be a multiple of two */
	c.width = 352
	c.height = 288
	/* frames per second */
	c.time_base = C.AVRational{1, 25}
	c.framerate = C.AVRational{25, 1}

	/* emit one intra frame every ten frames
	 * check frame pict_type before passing frame
	 * to encoder, if frame->pict_type is AV_PICTURE_TYPE_I
	 * then gop_size is ignored and the output of encoder
	 * will always be I frame irrespective to gop_size
	 */
	c.gop_size = 10
	c.max_b_frames = 1
	c.pix_fmt = C.AV_PIX_FMT_YUV420P

	//if (codec->id == AV_CODEC_ID_H264)
	//	av_opt_set(c->priv_data, "preset", "slow", 0);

	/* open it */
	ret := C.avcodec_open2(c, codec, nil)
	if ret < 0 {
		return nil, fmt.Errorf("Could not open codec: %v\n", ret)
	}

	frame := C.av_frame_alloc()
	if frame == nil {
		return nil, fmt.Errorf("could not allocate video frame")
	}
	frame.format = C.AV_PIX_FMT_YUV420P
	frame.width = c.width
	frame.height = c.height

	ret = C.av_frame_get_buffer(frame, 0)
	if ret < 0 {
		return nil, fmt.Errorf("could not allocate the video frame data")
	}

	e := &Encoder{f, codec, c, frame, pkt, make([]byte, 16*1024)}
	return e, nil
}

func (e *Encoder) CreateFrame(pts int) {
	C.fflush(C.stdout)

	ret := C.av_frame_make_writable(e.frame)
	if ret < 0 {
		fmt.Printf("ret less than zero: %v", ret)
	}

	for y := 0; y < int(e.context.height); y++ {
		for x := 0; x < int(e.context.width); x++ {
			e.frame.data[0][y*e.frame.linesize[0]+x] = x + y + pts*3
		}
	}

	for y := 0; y < int(e.context.height/2); y++ {
		for x := 0; x < int(e.context.width/2); x++ {
			e.frame.data[1][y*e.frame.linesize[1]+x] = 128 + y + pts*2
			e.frame.data[2][y*e.frame.linesize[2]+x] = 64 + x + pts*5
		}
	}

	e.frame.pts = pts
}

func (e *Encoder) WriteLastFrame() {
	e.frame = nil
}

//func NewEncoder(codec uint32, in image.Image, out io.Writer) (*Encoder, error) {
//	codec := C.avcodec_find_encoder(codec)
//	if codec == nil {
//		return nil, fmt.Errorf("could not find codec")
//	}
//
//	c := C.avcodec_alloc_context3(codec)
//	f := C.av_frame_alloc()
//	pkt := C.av_packet_alloc()
//
//	c.bit_rate = 400000
//
//	// resolution must be a multiple of two
//	w, h := C.int(in.Bounds().Dx()), C.int(in.Bounds().Dy())
//	if w%2 == 1 || h%2 == 1 {
//		return nil, fmt.Errorf("Bad image dimensions (%d, %d), must be even", w, h)
//	}
//
//	log.Printf("Encoder dimensions: %d, %d", w, h)
//
//	c.width = w
//	c.height = h
//	c.time_base = C.AVRational{1, 25} // FPS
//	c.gop_size = 10                   // emit one intra frame every ten frames
//	c.max_b_frames = 1
//
//	underlying_im := image.NewYCbCr(in.Bounds(), image.YCbCrSubsampleRatio420)
//	c.pix_fmt = C.AV_PIX_FMT_YUV420P
//	f.data[0] = ptr(underlying_im.Y)
//	f.data[1] = ptr(underlying_im.Cb)
//	f.data[2] = ptr(underlying_im.Cr)
//	f.linesize[0] = w
//	f.linesize[1] = w / 2
//	f.linesize[2] = w / 2
//
//	if C.avcodec_open2(c, codec, nil) < 0 {
//		return nil, fmt.Errorf("could not open codec")
//	}
//
//	_swscontext := C.sws_getContext(w, h, C.AV_PIX_FMT_RGB0, w, h, C.AV_PIX_FMT_YUV420P,
//		C.SWS_BICUBIC, nil, nil, nil)
//
//	e := &Encoder{codec, in, underlying_im, out, codec, c, _swscontext, f, pkt, make([]byte, 16*1024)}
//	return e, nil
//}

//func (e *Encoder) WriteFrame() error {
//	e._frame.pts = C.int64_t(e._context.frame_number)
//
//	var input_data [3]*C.uint8_t
//	var input_linesize [3]C.int
//
//	switch im := e.im.(type) {
//	case *image.RGBA:
//		bpp := 4
//		input_data = [3]*C.uint8_t{ptr(im.Pix)}
//		input_linesize = [3]C.int{C.int(e.im.Bounds().Dx() * bpp)}
//	case *image.NRGBA:
//		bpp := 4
//		input_data = [3]*C.uint8_t{ptr(im.Pix)}
//		input_linesize = [3]C.int{C.int(e.im.Bounds().Dx() * bpp)}
//	default:
//		panic("Unknown input image type")
//	}
//
//	// Perform scaling from input type to output type
//	C.sws_scale(e._swscontext, &input_data[0], &input_linesize[0],
//		0, e._context.height,
//		&e._frame.data[0], &e._frame.linesize[0])
//
//	outsize := C.avcodec_encode_video2(e._context, ptr(e._outbuf),
//		C.int(len(e._outbuf)), e._frame)
//
//	if outsize == 0 {
//		return nil
//	}
//
//	n, err := e.Output.Write(e._outbuf[:outsize])
//	if err != nil {
//		return err
//	}
//	if n < int(outsize) {
//		return fmt.Errorf("Short write, expected %d, wrote %d", outsize, n)
//	}
//
//	return nil
//}

func (e *Encoder) Encode() {
	/* send the frame to the encoder */
	if e.frame != nil {
		fmt.Printf("Send frame %v", e.frame.pts)
	}

	ret := C.avcodec_send_frame(e.context, e.frame)
	if ret < 0 {
		fmt.Println("Error sending a frame for encoding")
	}

	for ret >= 0 {
		ret = C.avcodec_receive_packet(e.context, e.pkt)
		if ret == C.EAGAIN || ret == C.AVERROR_EOF {
			return
		}
		if ret < 0 {
			fmt.Println("error during encoding")
		}

		fmt.Printf("Write packet %v, %v", e.pkt.pts, e.pkt.size)
		C.fwrite(e.pkt.data, 1, e.pkt.size, e.Output)
		C.av_packet_unref(e.pkt)
	}
}

func (e *Encoder) Close() {
	C.fclose(e.Output)

	C.avcodec_free_context(e.context)
	C.av_frame_free(e.frame)
	C.av_packet_free(e.pkt)
}

//func (e *Encoder) Encode() error {
//	e._frame.pts = C.int64_t(e._context.frame_number)
//
//	//var input_data [3]*C.uint8_t
//	//var input_linesize [3]C.int
//	//
//	//switch im := e.im.(type) {
//	//case *image.RGBA:
//	//	bpp := 4
//	//	input_data = [3]*C.uint8_t{ptr(im.Pix)}
//	//	input_linesize = [3]C.int{C.int(e.im.Bounds().Dx() * bpp)}
//	//case *image.NRGBA:
//	//	bpp := 4
//	//	input_data = [3]*C.uint8_t{ptr(im.Pix)}
//	//	input_linesize = [3]C.int{C.int(e.im.Bounds().Dx() * bpp)}
//	//default:
//	//	panic("Unknown input image type")
//	//}
//	//
//	//// Perform scaling from input type to output type
//	//C.sws_scale(e._swscontext, &input_data[0], &input_linesize[0],
//	//	0, e._context.height,
//	//	&e._frame.data[0], &e._frame.linesize[0])
//
//	var ret C.int
//
//	ret = C.avcodec_send_frame(e._context, e._frame)
//	if ret < 0 {
//		println("Error sending a frame for encoding")
//	}
//
//	for ret >= 0 {
//		ret = C.avcodec_receive_packet(e._context, e._pkt)
//		if ret == C.EAGAIN || ret == C.AVERROR_EOF {
//			return nil
//		} else if ret < 0 {
//			println("Error during encoding")
//		}
//
//		//printf("Write packet %3"PRId64" (size=%5d)\n", pkt->pts, pkt->size);
//		//fwrite(pkt->data, 1, pkt->size, outfile);
//
//		//_, err := e.Output.Write(e._pkt.data[:e._pkt.size])
//		//if err != nil {
//		//	return err
//		//}
//		//C.av_packet_unref(e._pkt)
//	}
//	return nil
//}

//func (e *Encoder) Close() {
//
//	// Process "delayed" frames
//	for {
//		outsize := C.avcodec_encode_video2(e._context, ptr(e._outbuf),
//			C.int(len(e._outbuf)), nil)
//
//		if outsize == 0 {
//			break
//		}
//
//		n, err := e.Output.Write(e._outbuf[:outsize])
//		if err != nil {
//			panic(err)
//		}
//		if n < int(outsize) {
//			panic(fmt.Errorf("Short write, expected %d, wrote %d", outsize, n))
//		}
//	}
//
//	n, err := e.Output.Write([]byte{0, 0, 1, 0xb7})
//	if err != nil || n != 4 {
//		log.Panicf("Error finishing mpeg file: %q; n = %d", err, n)
//	}
//
//	C.avcodec_close((*C.AVCodecContext)(unsafe.Pointer(e._context)))
//	C.av_free(unsafe.Pointer(e._context))
//	C.av_free(unsafe.Pointer(e._frame))
//	e._frame, e.codec = nil, nil
//}
