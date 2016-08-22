package core

import (
	"reflect"
	"sync"

	"github.com/imdario/mergo"
	"github.com/supergiant/supergiant/pkg/models"
)

type Collection struct {
	core *Core
}

func (c *Collection) Create(m models.Model) error {
	return c.core.DB.Create(m)
}

func (c *Collection) Get(id *int64, m models.Model) error {
	return c.core.DB.First(m, *id)
}

func (c *Collection) GetWithIncludes(id *int64, m models.Model, includes []string) error {
	scope := c.core.DB
	for _, include := range includes {
		scope = scope.Preload(include)
	}
	return scope.First(m, *id)
}

func (c *Collection) Update(id *int64, oldM models.Model, m models.Model) error {
	// models.ZeroReadonlyFields(m)

	// oldM := reflect.New(reflect.TypeOf(m)).Elem().Elem().Interface()

	if err := c.core.DB.First(oldM, *id); err != nil {
		return err
	}

	// Merge old item attributes into the empty fields of the newItem
	if err := mergo.Merge(m, oldM); err != nil {
		return err
	}

	return c.core.DB.Save(m)
}

func (c *Collection) Delete(id *int64, m models.Model) error { // Loaded so we can render out
	if err := c.core.DB.First(m, *id); err != nil {
		return err
	}
	return c.core.DB.Delete(m)
}

////////////////////////////////////////////////////////////////////////////////
// Private methods                                                            //
////////////////////////////////////////////////////////////////////////////////

func (c *Collection) inParallel(models interface{}, fn func(interface{}) error) (err error) {
	mv := reflect.ValueOf(models)
	count := mv.Len()

	var wg sync.WaitGroup
	wg.Add(count)

	for i := 0; i < count; i++ {
		go func(idx int) {
			defer wg.Done()
			err = fn(mv.Index(idx).Interface())
		}(i)
	}

	wg.Wait()
	return
}
