package caches

import (
	"github.com/cheekybits/genny/generic"
)

type KeyType generic.Type
type ValueType generic.Type

type KeyTypeValueTypeCache interface{
	Put(KeyType, ValueType) ValueType
	Get(KeyType) (ValueType, error)
}
