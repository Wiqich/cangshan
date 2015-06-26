package application

var (
	moduleCreaters = make(map[string]func() interface{})
)

type Initializable interface {
	Initialize() error
}

type Runable interface {
	Run() error
}

func RegisterModuleCreater(name string, creater func() interface{}) {
	moduleCreaters[name] = creater
}
