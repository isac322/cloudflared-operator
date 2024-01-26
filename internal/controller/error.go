package controller

import v1 "github.com/isac322/cloudflared-operator/api/v1"

type ErrorWithReason struct {
	Reason v1.TunnelConditionReason
	cause  error
}

func WrapError(cause error, reason v1.TunnelConditionReason) ErrorWithReason {
	return ErrorWithReason{Reason: reason, cause: cause}
}

func (r ErrorWithReason) Error() string {
	return r.cause.Error()
}

func (r ErrorWithReason) Cause() error {
	return r.cause
}
