package controller

import v1 "github.com/isac322/cloudflared-operator/api/v1"

type Reasons interface {
	v1.TunnelConditionReason | v1.TunnelIngressConditionReason
}

type ReasonedError[T Reasons] struct {
	Reason T
	cause  error
}

func WrapError[T Reasons](cause error, reason T) ReasonedError[T] {
	return ReasonedError[T]{Reason: reason, cause: cause}
}

func (r ReasonedError[T]) Error() string {
	return r.cause.Error()
}

func (r ReasonedError[T]) Cause() error {
	return r.cause
}

func (r ReasonedError[T]) Unwrap() error {
	return r.cause
}
