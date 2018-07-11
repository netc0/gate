package backend

type Backend struct {
	name    string // 节点名称
	address string // 节点地址

	routes []uint32
}

