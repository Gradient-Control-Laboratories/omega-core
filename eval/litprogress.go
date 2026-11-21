package eval

// litShares are where each phase of a lit render ends, as a share of its work
// (LitOptions.Progress), in the order renderLit runs them. The march is nearly
// all of the time: on a scene of 24 threaded fasteners it took 97% of a
// 340x215 render (9.2 s) and 98% of a 680x430 one (33.7 s), the share of rays
// it had settled moving close to linearly with the time, and the normals, the
// occlusion taps, the appearance and the glass shared the rest. A scene with
// fewer rays still marching, or with glass, spends differently, which is why
// the progress is an estimate.
var litShares = struct {
	march, normals, occlusion, appearance, glass float64
}{
	march:      0.97,
	normals:    0.98,
	occlusion:  0.99,
	appearance: 0.995,
	glass:      0.997,
}

// litProgress reports a render's progress to LitOptions.Progress as the share of
// its work done. It never reports less than it has, so the share only goes up.
// A nil *litProgress, or one without a callback, reports nothing.
type litProgress struct {
	report func(share float64)
	last   float64
}

func newLitProgress(report func(share float64)) *litProgress {
	if report == nil {
		return nil
	}
	return &litProgress{report: report}
}

// at reports that the share f of the work is done.
func (p *litProgress) at(f float64) {
	if p == nil {
		return
	}
	if f > 1 {
		f = 1
	}
	if f <= p.last {
		return
	}
	p.last = f
	p.report(f)
}

// span is a reporter for a phase that runs from the share lo to the share hi:
// given the share g of the phase done, it reports lo + (hi-lo)·g. Nil without
// a callback.
func (p *litProgress) span(lo, hi float64) func(g float64) {
	if p == nil {
		return nil
	}
	return func(g float64) {
		if g < 0 {
			g = 0
		}
		if g > 1 {
			g = 1
		}
		p.at(lo + (hi-lo)*g)
	}
}
