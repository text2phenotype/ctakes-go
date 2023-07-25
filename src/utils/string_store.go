package utils

import (
	"strings"
	"sync"
)

var storeStoreInstance *stringStoreImpl
var stringStoreInitializer sync.Once

type StringStore interface {
	GetPointer(s string) *string
	GetPointers(ss []string) []*string

	// methods to avoid memory leaks. When all dictionaries are loaded the service locks string store.
	// when store is locked it doesn't save new pointers
	Lock()
	IsLocked() bool
}

type stringStoreImpl struct {
	StringStore
	store    sync.Map //map[string] *string
	isLocked bool
}

func (stringStore *stringStoreImpl) GetPointer(s string) *string {
	lowerS := strings.ToLower(s)

	if !stringStore.isLocked {
		ptr, _ := stringStore.store.LoadOrStore(lowerS, &lowerS)
		return ptr.(*string)
	}

	ptr, ok := stringStore.store.Load(lowerS)
	if !ok {
		return &lowerS
	}

	return ptr.(*string)

}

func (stringStore *stringStoreImpl) GetPointers(ss []string) []*string {
	ptrs := make([]*string, len(ss))
	for i, s := range ss {
		ptrs[i] = stringStore.GetPointer(s)
	}
	return ptrs
}

func (stringStore *stringStoreImpl) Lock() {
	stringStore.isLocked = true
}

func (stringStore *stringStoreImpl) IsLocked() bool {
	return stringStore.isLocked
}

func GlobalStringStore() StringStore {
	stringStoreInitializer.Do(func() {
		storeStoreInstance = new(stringStoreImpl)
		storeStoreInstance.isLocked = false
	})

	return storeStoreInstance
}
