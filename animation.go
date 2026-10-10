package senjaui

import "math"

// AnimType specifies the transition animation style for opening or closing a banner.
type AnimType int

const (
	// AnimNone disables transitions; open/close states apply instantaneously.
	AnimNone AnimType = iota
	// AnimScaleIn scales the banner from 0.0 to 1.0 around its pivot.
	AnimScaleIn
	// AnimFadeIn fades the banner opacity from 0.0 to 1.0.
	AnimFadeIn
	// AnimScaleAndFade combines scaling and alpha fading for a modern dialogue popup feel.
	AnimScaleAndFade
	// AnimSlideInTop slides the banner into position from above.
	AnimSlideInTop
	// AnimSlideInBottom slides the banner into position from below.
	AnimSlideInBottom
	// AnimSlideInLeft slides the banner into position from the left.
	AnimSlideInLeft
	// AnimSlideInRight slides the banner into position from the right.
	AnimSlideInRight
	// AnimCustom delegates interpolation step calculations to a custom AnimationStepFunc.
	AnimCustom
)

// EaseFunc takes normalized time t in [0.0, 1.0] and returns an eased progress value.
type EaseFunc func(t float64) float64

// EaseLinear performs linear interpolation without acceleration.
func EaseLinear(t float64) float64 {
	return t
}

// EaseOutQuad applies a quadratic deceleration curve (fast start, gentle stop).
func EaseOutQuad(t float64) float64 {
	return 1.0 - (1.0-t)*(1.0-t)
}

// EaseInOutQuad applies quadratic acceleration followed by deceleration.
func EaseInOutQuad(t float64) float64 {
	if t < 0.5 {
		return 2.0 * t * t
	}
	return 1.0 - math.Pow(-2.0*t+2.0, 2)/2.0
}

// EaseOutCubic applies cubic deceleration for punchier UI motion.
func EaseOutCubic(t float64) float64 {
	inv := 1.0 - t
	return 1.0 - inv*inv*inv
}

// EaseInOutCubic applies cubic acceleration followed by cubic deceleration.
func EaseInOutCubic(t float64) float64 {
	if t < 0.5 {
		return 4.0 * t * t * t
	}
	return 1.0 - math.Pow(-2.0*t+2.0, 3)/2.0
}

// EaseOutBack creates a slight overshoot past the target size before settling back.
func EaseOutBack(t float64) float64 {
	const c1 = 1.70158
	const c3 = c1 + 1.0
	inv := t - 1.0
	return 1.0 + c3*inv*inv*inv + c1*inv*inv
}

// EaseOutBounce produces a playful bounce effect at the end of the transition.
func EaseOutBounce(t float64) float64 {
	const n1 = 7.5625
	const d1 = 2.75

	if t < 1.0/d1 {
		return n1 * t * t
	} else if t < 2.0/d1 {
		t -= 1.5 / d1
		return n1*t*t + 0.75
	} else if t < 2.5/d1 {
		t -= 2.25 / d1
		return n1*t*t + 0.9375
	} else {
		t -= 2.625 / d1
		return n1*t*t + 0.984375
	}
}

// AnimationStepFunc is a callback invoked during each animation frame with progress in [0.0, 1.0].
type AnimationStepFunc func(b *Banner, t float64)

