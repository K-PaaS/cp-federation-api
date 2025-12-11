package resources

import (
	"github.com/karmada-io/dashboard/pkg/dataselect"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type ResourceCell struct {
	unstructured.Unstructured
}

// GetProperty returns the given property of the Resource
func (c ResourceCell) GetProperty(name dataselect.PropertyName) dataselect.ComparableValue {
	switch name {
	case dataselect.NameProperty:
		return dataselect.StdComparableString(c.GetName())
	case dataselect.CreationTimestampProperty:
		return dataselect.StdComparableTime(c.GetCreationTimestamp().Time)
	case dataselect.NamespaceProperty:
		return dataselect.StdComparableString(c.GetNamespace())
	default:
		// if name is not supported then just return a constant dummy value, sort will have no effect.
		return nil
	}
}

func toCells(std []unstructured.Unstructured) []dataselect.DataCell {
	cells := make([]dataselect.DataCell, len(std))
	for i := range std {
		cells[i] = ResourceCell{Unstructured: std[i]}
	}
	return cells
}

func fromCells(cells []dataselect.DataCell) []unstructured.Unstructured {
	std := make([]unstructured.Unstructured, len(cells))
	for i := range std {
		std[i] = cells[i].(ResourceCell).Unstructured
	}
	return std
}
