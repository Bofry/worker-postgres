package internal

type Router map[string]RouteComponent

func (r Router) Add(slot string, handler MessageHandler, handlerComponentID string) {
	r[slot] = RouteComponent{
		MessageHandler:     handler,
		HandlerComponentID: handlerComponentID,
	}
}

func (r Router) Remove(stream string) {
	delete(r, stream)
}

func (r Router) Get(stream string) MessageHandler {
	if r == nil {
		return nil
	}

	if v, ok := r[stream]; ok {
		return v.MessageHandler
	}
	return nil
}

func (r Router) Has(stream string) bool {
	if r == nil {
		return false
	}

	if _, ok := r[stream]; ok {
		return true
	}
	return false
}

func (r Router) FindHandlerComponentID(stream string) string {
	if r == nil {
		return ""
	}

	if v, ok := r[stream]; ok {
		return v.HandlerComponentID
	}
	return ""
}

func (r Router) GetRouteComponent(stream string) RouteComponent {
	if r == nil {
		return RouteComponent{}
	}

	return r[stream]
}
