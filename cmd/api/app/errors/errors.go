package errors

import (
	errmsg "github.com/karmada-io/dashboard/cmd/api/app/msgkey"
	"net/http"
)

type HttpError struct {
	Code int
	Msg  string
}

func (e *HttpError) Error() string {
	return e.Msg
}

func NewHttpError(code int, msg string) *HttpError {
	return &HttpError{
		Code: code,
		Msg:  msg,
	}
}

var (
	RequestValueInvalid                        = NewHttpError(http.StatusBadRequest, errmsg.RequestValueInvalid)
	InvalidYamlFormat                          = NewHttpError(http.StatusBadRequest, errmsg.InvalidYamlFormat)
	InvalidKeyValueFormat                      = NewHttpError(http.StatusBadRequest, errmsg.InvalidKeyValueFormat)
	ResourceMismatch                           = NewHttpError(http.StatusBadRequest, errmsg.ResourceMismatch)
	PolicyMissingTargetClusters                = NewHttpError(http.StatusBadRequest, errmsg.PolicyMissingTargetClusters)
	InvalidSortProperty                        = NewHttpError(http.StatusBadRequest, errmsg.InvalidSortProperty)
	InvalidSortOrder                           = NewHttpError(http.StatusBadRequest, errmsg.InvalidSortOrder)
	UnsupportedResourceKind                    = NewHttpError(http.StatusBadRequest, errmsg.UnsupportedResourceKind)
	InvalidPolicyReplicaWeight                 = NewHttpError(http.StatusBadRequest, errmsg.InvalidPolicyReplicaWeight)
	UnsupportedPolicyReplicaSchedulingType     = NewHttpError(http.StatusBadRequest, errmsg.UnsupportedPolicyReplicaSchedulingType)
	UnsupportedPolicyReplicaDivisionPreference = NewHttpError(http.StatusBadRequest, errmsg.UnsupportedPolicyReplicaDivisionPreference)
	InvalidStaticWeightClusters                = NewHttpError(http.StatusBadRequest, errmsg.InvalidStaticWeightClusters)
	EmptyStaticWeightClusters                  = NewHttpError(http.StatusBadRequest, errmsg.EmptyStaticWeightClusters)
	ResourceNamespaceRequired                  = NewHttpError(http.StatusBadRequest, errmsg.ResourceNamespaceRequired)
	NotAllowedNamespace                        = NewHttpError(http.StatusBadRequest, errmsg.NotAllowedNamespace)
	ClusterNotFound                            = NewHttpError(http.StatusNotFound, errmsg.ClusterNotFound)
	ClusterNotFoundInKarmada                   = NewHttpError(http.StatusNotFound, errmsg.ClusterNotFoundInKarmada)
	ResourceNotFound                           = NewHttpError(http.StatusNotFound, errmsg.ResourceNotFound)
	NamespaceNotFound                          = NewHttpError(http.StatusNotFound, errmsg.NamespaceNotFound)
	PolicyContainsUnauthorizedClusters         = NewHttpError(http.StatusForbidden, errmsg.PolicyContainsUnauthorizedClusters)
	ClusterAlreadyRegistered                   = NewHttpError(http.StatusConflict, errmsg.ClusterAlreadyRegistered)
	ClusterAlreadyRegisteredInKarmada          = NewHttpError(http.StatusConflict, errmsg.ClusterAlreadyRegisteredInKarmada)
	ResourceAlreadyExists                      = NewHttpError(http.StatusConflict, errmsg.ResourceAlreadyExists)
	ResourceUnprocessableEntity                = NewHttpError(http.StatusUnprocessableEntity, errmsg.ResourceUnprocessableEntity)
	FailedToReadClusterInfo                    = NewHttpError(http.StatusInternalServerError, errmsg.FailedToReadClusterInfo)
	FailedRequest                              = NewHttpError(http.StatusInternalServerError, errmsg.RequestFailed)
	ResourceOperationFailed                    = NewHttpError(http.StatusInternalServerError, errmsg.ResourceOperationFailed)
	ClusterRegistrationFailed                  = NewHttpError(http.StatusInternalServerError, errmsg.ClusterRegistrationFailed)
	ClusterVerificationFailed                  = NewHttpError(http.StatusInternalServerError, errmsg.ClusterVerificationFailed)
	ClusterMappingSaveFailed                   = NewHttpError(http.StatusInternalServerError, errmsg.ClusterMappingSaveFailed)
	ClusterLoadConfigFailed                    = NewHttpError(http.StatusInternalServerError, errmsg.ClusterLoadConfigFailed)

	//	ResourceUpdateFailed                       = NewHttpError(http.StatusInternalServerError, errmsg.ResourceUpdateFailed)
	//	ResourceDeleteFailed                       = NewHttpError(http.StatusInternalServerError, errmsg.ResourceDeleteFailed)
	//	ResourceCreateFailed                       = NewHttpError(http.StatusInternalServerError, errmsg.ResourceCreateFailed)
)
