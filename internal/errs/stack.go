package errs

import "runtime"

func stack(skip int) []Frame {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(skip, pcs[:])
	if n == 0 {
		return nil
	}

	frames := runtime.CallersFrames(pcs[:n])
	out := make([]Frame, 0, n)
	for {
		frame, more := frames.Next()
		out = append(out, Frame{
			Function: frame.Function,
			File:     frame.File,
			Line:     frame.Line,
		})
		if !more {
			break
		}
	}
	return out
}
