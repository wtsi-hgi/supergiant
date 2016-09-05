package core

type safeMapOpType int

const (
	safeMapList safeMapOpType = iota
	safeMapGet
	safeMapPut
	safeMapDelete
)

type safeMapOp struct {
	returnChannel chan interface{}
	opType        safeMapOpType

	key   string
	value interface{}

	// Optional, used for logging (for example, Action start/stop)
	desc string
}

// SafeMap is a concurrently-accessible map
type SafeMap struct {
	core *Core
	m    map[string]interface{}
	ch   chan *safeMapOp
}

func NewSafeMap(core *Core) *SafeMap {
	m := &SafeMap{
		core,
		make(map[string]interface{}),
		make(chan *safeMapOp),
	}

	go func() {
		for {
			op := <-m.ch
			switch op.opType {

			case safeMapPut:
				m.core.Log.Infof("PUT :: %s", op.desc)
				m.m[op.key] = op.value

			case safeMapDelete:
				m.core.Log.Infof("DEL :: %s", op.desc)
				delete(m.m, op.key)
			}

			var returnData interface{}
			if op.opType == safeMapList {
				var values []interface{}
				for _, value := range m.m {
					values = append(values, value)
				}
				returnData = values
			} else {
				returnData = m.m[op.key]
			}

			op.returnChannel <- returnData
		}
	}()

	return m
}

//------------------------------------------------------------------------------

func (m *SafeMap) List() []interface{} {
	return m.op(safeMapList, "", "", nil).([]interface{})
}

func (m *SafeMap) Get(key string) interface{} {
	return m.op(safeMapGet, "", key, nil)
}

func (m *SafeMap) Put(desc string, key string, value interface{}) {
	m.op(safeMapPut, desc, key, value)
}

func (m *SafeMap) Delete(desc string, key string) {
	m.op(safeMapDelete, desc, key, nil)
}

////////////////////////////////////////////////////////////////////////////////
// Private                                                                    //
////////////////////////////////////////////////////////////////////////////////

func (m *SafeMap) op(opType safeMapOpType, desc string, key string, value interface{}) interface{} {
	ch := make(chan interface{})
	m.ch <- &safeMapOp{ch, opType, key, value, desc}
	return <-ch
}
