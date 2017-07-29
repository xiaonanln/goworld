package entity

//type CustomAttr interface {
//	CustomAttr() *customAttr
//}
//
//type customAttr struct {
//	owner  *Entity
//	parent interface{}
//	pkey   interface{} // key of this item in parent
//}

//func (ca *customAttr) getOwner() *Entity {
//	return ca.owner
//}
//
//func (ca *customAttr) sendAttrChangeToClients(key string, val interface{}) {
//	owner := ca.getOwner()
//	if owner != nil {
//		// send the change to owner's client
//		owner.sendAttrChangeToClients(ca, key, val)
//	}
//}
//
//func (ca *customAttr) sendAttrDelToClients(key string) {
//	owner := ca.getOwner()
//	if owner != nil {
//		owner.sendAttrDelToClients(ca, key)
//	}
//}
//
//func (ca *customAttr) getPathFromOwner() []string {
//	path := make([]string, 0, 4) // preallocate some Space
//	for {
//		if ca.parent != nil {
//			path = append(path, ca.pkey)
//			ca = ca.parent
//		} else { // la.parent  == nil, must be the root attr
//			if ca != ca.owner.Attrs {
//				gwlog.Panicf("Root attrs is not found")
//			}
//			break
//		}
//	}
//	return path
//}
