package entity

type IService interface {
	IEntity
}

// RegisterSerivces registers the user custom service type
func RegisterService(serviceName string, servicePtr IService) {
	RegisterEntity(serviceName, servicePtr, true)
}
