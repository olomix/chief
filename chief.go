package chief

type Worker interface {
	Start()
	Stop()
}

type Controller interface {
	Register(worker Worker)
	Close()
}



