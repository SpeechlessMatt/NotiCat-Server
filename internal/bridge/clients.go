package bridge

type Client string

// register Clients
const (
	ClientBili Client = "bili"
	ClientBUPT Client = "bupt"
)

var SupportedClients = map[Client]bool{
	ClientBili: true,
	ClientBUPT: true,
}
// end register

