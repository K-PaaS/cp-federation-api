package errors

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
)

func ResourceError(err error) error {
	klog.ErrorS(err, "Failed to process the resource")
	klog.Errorf("error type: %T", err)

	if statusErr, ok := err.(*apierrors.StatusError); ok {
		klog.Infof("Error Status Code: %d, Reason: %s, Message: %s",
			statusErr.ErrStatus.Code, statusErr.ErrStatus.Reason, statusErr.ErrStatus.Message)
	} else {
		klog.Infof("Non-StatusError type: %T", err)
	}

	switch {
	case apierrors.IsNotFound(err):
		return ResourceNotFound
	case apierrors.IsAlreadyExists(err), apierrors.IsConflict(err):
		return ResourceAlreadyExists
	case apierrors.IsInvalid(err):
		return ResourceUnprocessableEntity
	default:
		return ResourceOperationFailed
	}
}
