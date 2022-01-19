package errors

import (
	"fmt"
	"runtime"
)

var plFrames = locationFrames{
	m:  make(map[uint64]frame),
	ch: make(chan iFrame, 32),
}

type positionError uintptr

//Position 起名字最麻烦了，能定位抛出就好。
func Position() error {
	err := new(positionError)
	err.setLocation(1)
	return err
}

func (p *positionError) Error() string {
	return fmt.Sprintf("caller [%d]", *p)
}

func (p *positionError) Caller() uintptr {
	return uintptr(*p)
}

func (p *positionError) Location() (string, int) {
	return plFrames.get(uint64(*p)).file, plFrames.get(uint64(*p)).line
}

func (p *positionError) Callers() []uintptr { return plFrames.get(uint64(*p)).callers }

func (p *positionError) setLocation(skip int) {
	var f frame
	f.caller, f.file, f.line, _ = runtime.Caller(skip + 1)

	pc := make([]uintptr, 32)
	n := runtime.Callers(skip+2, pc)
	f.callers = pc[0:n]

	*p = positionError(f.caller)
	plFrames.put(uint64(*p), f)

	runtime.SetFinalizer(p, func(p *positionError) {
		plFrames.del(uint64(*p))
	})
}

func (p *positionError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case s.Flag('+'):
			_, _ = fmt.Fprintf(s, "%s", formatPlusV(p))
		default:
			_, _ = fmt.Fprintf(s, "%s", p.Error())
		}

	case 's':
		_, _ = fmt.Fprintf(s, "%s", p.Error())

	default:
		_, _ = fmt.Fprintf(s, "%%!%c(%T=%s)", verb, p, fmt.Sprintf("%s", p.Error()))
	}
}
