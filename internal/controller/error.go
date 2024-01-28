package controller

import v1 "github.com/isac322/cloudflared-operator/api/v1"

type Reasons interface {
	v1.TunnelConditionReason | v1.TunnelIngressConditionReason
}

type ErrorWithReason[T Reasons] struct {
	Reason T
	cause  error
}

func WrapError[T Reasons](cause error, reason T) ErrorWithReason[T] {
	return ErrorWithReason[T]{Reason: reason, cause: cause}
}

func (r ErrorWithReason[T]) Error() string {
	return r.cause.Error()
}

func (r ErrorWithReason[T]) Cause() error {
	return r.cause
}
