package mock

import "github.com/supergiant/guber"

func (f *FakeGuber) OnNamespaceCreate(clbk func(*guber.Namespace) error) *FakeGuber {
	return f.mockNamespaces(&FakeGuberNamespaces{
		CreateFn: func(namespace *guber.Namespace) (*guber.Namespace, error) {
			if err := clbk(namespace); err != nil {
				return nil, err
			}
			return namespace, nil
		},
	})
}

func (f *FakeGuber) OnNamespaceDelete(clbk func(string) error) *FakeGuber {
	return f.mockNamespaces(&FakeGuberNamespaces{
		DeleteFn: func(name string) error {
			return clbk(name)
		},
	})
}

func (f *FakeGuber) mockNamespaces(namespaces *FakeGuberNamespaces) *FakeGuber {
	f.NamespacesFn = func() guber.NamespaceCollection {
		return namespaces
	}
	return f
}

type FakeGuber struct {
	NamespacesFn             func() guber.NamespaceCollection
	EventsFn                 func(namespace string) guber.EventCollection
	SecretsFn                func(namespace string) guber.SecretCollection
	ServicesFn               func(namespace string) guber.ServiceCollection
	ReplicationControllersFn func(namespace string) guber.ReplicationControllerCollection
	PodsFn                   func(namespace string) guber.PodCollection
	NodesFn                  func() guber.NodeCollection
}

func (f *FakeGuber) Namespaces() guber.NamespaceCollection {
	return f.NamespacesFn()
}

func (f *FakeGuber) Events(namespace string) guber.EventCollection {
	return f.EventsFn(namespace)
}

func (f *FakeGuber) Secrets(namespace string) guber.SecretCollection {
	return f.SecretsFn(namespace)
}

func (f *FakeGuber) Services(namespace string) guber.ServiceCollection {
	return f.ServicesFn(namespace)
}

func (f *FakeGuber) ReplicationControllers(namespace string) guber.ReplicationControllerCollection {
	return f.ReplicationControllersFn(namespace)
}

func (f *FakeGuber) Pods(namespace string) guber.PodCollection {
	return f.PodsFn(namespace)
}

func (f *FakeGuber) Nodes() guber.NodeCollection {
	return f.NodesFn()
}

type FakeGuberNamespaces struct {
	MetaFn   func() *guber.CollectionMeta
	NewFn    func() *guber.Namespace
	CreateFn func(e *guber.Namespace) (*guber.Namespace, error)
	QueryFn  func(q *guber.QueryParams) (*guber.NamespaceList, error)
	ListFn   func() (*guber.NamespaceList, error)
	GetFn    func(name string) (*guber.Namespace, error)
	UpdateFn func(name string, r *guber.Namespace) (*guber.Namespace, error)
	DeleteFn func(name string) error
}

func (f *FakeGuberNamespaces) Meta() *guber.CollectionMeta {
	return f.MetaFn()
}

func (f *FakeGuberNamespaces) New() *guber.Namespace {
	return f.NewFn()
}

func (f *FakeGuberNamespaces) Create(e *guber.Namespace) (*guber.Namespace, error) {
	return f.CreateFn(e)
}

func (f *FakeGuberNamespaces) Query(q *guber.QueryParams) (*guber.NamespaceList, error) {
	return f.QueryFn(q)
}

func (f *FakeGuberNamespaces) List() (*guber.NamespaceList, error) {
	return f.ListFn()
}

func (f *FakeGuberNamespaces) Get(name string) (*guber.Namespace, error) {
	return f.GetFn(name)
}

func (f *FakeGuberNamespaces) Update(name string, r *guber.Namespace) (*guber.Namespace, error) {
	return f.UpdateFn(name, r)
}

func (f *FakeGuberNamespaces) Delete(name string) error {
	return f.DeleteFn(name)
}
