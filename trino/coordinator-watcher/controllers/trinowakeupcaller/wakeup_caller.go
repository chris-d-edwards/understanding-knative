package trinowakeupcaller

type iTrinoWakeup interface {
	Trigger(sink string) error
}
