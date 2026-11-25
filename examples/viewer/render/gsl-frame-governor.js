// The frame-rate governor: while the view moves, a viewer that cannot draw
// FLOOR_FPS frames a second draws its moving frames smaller until it can.
//
// A moving frame is drawn into the draft target (gsl-render-tier.js), whose
// size is a share of the shape target's. The governor owns that share, the
// scale: 1 draws the moving frame as large as the still one (Moving: Full),
// 0.5 is Draft, and below that is what the governor adds. The viewer reads
// scale to size the draft target, and shows what was asked for (the Resolution
// buttons, the Moving tier) next to what is happening, as the Draw on buttons
// do for the device that drew the frame.
//
// It is fed one call per frame drawn while the view moves (frame) and one when
// the view comes to rest (rest). The rate is the time between frames drawn
// back to back, so a lone frame (the still one that follows a drag) and a
// long gap (the user paused) are not rates. A step down needs SLOW_RUN frames
// in a row below the floor, so one hitch (a shader compile) does not trigger it.
//
// Stepping back up is a guess: the cost of a frame follows its pixel count, so
// from the median of the last FAST_RUN frames the time at the next larger
// scale is predicted, and the step is taken when that would still clear the
// floor by RECOVER_MARGIN. A step down soon after a step up shows the guess
// was wrong (the frame was not paying by the pixel), and the wait before the
// next try doubles.

export const FLOOR_FPS = 5;

// The scales a moving frame can be drawn at, as a share of the still frame's
// size, from largest to smallest.
export const MOVE_SCALES = [1, 0.5, 0.25, 0.125];

const SLOW_RUN = 4;
const FAST_RUN = 8;
const GAP_MS = 1000;
const RECOVER_MARGIN = 1.5;
const HOLD_MS = 3000;
const MAX_HOLD_MS = 30000;
const FLAP_MS = 10000;

function median(values) {
  const s = [...values].sort((a, b) => a - b);
  return s[Math.floor(s.length / 2)];
}

export class FrameGovernor {
  // onChange(scale) runs when the scale changes. baseline is the scale the
  // user asked for with Moving (0.5 for Draft, 1 for Full); the governor never
  // goes above it.
  constructor(onChange, baseline = 1, floorFps = FLOOR_FPS) {
    this.onChange = onChange;
    this.floorFps = floorFps;
    this.slowMs = 1000 / floorFps;
    this.baseline = baseline;
    this.scale = baseline;
    this.holdMs = HOLD_MS;
    this.holdUntil = 0;
    this.recoveredAt = -Infinity;
    this.rest();
  }

  // setBaseline records a new Moving choice. The scale follows it when it was
  // at the old one, and is cut down to it when it would be above it.
  setBaseline(baseline) {
    const was = this.baseline;
    this.baseline = baseline;
    const next = this.scale === was ? baseline : Math.min(this.scale, baseline);
    this.holdMs = HOLD_MS;
    this.set(next);
  }

  // rest is called when the view comes to rest: the next frame starts a new run.
  rest() {
    this.last = null;
    this.gaps = [];
  }

  // frame is called with the time, in ms, of each frame drawn while the view
  // moves.
  frame(now) {
    const prev = this.last;
    this.last = now;
    if (prev === null || now - prev > GAP_MS) {
      this.gaps = [];
      return;
    }
    this.gaps.push(now - prev);
    if (this.gaps.length > FAST_RUN) this.gaps.shift();

    const recent = this.gaps.slice(-SLOW_RUN);
    if (recent.length === SLOW_RUN && recent.every((g) => g > this.slowMs)) {
      this.stepDown(now);
    } else if (this.scale < this.baseline && this.gaps.length === FAST_RUN && now >= this.holdUntil) {
      const up = MOVE_SCALES[MOVE_SCALES.indexOf(this.scale) - 1];
      const predicted = median(this.gaps) * (up / this.scale) ** 2;
      if (predicted <= this.slowMs / RECOVER_MARGIN) this.stepUp(now, up);
    }
  }

  stepDown(now) {
    const i = MOVE_SCALES.indexOf(this.scale);
    this.rest();
    if (i < 0 || i === MOVE_SCALES.length - 1) return;
    if (now - this.recoveredAt < FLAP_MS) this.holdMs = Math.min(this.holdMs * 2, MAX_HOLD_MS);
    this.holdUntil = now + this.holdMs;
    this.set(MOVE_SCALES[i + 1]);
  }

  stepUp(now, up) {
    this.rest();
    this.recoveredAt = now;
    this.set(up);
  }

  set(scale) {
    if (scale === this.scale) return;
    this.scale = scale;
    this.onChange?.(scale);
  }
}
