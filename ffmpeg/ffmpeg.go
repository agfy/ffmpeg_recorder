package ffmpeg

// #cgo CFLAGS: -I/home/max/Downloads/ffmpeg-6.0
// #cgo LDFLAGS: -L/home/max/Downloads/ffmpeg-6.0 -lavcodec -lavformat -lavutil -lswscale
// #cgo pkg-config: libavformat libavcodec libavutil libswscale
// #include <libavformat/avformat.h>
// #include <libavcodec/avcodec.h>
// #include <libavutil/imgutils.h>
// #include <libswscale/swscale.h>
// // Function to use the AVERROR macro
// int my_averror(int errnum) {
//   return AVERROR(errnum);
// }
import "C"
import (
	"fmt"
	"image"
	"image/color"
	"unsafe"
)

const (
	width  = 640
	height = 514
	fps    = 25
)

type Encoder struct {
	file     unsafe.Pointer
	frameNum int

	codec   *C.AVCodec
	context *C.AVCodecContext
	frame   *C.AVFrame
	packet  *C.AVPacket
}

func rgbaToYuv(rgba color.RGBA) (y, u, v uint8) {
	r := float64(rgba.R)
	g := float64(rgba.G)
	b := float64(rgba.B)

	y = uint8((0.257 * r) + (0.504 * g) + (0.098 * b) + 16)
	u = uint8(-(0.148 * r) - (0.291 * g) + (0.439 * b) + 128)
	v = uint8((0.439 * r) - (0.368 * g) - (0.071 * b) + 128)

	return y, u, v
}

func NewEncoder(fileName string) (*Encoder, error) {
	avcodec := C.avcodec_find_encoder(C.AV_CODEC_ID_H264)
	if avcodec == nil {
		panic("Encoder not found.")
	}

	avctx := C.avcodec_alloc_context3(avcodec)
	if avctx == nil {
		panic("Cannot allocate codec context.")
	}

	avctx.width = C.int(width)
	avctx.height = C.int(height)
	avctx.time_base = C.struct_AVRational{1, fps}
	avctx.pix_fmt = C.AV_PIX_FMT_YUV420P

	if C.avcodec_open2(avctx, avcodec, nil) < 0 {
		panic("Cannot open codec.")
	}

	outfile, err := C.fopen(C.CString(fileName), C.CString("wb"))
	if err != nil {
		panic("Cannot open output file.")
	}

	frame := C.av_frame_alloc()
	if frame == nil {
		return nil, fmt.Errorf("could not allocate video frame")
	}
	frame.format = C.AV_PIX_FMT_YUV420P
	frame.width = width
	frame.height = height

	ret := C.av_frame_get_buffer(frame, 0)
	if ret < 0 {
		return nil, fmt.Errorf("could not allocate the video frame data")
	}

	packet := C.av_packet_alloc()
	if packet == nil {
		return nil, fmt.Errorf("cannot allocate packet")
	}

	e := &Encoder{
		file:    unsafe.Pointer(outfile),
		codec:   avcodec,
		context: avctx,
		frame:   frame,
		packet:  packet,
	}

	return e, nil
}

func (e *Encoder) Encode(img image.Image) error {
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			yy, uu, vv := rgbaToYuv(color.RGBA{uint8(r), uint8(g), uint8(b), 255})

			// Set pixel at (x, y) to the corresponding color in Y plane.
			pointer := unsafe.Pointer(uintptr(unsafe.Pointer(e.frame.data[0])) + uintptr(y)*uintptr(e.frame.linesize[0]) + uintptr(x))
			*(*C.uint8_t)(pointer) = C.uint8_t(yy)

			// The U and V planes are half the size of the Y plane in YUV420P.
			if y%2 == 0 && x%2 == 0 {
				// Set pixel at (x/2, y/2) to the corresponding color in U and V planes.
				pointer = unsafe.Pointer(uintptr(unsafe.Pointer(e.frame.data[1])) + uintptr(y/2)*uintptr(e.frame.linesize[1]) + uintptr(x/2))
				*(*C.uint8_t)(pointer) = C.uint8_t(uu)
				pointer = unsafe.Pointer(uintptr(unsafe.Pointer(e.frame.data[2])) + uintptr(y/2)*uintptr(e.frame.linesize[2]) + uintptr(x/2))
				*(*C.uint8_t)(pointer) = C.uint8_t(vv)
			}
		}
	}

	e.frame.pts = C.int64_t(e.frameNum)
	e.frameNum++

	ret := C.avcodec_send_frame(e.context, e.frame)
	if ret < 0 {
		return fmt.Errorf("avcodec_send_frame returned: %v", ret)
	}

	for {
		ret = C.avcodec_receive_packet(e.context, e.packet)
		if ret == C.my_averror(C.EAGAIN) || ret == C.my_averror(C.EOF) {
			break
		} else if ret < 0 {
			return fmt.Errorf("avcodec_receive_packet returned: %v", ret)
		}

		C.fwrite(unsafe.Pointer(e.packet.data), 1, C.size_t(e.packet.size), (*C.FILE)(e.file))

		C.av_packet_unref(e.packet)
	}

	return nil
}

func (e *Encoder) Close() {
	C.fclose((*C.FILE)(e.file))
	C.av_frame_free(&e.frame)
	C.av_packet_free(&e.packet)
	C.avcodec_free_context(&e.context)
}
