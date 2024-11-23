package voice

type Server struct {
	channels map[string]*Channel
}

func NewServer() *Server {
	return &Server{
		channels: make(map[string]*Channel),
	}
}

func (s *Server) channel(name string) *Channel {
	if _, ok := s.channels[name]; !ok {
		s.channels[name] = NewChannel()
	}

	return s.channels[name]
}
